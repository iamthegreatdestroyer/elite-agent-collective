import io
import json
import unittest
import unittest.mock

from mcp_server import server as srv
from mcp_server.registry import AgentRegistry
from mcp_server.tests.fixtures import make_agents_dir


class ServerTestCase(unittest.TestCase):
    def setUp(self):
        self.tmp = make_agents_dir()
        self.addCleanup(self.tmp.cleanup)
        self.server = srv.McpServer(AgentRegistry(self.tmp.name).load())
        self._id = 0

    def rpc(self, method, params=None):
        self._id += 1
        message = {"jsonrpc": "2.0", "id": self._id, "method": method}
        if params is not None:
            message["params"] = params
        reply = self.server.handle(message)
        self.assertEqual(reply["jsonrpc"], "2.0")
        self.assertEqual(reply["id"], self._id)
        return reply

    def init(self):
        return self.rpc("initialize", {"protocolVersion": "2025-06-18", "capabilities": {}})

    def call(self, name, arguments=None):
        return self.rpc("tools/call", {"name": name, "arguments": arguments or {}})


class LifecycleTests(ServerTestCase):
    def test_initialize(self):
        result = self.init()["result"]
        self.assertEqual(result["protocolVersion"], "2025-06-18")
        self.assertEqual(result["serverInfo"]["name"], "elite-agent-collective")
        self.assertEqual(set(result["capabilities"]), {"tools", "resources"})

    def test_initialize_echoes_older_supported_version(self):
        reply = self.rpc("initialize", {"protocolVersion": "2024-11-05"})
        self.assertEqual(reply["result"]["protocolVersion"], "2024-11-05")

    def test_initialize_unknown_version_answers_newest(self):
        reply = self.rpc("initialize", {"protocolVersion": "1999-01-01"})
        self.assertEqual(reply["result"]["protocolVersion"], srv.PROTOCOL_VERSIONS[0])

    def test_requests_before_initialize_are_refused(self):
        reply = self.rpc("tools/list")
        self.assertEqual(reply["error"]["code"], srv.INVALID_REQUEST)

    def test_ping_allowed_before_initialize(self):
        self.assertEqual(self.rpc("ping")["result"], {})

    def test_notifications_get_no_reply(self):
        self.assertIsNone(self.server.handle({"jsonrpc": "2.0", "method": "notifications/initialized"}))
        self.assertIsNone(self.server.handle({"jsonrpc": "2.0", "method": "no/such/notification"}))

    def test_stray_response_is_ignored(self):
        self.assertIsNone(self.server.handle({"jsonrpc": "2.0", "id": 9, "result": {}}))


