"""stdio MCP server for the Elite Agent Collective.

Speaks JSON-RPC 2.0 over stdin/stdout, one JSON message per line (the MCP
stdio transport). Standard library only. It never calls a model and never
opens a network connection: ``invoke_agent`` returns the agent's prompt for
the caller's own model to run.

    python -m mcp_server            # serve .github/agents of this repo
    AGENTS_DIR=/some/dir python -m mcp_server
"""

from __future__ import annotations

import json
import math
import sys
from typing import Any, BinaryIO

from . import __version__
from .registry import AgentRegistry, RegistryError

SERVER_NAME = "elite-agent-collective"
# Newest first. initialize echoes the client's version when it is one of these,
# otherwise answers with the newest (the client then decides whether to go on).
PROTOCOL_VERSIONS = ("2025-06-18", "2025-03-26", "2024-11-05")

PARSE_ERROR = -32700
INVALID_REQUEST = -32600
METHOD_NOT_FOUND = -32601
INVALID_PARAMS = -32602
INTERNAL_ERROR = -32603
RESOURCE_NOT_FOUND = -32002

RESOURCE_SCHEME = "agents://"
ROSTER_URI = RESOURCE_SCHEME + "roster"
TASK_ECHO_CHARS = 200  # find_agent repeats at most this much of the task back

NO_MODEL_NOTE = (
    "No model was called. This server does not run agents: the text below is "
    "the agent's prompt. Use it as the system prompt (or leading instructions) "
    "for your own model and answer the task yourself."
)

TOOLS = [
    {
        "name": "list_agents",
        "description": "List the agents with tier, category and one-line role. "
        "Optional filters: tier (number) and category (substring, case-insensitive).",
        "inputSchema": {
            "type": "object",
            "properties": {
                "tier": {"type": "integer", "description": "Only this tier."},
                "category": {"type": "string", "description": "Only categories containing this text."},
            },
            "additionalProperties": False,
        },
    },
    {
        "name": "find_agent",
        "description": "Rank agents for a task by word overlap with each agent's own "
        "definition (lexical match, not semantic). Returns up to `limit` agents with "
        "the words that matched, or says that nothing matched.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "task_description": {"type": "string", "description": "What you need help with."},
                "limit": {"type": "integer", "minimum": 1, "maximum": 10, "description": "How many to return (default 3)."},
            },
            "required": ["task_description"],
            "additionalProperties": False,
        },
    },
    {
        "name": "get_agent",
        "description": "Return one agent's full definition file (front matter and prompt).",
        "inputSchema": {
            "type": "object",
            "properties": {"name": {"type": "string", "description": "Agent codename, e.g. APEX."}},
            "required": ["name"],
            "additionalProperties": False,
        },
    },
    {
        "name": "invoke_agent",
        "description": "Return an agent's prompt, framed around an optional task, for the "
        "CALLER to run. This server never calls a model, local or cloud; it produces no "
        "agent answer itself.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "name": {"type": "string", "description": "Agent codename, e.g. CIPHER."},
                "task": {"type": "string", "description": "The task to put to the agent (optional)."},
            },
            "required": ["name"],
            "additionalProperties": False,
        },
    },
]
_TOOL_SCHEMAS = {tool["name"]: tool["inputSchema"] for tool in TOOLS}


class RpcError(Exception):
    """A JSON-RPC protocol error (as opposed to a failed tool call)."""

    def __init__(self, code: int, message: str, data: Any = None):
        super().__init__(message)
        self.code, self.message, self.data = code, message, data


class ToolError(Exception):
    """A tool ran and failed; reported in the result with isError=true."""


def _text(text: str, structured: dict | None = None, is_error: bool = False) -> dict:
    result: dict[str, Any] = {"content": [{"type": "text", "text": text}], "isError": is_error}
    if structured is not None:
        result["structuredContent"] = structured
    return result


_JSON_TYPES = {"string": str, "integer": int, "object": dict}


def _check_arguments(tool: str, arguments: Any) -> dict:
    """Validate arguments against the tool's inputSchema (the subset we use)."""
    if arguments is None:
        arguments = {}
    if not isinstance(arguments, dict):
        raise RpcError(INVALID_PARAMS, f"{tool}: arguments must be an object")
    schema = _TOOL_SCHEMAS[tool]
    properties = schema["properties"]
    for key in schema.get("required", []):
        if key not in arguments:
            raise RpcError(INVALID_PARAMS, f"{tool}: missing required argument '{key}'")
    for key, value in arguments.items():
        if key not in properties:
            raise RpcError(INVALID_PARAMS, f"{tool}: unknown argument '{key}'")
        spec = properties[key]
        wanted = _JSON_TYPES[spec["type"]]
        if isinstance(value, float) and wanted is int and value.is_integer():
            value = arguments[key] = int(value)
        if not isinstance(value, wanted) or isinstance(value, bool):
            raise RpcError(INVALID_PARAMS, f"{tool}: '{key}' must be of type {spec['type']}")
        if "minimum" in spec and value < spec["minimum"] or "maximum" in spec and value > spec["maximum"]:
            raise RpcError(
                INVALID_PARAMS, f"{tool}: '{key}' must be between {spec['minimum']} and {spec['maximum']}"
            )
    return arguments


