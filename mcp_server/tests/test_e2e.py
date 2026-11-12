"""End to end: start the real server as a child process and speak JSON-RPC to it over stdio."""

import json
import os
import subprocess
import sys
import threading
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
TIMEOUT_S = 20
# No console window flashes up on Windows when the suite runs in the background.
_NO_WINDOW = getattr(subprocess, "CREATE_NO_WINDOW", 0)


def spawn(env_overrides=None):
    env = dict(os.environ)
    env.pop("AGENTS_DIR", None)
    env.update(env_overrides or {})
    return subprocess.Popen(
        [sys.executable, "-m", "mcp_server"],
        cwd=REPO_ROOT,
        env=env,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        creationflags=_NO_WINDOW,
    )


class Client:
    def __init__(self, proc):
        self.proc = proc
        self._id = 0

    def send(self, message):
        self.proc.stdin.write(json.dumps(message).encode("utf-8") + b"\n")
        self.proc.stdin.flush()

    def read(self):
        box = []
        reader = threading.Thread(target=lambda: box.append(self.proc.stdout.readline()), daemon=True)
        reader.start()
        reader.join(TIMEOUT_S)
        if not box or not box[0]:
            raise AssertionError("no reply from server within %ss" % TIMEOUT_S)
        return json.loads(box[0].decode("utf-8"))

    def request(self, method, params=None):
        self._id += 1
        message = {"jsonrpc": "2.0", "id": self._id, "method": method}
        if params is not None:
            message["params"] = params
        self.send(message)
        reply = self.read()
        assert reply["id"] == self._id, reply
        return reply

    def call(self, name, arguments=None):
        return self.request("tools/call", {"name": name, "arguments": arguments or {}})["result"]


class EndToEndTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.proc = spawn()
        cls.client = Client(cls.proc)
        cls.init = cls.client.request(
            "initialize",
            {
                "protocolVersion": "2025-06-18",
                "capabilities": {},
                "clientInfo": {"name": "e2e-test", "version": "0"},
            },
        )
        cls.client.send({"jsonrpc": "2.0", "method": "notifications/initialized"})

    @classmethod
    def tearDownClass(cls):
        cls.proc.stdin.close()
        try:
            cls.exit_code = cls.proc.wait(TIMEOUT_S)
        finally:
            if cls.proc.poll() is None:
                cls.proc.kill()
            cls.proc.stdout.close()
            cls.proc.stderr.close()
        assert cls.exit_code == 0, "server did not exit cleanly when stdin closed"

    def test_initialize_handshake(self):
        result = self.init["result"]
        self.assertEqual(result["protocolVersion"], "2025-06-18")
        self.assertEqual(result["serverInfo"]["name"], "elite-agent-collective")
        self.assertIn("tools", result["capabilities"])

    def test_tools_list(self):
        tools = self.client.request("tools/list")["result"]["tools"]
        self.assertEqual(
            sorted(t["name"] for t in tools), ["find_agent", "get_agent", "invoke_agent", "list_agents"]
        )

    def test_list_agents_serves_the_forty_repo_definitions(self):
        result = self.client.call("list_agents")
        self.assertEqual(result["structuredContent"]["count"], 40)
        on_disk = {p.name[: -len(".agent.md")] for p in (REPO_ROOT / ".github" / "agents").glob("*.agent.md")}
        self.assertEqual({a["name"] for a in result["structuredContent"]["agents"]}, on_disk)

    def test_get_agent_matches_file_on_disk(self):
        text = self.client.call("get_agent", {"name": "apex"})["content"][0]["text"]
        on_disk = (REPO_ROOT / ".github" / "agents" / "APEX.agent.md").read_text(encoding="utf-8")
        self.assertEqual(text, on_disk.replace("\r\n", "\n"))

    def test_find_agent(self):
        result = self.client.call("find_agent", {"task_description": "audit my encryption and key exchange"})
        self.assertIn("CIPHER", [m["name"] for m in result["structuredContent"]["matches"]])

    def test_invoke_agent_returns_prompt_without_running_a_model(self):
        result = self.client.call("invoke_agent", {"name": "CIPHER", "task": "review this handshake"})
        self.assertIs(result["structuredContent"]["executed"], False)
        text = result["content"][0]["text"]
        self.assertIn("No model was called", text)
        self.assertIn("review this handshake", text)
        self.assertIn("@CIPHER", text)

    def test_unknown_agent_is_a_tool_error(self):
        self.assertTrue(self.client.call("get_agent", {"name": "NOPE"})["isError"])

    def test_resources(self):
        resources = self.client.request("resources/list")["result"]["resources"]
        self.assertEqual(len(resources), 41)  # roster + 40 agents
        contents = self.client.request("resources/read", {"uri": "agents://roster"})["result"]["contents"]
        self.assertIn("(40 agents)", contents[0]["text"])

    def test_protocol_errors_keep_the_session_alive(self):
        self.client.send({"jsonrpc": "2.0", "method": "notifications/cancelled", "params": {"requestId": 99}})
        self.assertEqual(self.client.request("no/such/method")["error"]["code"], -32601)
        self.proc.stdin.write(b"this is not json\n")
        self.proc.stdin.flush()
        self.assertEqual(self.client.read()["error"]["code"], -32700)
        self.assertEqual(self.client.request("ping")["result"], {})


class StartupTests(unittest.TestCase):
    def test_bad_agents_dir_exits_nonzero_with_stderr_message_and_silent_stdout(self):
        proc = spawn({"AGENTS_DIR": str(REPO_ROOT / "no-such-dir")})
        out, err = proc.communicate(timeout=TIMEOUT_S)
        self.assertEqual(proc.returncode, 1)
        self.assertEqual(out, b"")
        self.assertIn(b"cannot start", err)

    def test_agents_dir_override(self):
        proc = spawn({"AGENTS_DIR": str(REPO_ROOT / ".github" / "agents")})
        request = {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}
        out, err = proc.communicate(json.dumps(request).encode() + b"\n", timeout=TIMEOUT_S)
        self.assertEqual(proc.returncode, 0)
        self.assertEqual(json.loads(out)["id"], 1)
        self.assertIn(b"40 agents", err)


if __name__ == "__main__":
    unittest.main()