class ProtocolErrorTests(ServerTestCase):
    def test_unknown_method(self):
        self.init()
        self.assertEqual(self.rpc("prompts/list")["error"]["code"], srv.METHOD_NOT_FOUND)

    def test_not_jsonrpc(self):
        reply = self.server.handle({"id": 1, "method": "ping"})
        self.assertEqual(reply["error"]["code"], srv.INVALID_REQUEST)
        self.assertEqual(self.server.handle("text")["error"]["code"], srv.INVALID_REQUEST)

    def test_params_must_be_object(self):
        reply = self.server.handle({"jsonrpc": "2.0", "id": 1, "method": "ping", "params": [1]})
        self.assertEqual(reply["error"]["code"], srv.INVALID_PARAMS)

    def test_parse_error(self):
        reply = json.loads(self.server.handle_line(b"{not json\n"))
        self.assertEqual(reply["error"]["code"], srv.PARSE_ERROR)
        self.assertIsNone(reply["id"])

    def test_deeply_nested_json_is_a_parse_error_not_a_crash(self):
        reply = json.loads(self.server.handle_line(b"[" * 200000 + b"\n"))
        self.assertEqual(reply["error"]["code"], srv.PARSE_ERROR)
        self.assertEqual(self.rpc("ping")["result"], {})

    def test_null_params_means_no_params(self):
        reply = self.server.handle({"jsonrpc": "2.0", "id": 1, "method": "ping", "params": None})
        self.assertEqual(reply["result"], {})

    def test_id_must_be_string_or_number(self):
        for bad in ({"a": 1}, [1], True):
            reply = self.server.handle({"jsonrpc": "2.0", "id": bad, "method": "ping"})
            self.assertEqual(reply["error"]["code"], srv.INVALID_REQUEST)
            self.assertIsNone(reply["id"])
        self.assertEqual(self.server.handle({"jsonrpc": "2.0", "id": "x", "method": "ping"})["id"], "x")

    def test_non_finite_numbers_are_parse_errors_and_never_echoed(self):
        for line in (
            b'{"jsonrpc":"2.0","id":NaN,"method":"ping"}',
            b'{"jsonrpc":"2.0","id":-Infinity,"method":"ping"}',
            b'{"jsonrpc":"2.0","id":1e999,"method":"ping"}',
            b'{"jsonrpc":"2.0","id":1,"method":"ping","params":{"x":1e999}}',
        ):
            raw = self.server.handle_line(line + b"\n")
            self.assertNotIn(b"NaN", raw)
            self.assertNotIn(b"Infinity", raw)
            reply = json.loads(raw, parse_constant=self.fail)
            self.assertEqual((reply["id"], reply["error"]["code"]), (None, srv.PARSE_ERROR), line)
        # a non-finite id that reaches handle() directly is refused too, with a null id
        reply = self.server.handle({"jsonrpc": "2.0", "id": float("inf"), "method": "ping"})
        self.assertEqual((reply["id"], reply["error"]["code"]), (None, srv.INVALID_REQUEST))
        self.assertEqual(self.server.handle({"jsonrpc": "2.0", "id": 1.5, "method": "ping"})["id"], 1.5)

    def test_bad_id_is_never_echoed_even_without_a_method(self):
        for message in (
            {"jsonrpc": "2.0", "id": {"a": 1}},
            {"jsonrpc": "2.0", "id": [1]},
            {"jsonrpc": "2.0", "id": True, "method": 5},
        ):
            reply = self.server.handle(message)
            self.assertEqual((reply["id"], reply["error"]["code"]), (None, srv.INVALID_REQUEST), message)
        # a valid id is still echoed on a method-less message
        self.assertEqual(self.server.handle({"jsonrpc": "2.0", "id": 7})["id"], 7)

    def test_blank_line_ignored(self):
        self.assertIsNone(self.server.handle_line(b"  \r\n"))

    def test_batch(self):
        line = json.dumps(
            [
                {"jsonrpc": "2.0", "id": "a", "method": "ping"},
                {"jsonrpc": "2.0", "method": "notifications/initialized"},
                {"jsonrpc": "2.0", "id": "b", "method": "nope"},
            ]
        ).encode()
        replies = json.loads(self.server.handle_line(line))
        self.assertEqual([r["id"] for r in replies], ["a", "b"])
        self.assertEqual(json.loads(self.server.handle_line(b"[]"))["error"]["code"], srv.INVALID_REQUEST)

    def test_internal_error_does_not_leak_or_crash(self):
        self.init()
        self.server.registry = None  # force a bug inside a handler
        with unittest.mock.patch("sys.stderr", io.StringIO()):
            reply = self.call("list_agents")
        self.assertEqual(reply["error"], {"code": srv.INTERNAL_ERROR, "message": "Internal error"})


