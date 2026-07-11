"""Per-capability eval tasks with deterministic checkers.

Capabilities mirror the LangChain Deep Agents split: tool-calling, retrieval
(grounding), multi-turn, summarization, instruction-following. Every checker is
deterministic (substring / regex / JSON parse) so the scoreboard is reproducible
and cheap — no LLM-as-judge.

A Task carries its baseline `messages`; when a harness profile is applied
(Increment 2/3) it may rewrite them, but the checker is unchanged.
"""
import re

from harness.gateway import extract_json


class Task:
    def __init__(self, id, capability, messages, check, desc="", context=None):
        self.id = id
        self.capability = capability
        self.messages = messages
        self.check = check          # (text, message) -> bool
        self.desc = desc
        self.context = context      # optional retrieved context, for profile inject-mode


TOOLS_SYS = (
    "You have these tools:\n"
    "- search_notes(query): search the knowledge base\n"
    "- read_file(path): read a file from disk\n"
    "- get_weather(city): current weather for a city\n"
    "When a tool is needed, respond with ONLY a JSON object of the form "
    '{"tool": "<name>", "args": {...}} and nothing else.'
)


def _norm(s):
    """Lowercase + fold typographic quotes to ASCII, so a curly apostrophe in
    'don’t' still matches a 'don't' marker (a real checker bug caught by the
    baseline run)."""
    return (s or "").lower().replace("’", "'").replace("‘", "'").replace("“", '"').replace("”", '"')


def _tool_is(obj, name):
    return isinstance(obj, dict) and str(obj.get("tool", "")).strip().lower() == name


def _args_str(obj):
    return str((obj or {}).get("args", "")).lower()


# --------------------------- tool-calling ---------------------------
def _c_weather(text, msg):
    o = extract_json(text)
    return _tool_is(o, "get_weather") and "paris" in _args_str(o)


def _c_search(text, msg):
    o = extract_json(text)
    return _tool_is(o, "search_notes") and len(_args_str(o)) > 2


def _c_readfile(text, msg):
    o = extract_json(text)
    return _tool_is(o, "read_file") and "/etc/hosts" in _args_str(o)


# --------------------------- retrieval / grounding ---------------------------
def _c_vaultkey(text, msg):
    t = text.lower()
    return ("0600" in t) or ("argon2id" in t) or ("aes-256" in t)


def _c_pm0(text, msg):
    return "pm0" in text.lower()


def _c_nohallucinate(text, msg):
    # The author of the fictional "Florbnix" module is NOT in the context, so a
    # grounded model must decline rather than invent one.
    t = _norm(text)
    markers = ["not mentioned", "not provided", "not in", "does not", "doesn't",
               "no author", "not specified", "not stated", "unknown", "don't know",
               "cannot", "can't", "no information", "not available", "not given",
               "isn't", "no mention"]
    return any(m in t for m in markers)


# --------------------------- multi-turn ---------------------------
def _c_teal(text, msg):
    return "teal" in text.lower()


def _c_codeword(text, msg):
    t = text.lower()
    return "falcon-9" in t or ("falcon" in t and "9" in t)


# --------------------------- summarization ---------------------------
_SUMM_TEXT = (
    "The sigma-index service is a Go server on port 8200 that provides hybrid "
    "search over the imh-notes namespace. It fuses HNSW vector search with an "
    "Okapi BM25 lexical index using Reciprocal Rank Fusion. It persists to disk "
    "via a write-ahead log and takes periodic snapshots. In-my-head, the vault, "
    "and now the Elite Agent Collective all read from it. The whole thing runs "
    "on a two-core CPU-only box with no GPU."
)


def _c_summary(text, msg):
    t = text.lower()
    hits = sum(1 for k in ["8200", "hybrid", "bm25", "rrf", "reciprocal", "hnsw", "cpu", "go"] if k in t)
    brief = len(text.split()) < 90
    return hits >= 2 and brief


_LONGTAIL_TEXT = (
    "This document describes the release process. First, engineers open a pull "
    "request. Reviewers check the diff and run the test suite. Continuous "
    "integration builds artifacts for every commit. Release managers tag a "
    "version once the checklist is green. Notes are drafted and circulated for "
    "sign-off. Many teams argue about cadence and ceremonies at length. "
    "After all of that, the single most important operational fact: the "
    "production deployment target is a Raspberry Pi 5 running Debian 13."
)