class McpServer:
    def __init__(self, registry: AgentRegistry):
        self.registry = registry
        self.initialized = False

    # -- tools -----------------------------------------------------------

    def _agent(self, name: str):
        agent = self.registry.get(name)
        if agent is None:
            raise ToolError(f"Agent not found: {name!r}. Use list_agents to see the {len(self.registry)} available.")
        return agent

    def tool_list_agents(self, args: dict) -> dict:
        agents = self.registry.list(tier=args.get("tier"), category=args.get("category"))
        return _text(
            self.registry.roster_markdown(agents),
            {"count": len(agents), "agents": [a.summary() for a in agents]},
        )

    def tool_find_agent(self, args: dict) -> dict:
        task = args["task_description"]
        if not task.strip():
            raise ToolError("task_description is empty.")
        matches = self.registry.find(task, limit=args.get("limit", 3))
        if not matches:
            text = (
                "No agent definition shares a meaningful word with that task. "
                "Rephrase it, or use list_agents to choose by hand."
            )
        else:
            shown = task if len(task) <= TASK_ECHO_CHARS else task[:TASK_ECHO_CHARS] + "..."
            lines = [f'## Agents matching: "{shown}"', ""]
            for m in matches:
                lines.append(
                    f"- **@{m['name']}** (score {m['score']}; matched: {', '.join(m['matched'])}) - {m['description']}"
                )
            lines += ["", "Ranking is by word overlap with each definition, not by meaning."]
            text = "\n".join(lines)
        return _text(text, {"matches": matches})

    def tool_get_agent(self, args: dict) -> dict:
        return _text(self._agent(args["name"]).content)

    def tool_invoke_agent(self, args: dict) -> dict:
        agent = self._agent(args["name"])
        task = (args.get("task") or "").strip()
        parts = [f"[{NO_MODEL_NOTE}]", "", f"Agent: @{agent.name}"]
        if task:
            parts += ["Task:", "", task]
        parts += ["", "---", "", agent.prompt]
        return _text(
            "\n".join(parts),
            {"agent": agent.name, "task": task or None, "executed": False, "prompt": agent.prompt},
        )

    # -- JSON-RPC methods ------------------------------------------------

    def rpc_initialize(self, params: dict) -> dict:
        requested = params.get("protocolVersion")
        version = requested if requested in PROTOCOL_VERSIONS else PROTOCOL_VERSIONS[0]
        self.initialized = True
        return {
            "protocolVersion": version,
            "capabilities": {"tools": {}, "resources": {}},
            "serverInfo": {"name": SERVER_NAME, "version": __version__},
            "instructions": (
                f"Serves the {len(self.registry)} Elite Agent Collective definitions. "
                "invoke_agent returns a prompt for you to run; this server calls no model."
            ),
        }

    def rpc_ping(self, params: dict) -> dict:
        return {}

    def rpc_tools_list(self, params: dict) -> dict:
        return {"tools": TOOLS}

    def rpc_tools_call(self, params: dict) -> dict:
        name = params.get("name")
        if not isinstance(name, str) or name not in _TOOL_SCHEMAS:
            raise RpcError(INVALID_PARAMS, f"Unknown tool: {name!r}")
        args = _check_arguments(name, params.get("arguments"))
        try:
            return getattr(self, "tool_" + name)(args)
        except ToolError as exc:
            return _text(str(exc), is_error=True)

    def rpc_resources_list(self, params: dict) -> dict:
        resources = [
            {
                "uri": ROSTER_URI,
                "name": "Agent roster",
                "description": "All agents with tier, category and role",
                "mimeType": "text/markdown",
            }
        ]
        for agent in self.registry.list():
            resources.append(
                {
                    "uri": RESOURCE_SCHEME + agent.name,
                    "name": "@" + agent.name,
                    "description": agent.description,
                    "mimeType": "text/markdown",
                }
            )
        return {"resources": resources}

    def rpc_resources_read(self, params: dict) -> dict:
        uri = params.get("uri")
        if not isinstance(uri, str):
            raise RpcError(INVALID_PARAMS, "resources/read: 'uri' must be a string")
        if uri == ROSTER_URI:
            text = self.registry.roster_markdown()
        else:
            agent = self.registry.get(uri[len(RESOURCE_SCHEME):]) if uri.startswith(RESOURCE_SCHEME) else None
            if agent is None:
                raise RpcError(RESOURCE_NOT_FOUND, "Resource not found", {"uri": uri})
            text = agent.content
        return {"contents": [{"uri": uri, "mimeType": "text/markdown", "text": text}]}

    METHODS = {
        "initialize": rpc_initialize,
        "ping": rpc_ping,
        "tools/list": rpc_tools_list,
        "tools/call": rpc_tools_call,
        "resources/list": rpc_resources_list,
        "resources/read": rpc_resources_read,
    }

    # -- dispatch --------------------------------------------------------

    def handle(self, message: Any) -> dict | None:
        """Handle one decoded message. Returns the response, or None for a notification."""
        if not isinstance(message, dict) or message.get("jsonrpc") != "2.0":
            return _error(None, INVALID_REQUEST, "Not a JSON-RPC 2.0 message")
        msg_id = message.get("id")
        # Only a string or a finite number is ever echoed back as an id; anything else answers with null.
        if msg_id is not None and (
            isinstance(msg_id, bool)
            or not isinstance(msg_id, (str, int, float))
            or (isinstance(msg_id, float) and not math.isfinite(msg_id))
        ):
            return _error(None, INVALID_REQUEST, "Id must be a string or a number")
        if "method" not in message:
            if "result" in message or "error" in message:
                return None  # a response to a request we never sent; ignore
            return _error(msg_id, INVALID_REQUEST, "Missing method")
        method = message["method"]
        if not isinstance(method, str):
            return _error(msg_id, INVALID_REQUEST, "Method must be a string")
        if "id" not in message:
            return None  # notifications (notifications/initialized, cancelled, ...) get no reply
        params = message.get("params")
        if params is None:
            params = {}
        if not isinstance(params, dict):
            return _error(msg_id, INVALID_PARAMS, "Params must be an object")
        handler = self.METHODS.get(method)
        if handler is None:
            return _error(msg_id, METHOD_NOT_FOUND, f"Method not found: {method}")
        if not self.initialized and method not in ("initialize", "ping"):
            return _error(msg_id, INVALID_REQUEST, "Server not initialized: send initialize first")
        try:
            return {"jsonrpc": "2.0", "id": msg_id, "result": handler(self, params)}
        except RpcError as exc:
            return _error(msg_id, exc.code, exc.message, exc.data)
        except Exception as exc:  # a bug here must not kill the session
            print(f"{SERVER_NAME}: internal error in {method}: {exc!r}", file=sys.stderr)
            return _error(msg_id, INTERNAL_ERROR, "Internal error")

    def handle_line(self, line: bytes) -> bytes | None:
        """Handle one line off the wire; returns the encoded reply line or None."""
        if not line.strip():
            return None
        try:
            message = json.loads(line.decode("utf-8"), parse_constant=_reject_constant, parse_float=_finite_float)
        except (UnicodeDecodeError, ValueError, RecursionError):  # RecursionError: absurdly deep nesting
            reply: Any = _error(None, PARSE_ERROR, "Parse error")
        else:
            if isinstance(message, list):  # JSON-RPC batch
                if not message:
                    reply = _error(None, INVALID_REQUEST, "Empty batch")
                else:
                    reply = [r for r in map(self.handle, message) if r is not None] or None
            else:
                reply = self.handle(message)
        if reply is None:
            return None
        # ensure_ascii keeps every message on one line and safe on any console codepage
        return json.dumps(reply, separators=(",", ":"), allow_nan=False).encode("ascii") + b"\n"

    def serve(self, stdin: BinaryIO, stdout: BinaryIO) -> None:
        """Read messages until stdin closes. Binary streams: no CRLF translation on Windows."""
        for line in iter(stdin.readline, b""):
            reply = self.handle_line(line)
            if reply is not None:
                stdout.write(reply)
                stdout.flush()


def _reject_constant(name: str) -> None:
    """NaN, Infinity and -Infinity are not JSON; Python's parser accepts them unless told otherwise."""
    raise ValueError(f"{name} is not valid JSON")


def _finite_float(text: str) -> float:
    """A number too large for a float (1e999) would become inf and could not be written back as JSON."""
    value = float(text)
    if not math.isfinite(value):
        raise ValueError("number out of range")
    return value


def _error(msg_id: Any, code: int, message: str, data: Any = None) -> dict:
    error: dict[str, Any] = {"code": code, "message": message}
    if data is not None:
        error["data"] = data
    return {"jsonrpc": "2.0", "id": msg_id, "error": error}


def main() -> int:
    try:
        registry = AgentRegistry().load()
    except (RegistryError, OSError, UnicodeDecodeError) as exc:
        print(f"{SERVER_NAME}: cannot start: {exc}", file=sys.stderr)
        return 1
    print(f"{SERVER_NAME} {__version__}: {len(registry)} agents from {registry.agents_dir}", file=sys.stderr)
    try:
        McpServer(registry).serve(sys.stdin.buffer, sys.stdout.buffer)
    except KeyboardInterrupt:
        pass
    return 0
