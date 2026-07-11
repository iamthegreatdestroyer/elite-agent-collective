"""Fast, no-network self-tests for the harness's deterministic bits.

Runs in milliseconds (no model calls), so it's safe to run while an eval sweep
occupies the box. Validates extract_json, the task checkers, the middleware, and
the profile's point-of-need injection.
"""
from harness.gateway import extract_json
from harness.eval.tasks import TASKS
from harness import middleware, profiles


def _task(tid):
    return next(t for t in TASKS if t.id == tid)


def main():
    # extract_json robustness
    assert extract_json('prefix {"a": 1, "b": {"c": 2}} suffix') == {"a": 1, "b": {"c": 2}}
    assert extract_json("no json here") is None
    assert extract_json('```json\n{"tool":"x","args":{}}\n```')["tool"] == "x"

    # checkers fire correctly on good vs bad text
    assert _task("tool_weather").check('{"tool":"get_weather","args":{"city":"Paris"}}', None) is True
    assert _task("tool_weather").check('{"tool":"search_notes","args":{}}', None) is False
    assert _task("tool_readfile").check('{"tool":"read_file","args":{"path":"/etc/hosts"}}', None) is True
    assert _task("ret_pm0").check("It runs on the pm0 interface.", None) is True
    assert _task("ret_pm0").check("It runs on eth0.", None) is False
    assert _task("ret_nohalluc").check("The author is not mentioned in the context.", None) is True
    assert _task("ret_nohalluc").check("I don’t know.", None) is True  # curly apostrophe must still match
    assert _task("ret_nohalluc").check("The author is Jane Doe.", None) is False
    assert _task("mt_codeword").check("The code word was falcon-9.", None) is True
    assert _task("if_oneword").check("Tokyo", None) is True
    assert _task("if_oneword").check("The capital of Japan is Tokyo, a large city.", None) is False
    assert _task("if_json").check('{"status":"ok","count":3}', None) is True
    assert _task("if_json").check('{"status":"bad","count":3}', None) is False
    assert _task("sum_longtail").check("The production deployment target is a Raspberry Pi 5 running Debian 13.", None) is True

    # continuation-notice middleware
    assert "per-read limit" in middleware.continuation_notice("read_file", "l1\nl2", returned=100, limit=100)
    assert "per-read limit" not in middleware.continuation_notice("read_file", "short", returned=3, limit=100)
    assert middleware.is_valid_tool_call('{"tool":"x","args":{}}') is True
    assert middleware.is_valid_tool_call('{"tool":"read_file","path":"/x"}') is False  # flattened args -> triggers correction
    assert middleware.is_valid_tool_call("sorry, I cannot") is False

    # profile point-of-need injection: context leaves the user turn, question stays
    p = profiles.get("tuned")
    msgs = p.apply(_task("ret_pm0"), "granite4:1b")
    joined = " ".join(m["content"] for m in msgs)
    assert "pm0" in joined, "context must be preserved"
    assert any(m["role"] == "user" and m["content"].startswith("Relevant reference") for m in msgs), "context re-injected as its own message"
    assert any(m["role"] == "user" and "Question:" in m["content"] for m in msgs), "question separated"

    # non-context tasks are unaffected by inject-mode
    tw = p.apply(_task("tool_weather"), "granite4:1b")
    assert len(tw) == len(_task("tool_weather").messages)

    print("SELFTEST OK — %d tasks; extract_json / checkers / middleware / point-of-need inject all pass" % len(TASKS))


if __name__ == "__main__":
    main()