def _c_longtail(text, msg):
    t = text.lower()
    return ("raspberry pi" in t or "pi 5" in t) and ("debian" in t)


# --------------------------- instruction-following ---------------------------
def _c_oneword(text, msg):
    first = re.sub(r"[^a-z]", "", text.strip().lower().split()[0]) if text.strip() else ""
    return first == "tokyo" and len(text.split()) <= 4


def _c_json_format(text, msg):
    o = extract_json(text)
    return isinstance(o, dict) and str(o.get("status")) == "ok" and str(o.get("count")) == "3"


TASKS = [
    # tool-calling
    Task("tool_weather", "tool_calling",
         [{"role": "system", "content": TOOLS_SYS}, {"role": "user", "content": "What's the weather in Paris right now?"}],
         _c_weather, "emit get_weather(Paris)"),
    Task("tool_search", "tool_calling",
         [{"role": "system", "content": TOOLS_SYS}, {"role": "user", "content": "Find my notes about the vault encryption design."}],
         _c_search, "emit search_notes(query)"),
    Task("tool_readfile", "tool_calling",
         [{"role": "system", "content": TOOLS_SYS}, {"role": "user", "content": "Show me the contents of /etc/hosts."}],
         _c_readfile, "emit read_file(/etc/hosts)"),

    # retrieval / grounding
    Task("ret_vaultkey", "retrieval",
         [{"role": "system", "content": "Answer ONLY from the provided context."},
          {"role": "user", "content": "Context: The vault master key lives in a 0600 file and is never committed or logged. It uses AES-256-GCM with Argon2id.\n\nQuestion: How is the vault master key protected?"}],
         _c_vaultkey, "grounded answer from context", context="The vault master key lives in a 0600 file and is never committed or logged. It uses AES-256-GCM with Argon2id."),
    Task("ret_pm0", "retrieval",
         [{"role": "system", "content": "Answer ONLY from the provided context."},
          {"role": "user", "content": "Context: PhantomMesh-VPN uses X25519 for the handshake and runs on the pm0 interface.\n\nQuestion: Which network interface does PhantomMesh use?"}],
         _c_pm0, "grounded answer from context", context="PhantomMesh-VPN uses X25519 for the handshake and runs on the pm0 interface."),
    Task("ret_nohalluc", "retrieval",
         [{"role": "system", "content": "Answer ONLY from the provided context. If the answer is not in the context, say you don't know."},
          {"role": "user", "content": "Context: The Florbnix module handles glyph routing and was added in version 3.2.\n\nQuestion: Who is the author of the Florbnix module?"}],
         _c_nohallucinate, "decline when answer not in context", context="The Florbnix module handles glyph routing and was added in version 3.2."),

    # multi-turn
    Task("mt_color", "multi_turn",
         [{"role": "user", "content": "My favorite color is teal."},
          {"role": "assistant", "content": "Got it, teal."},
          {"role": "user", "content": "What is my favorite color?"}],
         _c_teal, "recall earlier turn"),
    Task("mt_codeword", "multi_turn",
         [{"role": "user", "content": "Remember: the code word is falcon-9."},
          {"role": "assistant", "content": "Understood."},
          {"role": "user", "content": "What was the code word I gave you?"}],
         _c_codeword, "recall earlier turn"),

    # summarization
    Task("sum_index", "summarization",
         [{"role": "user", "content": "Summarize the following in two sentences:\n\n" + _SUMM_TEXT}],
         _c_summary, "cover key facts, stay brief"),
    Task("sum_longtail", "summarization",
         [{"role": "user", "content": _LONGTAIL_TEXT + "\n\nQuestion: What is the production deployment target?"}],
         _c_longtail, "read to the end, not just the head"),

    # instruction-following
    Task("if_oneword", "instruction_following",
         [{"role": "user", "content": "Respond with exactly one word: the capital of Japan."}],
         _c_oneword, "obey format, report the concrete answer"),
    Task("if_json", "instruction_following",
         [{"role": "user", "content": 'Output a JSON object with keys "status" set to "ok" and "count" set to 3. Output only the JSON, nothing else.'}],
         _c_json_format, "obey structured-output format"),
]


def capabilities():
    seen = []
    for t in TASKS:
        if t.capability not in seen:
            seen.append(t.capability)
    return seen
