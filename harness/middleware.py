"""Reusable harness middleware — the point-of-need + validation layer.

These are the "engineer the environment, not the weights" tactics from the
LangChain Nemotron playbook: mechanical signal injection at the point of need,
and one-shot tool-call correction. All fail-open.
"""
from harness.gateway import content_of, extract_json


def continuation_notice(tool_name, result_text, returned, limit, offset=0):
    """Append a mechanical continuation hint when a paginated read hits its
    per-read limit (the ReadFileContinuationNotice pattern). Mechanical signal,
    not semantic clarification — which the playbook found generalizes better and
    avoids overfitting. Returns result_text unchanged when not at the limit.
    """
    if limit and returned >= limit:
        return (result_text or "") + (
            "\n[%s returned %d items starting at offset %d, the per-read limit. "
            "More results likely exist past this window; read again to continue.]"
            % (tool_name, returned, offset)
        )
    return result_text


TOOL_CORRECTION = (
    "Your previous reply was not a valid tool call. Put EVERY argument inside an "
    '"args" object, e.g. {"tool": "read_file", "args": {"path": "/x"}}. Respond '
    "with ONLY that JSON and nothing else."
)


def is_valid_tool_call(text):
    """A valid tool call has a tool name AND an args object. Calls that flatten
    arguments to the top level (e.g. {"tool":"read_file","path":"/x"}) are treated
    as invalid so the correction middleware can re-prompt — the exact failure the
    baseline surfaced for granite4:1b."""
    o = extract_json(text)
    return isinstance(o, dict) and "tool" in o and isinstance(o.get("args"), dict)


def correct_tool_call(messages, model, chat_fn, max_tokens=256):
    """One-shot correction: re-prompt the model to emit valid tool-call JSON.
    Returns the corrected text, or "" on failure. Best-effort / fail-open.
    """
    try:
        convo = list(messages) + [{"role": "user", "content": TOOL_CORRECTION}]
        msg, _finish, _dt = chat_fn(convo, model, max_tokens=max_tokens, temperature=0.0)
        return content_of(msg)
    except Exception:
        return ""
