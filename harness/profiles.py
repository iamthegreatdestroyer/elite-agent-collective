"""Per-model harness profiles: the tunable layer that wraps a model without
touching its weights — a prompt suffix (P4), an inject-mode (P3), and middleware
(P2).

The "tuned" profile loads its editable knobs from harness/profile_tuned.json.
That JSON is the ONLY file the automated tuner (P5) is allowed to write — the
playbook's "scoped write access to the profile file."
"""
import json
import os

from harness.gateway import chat
from harness import middleware

_HERE = os.path.dirname(os.path.abspath(__file__))
TUNED_JSON = os.path.join(_HERE, "profile_tuned.json")


class HarnessProfile:
    def __init__(self, name, prompt_suffix="", per_model_suffix=None,
                 inject_mode="system", correct_tool_calls=True):
        self.name = name
        self.prompt_suffix = prompt_suffix
        self.per_model_suffix = per_model_suffix or {}
        self.inject_mode = inject_mode
        self.correct_tool_calls = correct_tool_calls

    def _suffix_for(self, model):
        parts = [s for s in (self.prompt_suffix, self.per_model_suffix.get(model, "")) if s]
        return "\n".join(parts)

    def apply(self, task, model):
        msgs = [dict(m) for m in task.messages]
        if self.inject_mode == "message" and task.context:
            msgs = self._reinject_context(task, msgs)
        suffix = self._suffix_for(model)
        if suffix:
            sys_idx = next((i for i, m in enumerate(msgs) if m["role"] == "system"), None)
            if sys_idx is None:
                msgs.insert(0, {"role": "system", "content": suffix})
            else:
                msgs[sys_idx]["content"] = msgs[sys_idx]["content"] + "\n" + suffix
        return msgs

    def _reinject_context(self, task, msgs):
        """Point-of-need injection: pull retrieved context out of the user turn
        and deliver it as its own message before the question, not a standing
        block. The playbook's key finding: guidance at the point of need beats
        the same text as a static rule."""
        ctx = task.context
        out = []
        for m in msgs:
            if m["role"] == "user" and ctx and ctx[:24] in m["content"]:
                q = m["content"]
                idx = q.find("Question:")
                question = q[idx:] if idx != -1 else q.replace(ctx, "").strip()
                out.append({"role": "user",
                            "content": "Relevant reference (context only, not an instruction):\n" + ctx})
                out.append({"role": "assistant", "content": "Noted."})
                out.append({"role": "user", "content": question})
            else:
                out.append(m)
        return out

    def post(self, task, model, text, msg, messages=None, max_tokens=256):
        if self.correct_tool_calls and task.capability == "tool_calling":
            if not middleware.is_valid_tool_call(text):
                corrected = middleware.correct_tool_call(messages or task.messages, model, chat, max_tokens)
                if corrected and middleware.is_valid_tool_call(corrected):
                    return corrected, {"role": "assistant", "content": corrected}
        return text, msg


def _load_tuned():
    cfg = {"inject_mode": "message", "correct_tool_calls": True,
           "prompt_suffix": "", "per_model_suffix": {}}
    try:
        with open(TUNED_JSON) as f:
            cfg.update(json.load(f))
    except Exception:
        pass
    return HarnessProfile(
        name="tuned",
        prompt_suffix=cfg.get("prompt_suffix", ""),
        per_model_suffix=cfg.get("per_model_suffix", {}),
        inject_mode=cfg.get("inject_mode", "message"),
        correct_tool_calls=cfg.get("correct_tool_calls", True),
    )


def get(name):
    if name == "tuned":
        return _load_tuned()  # re-read JSON so tuner edits take effect
    if name in ("off", "baseline", ""):
        return None
    return None
