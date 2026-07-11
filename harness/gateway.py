"""Minimal stdlib client for the Ryzanstein OpenAI-compatible gateway.

No third-party deps (the box + the ecosystem's no-bloat ethos): urllib only.
"""
import json
import os
import time
import urllib.error
import urllib.request

GATEWAY_URL = os.getenv("EAC_GATEWAY_URL", "http://localhost:8000")


def chat(messages, model, tools=None, max_tokens=256, temperature=0.0, timeout=180):
    """POST /v1/chat/completions.

    Returns (message_dict, finish_reason, latency_seconds). Raises on transport
    or non-2xx errors so callers can record the failure.
    """
    body = {
        "model": model,
        "messages": messages,
        "stream": False,
        "temperature": temperature,
        "max_tokens": max_tokens,
    }
    if tools:
        body["tools"] = tools
    data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(
        GATEWAY_URL + "/v1/chat/completions",
        data=data,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    t0 = time.time()
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        payload = json.loads(resp.read().decode("utf-8"))
    dt = time.time() - t0
    choice = payload["choices"][0]
    return choice.get("message", {}), choice.get("finish_reason", ""), dt


def content_of(message):
    return (message or {}).get("content", "") or ""


def extract_json(text):
    """Return the first parseable JSON object found in text, or None.

    Small models wrap JSON in prose / code fences, so scan for the first '{'
    and brace-match to the close, then json.loads.
    """
    if not text:
        return None
    start = text.find("{")
    while start != -1:
        depth = 0
        in_str = False
        esc = False
        for i in range(start, len(text)):
            ch = text[i]
            if in_str:
                if esc:
                    esc = False
                elif ch == "\\":
                    esc = True
                elif ch == '"':
                    in_str = False
                continue
            if ch == '"':
                in_str = True
            elif ch == "{":
                depth += 1
            elif ch == "}":
                depth -= 1
                if depth == 0:
                    candidate = text[start : i + 1]
                    try:
                        return json.loads(candidate)
                    except Exception:
                        break
        start = text.find("{", start + 1)
    return None