class ToolTests(ServerTestCase):
    def setUp(self):
        super().setUp()
        self.init()

    def test_tools_list(self):
        tools = self.rpc("tools/list")["result"]["tools"]
        self.assertEqual(
            [t["name"] for t in tools], ["list_agents", "find_agent", "get_agent", "invoke_agent"]
        )
        for tool in tools:
            self.assertEqual(tool["inputSchema"]["type"], "object")
            self.assertTrue(tool["description"])

    def test_list_agents(self):
        result = self.call("list_agents")["result"]
        self.assertFalse(result["isError"])
        self.assertEqual(result["structuredContent"]["count"], 3)
        self.assertIn("| @ALPHA |", result["content"][0]["text"])

    def test_list_agents_filters(self):
        result = self.call("list_agents", {"tier": 2})["result"]
        self.assertEqual([a["name"] for a in result["structuredContent"]["agents"]], ["BETA"])
        result = self.call("list_agents", {"category": "FOUND"})["result"]
        self.assertEqual(result["structuredContent"]["count"], 1)
        # JSON has one number type: 2.0 is an acceptable integer
        self.assertEqual(self.call("list_agents", {"tier": 2.0})["result"]["structuredContent"]["count"], 1)

    def test_find_agent(self):
        result = self.call("find_agent", {"task_description": "kubernetes pipelines", "limit": 1})["result"]
        self.assertEqual([m["name"] for m in result["structuredContent"]["matches"]], ["GAMMA"])
        self.assertIn("@GAMMA", result["content"][0]["text"])

    def test_find_agent_does_not_echo_a_huge_task(self):
        result = self.call("find_agent", {"task_description": "kubernetes " * 5000})["result"]
        self.assertLess(len(result["content"][0]["text"]), 2000)

    def test_find_agent_no_match_is_said_plainly(self):
        result = self.call("find_agent", {"task_description": "zebra xylophone"})["result"]
        self.assertEqual(result["structuredContent"]["matches"], [])
        self.assertIn("No agent definition", result["content"][0]["text"])
        self.assertFalse(result["isError"])

    def test_find_agent_empty_task_is_tool_error(self):
        self.assertTrue(self.call("find_agent", {"task_description": "  "})["result"]["isError"])

    def test_get_agent(self):
        text = self.call("get_agent", {"name": "alpha"})["result"]["content"][0]["text"]
        self.assertTrue(text.startswith("---\nname: ALPHA"))

    def test_get_agent_unknown_is_tool_error_not_protocol_error(self):
        reply = self.call("get_agent", {"name": "NOPE"})
        self.assertNotIn("error", reply)
        self.assertTrue(reply["result"]["isError"])
        self.assertIn("Agent not found", reply["result"]["content"][0]["text"])

    def test_invoke_agent_returns_prompt_and_says_no_model_ran(self):
        result = self.call("invoke_agent", {"name": "ALPHA", "task": "review my cipher"})["result"]
        text = result["content"][0]["text"]
        self.assertIn("No model was called", text)
        self.assertIn("review my cipher", text)
        self.assertIn("# @ALPHA - Cryptography", text)
        self.assertNotIn("codename:", text)  # prompt only, not the front matter
        structured = result["structuredContent"]
        self.assertIs(structured["executed"], False)
        self.assertEqual(structured["agent"], "ALPHA")
        self.assertTrue(structured["prompt"].startswith("# @ALPHA"))

    def test_invoke_agent_without_task(self):
        result = self.call("invoke_agent", {"name": "BETA"})["result"]
        self.assertIsNone(result["structuredContent"]["task"])
        self.assertNotIn("Task:", result["content"][0]["text"])

    def test_bad_arguments_are_invalid_params(self):
        for name, arguments in (
            ("nope", {}),
            ("get_agent", {}),  # missing required
            ("get_agent", {"name": 5}),  # wrong type
            ("get_agent", {"name": "ALPHA", "extra": 1}),  # unknown key
            ("list_agents", {"tier": "one"}),
            ("list_agents", {"tier": True}),
            ("list_agents", {"tier": 1.5}),
            ("find_agent", {"task_description": "x", "limit": 0}),
            ("find_agent", {"task_description": "x", "limit": 11}),
            ("list_agents", ["tier"]),
        ):
            reply = self.rpc("tools/call", {"name": name, "arguments": arguments})
            self.assertEqual(reply.get("error", {}).get("code"), srv.INVALID_PARAMS, (name, arguments))

    def test_missing_arguments_key_means_no_arguments(self):
        self.assertEqual(self.rpc("tools/call", {"name": "list_agents"})["result"]["structuredContent"]["count"], 3)


class ResourceTests(ServerTestCase):
    def setUp(self):
        super().setUp()
        self.init()

    def test_resources_list(self):
        uris = [r["uri"] for r in self.rpc("resources/list")["result"]["resources"]]
        self.assertEqual(uris, ["agents://roster", "agents://ALPHA", "agents://BETA", "agents://GAMMA"])

    def test_read_roster_and_agent(self):
        roster = self.rpc("resources/read", {"uri": "agents://roster"})["result"]["contents"][0]
        self.assertIn("(3 agents)", roster["text"])
        agent = self.rpc("resources/read", {"uri": "agents://beta"})["result"]["contents"][0]
        self.assertEqual(agent["mimeType"], "text/markdown")
        self.assertIn("# @BETA", agent["text"])

    def test_read_unknown(self):
        for uri in ("agents://NOPE", "file:///etc/passwd"):
            reply = self.rpc("resources/read", {"uri": uri})
            self.assertEqual(reply["error"]["code"], srv.RESOURCE_NOT_FOUND)
        self.assertEqual(self.rpc("resources/read", {})["error"]["code"], srv.INVALID_PARAMS)


class WireTests(ServerTestCase):
    def test_serve_loop_one_line_per_message_ascii_lf(self):
        requests = [
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
            {"jsonrpc": "2.0", "method": "notifications/initialized"},
            {"jsonrpc": "2.0", "id": 2, "method": "tools/call",
             "params": {"name": "get_agent", "arguments": {"name": "BETA"}}},
        ]
        stdin = io.BytesIO(b"".join(json.dumps(r).encode() + b"\r\n" for r in requests))
        stdout = io.BytesIO()
        self.server.serve(stdin, stdout)
        raw = stdout.getvalue()
        self.assertNotIn(b"\r", raw)
        raw.decode("ascii")  # non-ASCII is escaped, never raw
        lines = raw.splitlines()
        self.assertEqual([json.loads(line)["id"] for line in lines], [1, 2])
        self.assertIn("Café", json.loads(lines[1])["result"]["content"][0]["text"])


if __name__ == "__main__":
    unittest.main()
