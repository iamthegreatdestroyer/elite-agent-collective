"""Automated harness tuner — the 'ralph loop'.

Evaluate -> diagnose -> propose a single-purpose profile edit -> re-evaluate ->
keep the edit only if it improves the score without regressing another
capability. Scoped write: touches ONLY harness/profile_tuned.json. A held-out
task split is never tuned on and validates generalization at the end.

The proposer here is a curated failure->block rule-base, so the loop runs fully
autonomously on the box with no frontier model. Swap `propose()` for an LLM /
Workflow-agent proposer to match the playbook's 'frontier model proposes edits'
upgrade path (the trajectories are already collected per run).
"""
import argparse
import json

from harness import profiles
from harness.eval import runner
from harness.eval.tasks import TASKS

TUNED_JSON = profiles.TUNED_JSON

# One single-purpose block per capability — "target one observed failure with a
# specific behavioral rule", not a broad rewrite.
BLOCKS = {
    "retrieval": "If the answer is not in the provided context, say you do not know. Never invent facts, names, or values that are not in the context.",
    "tool_calling": "When you call a tool, put every argument inside an \"args\" object, e.g. {\"tool\":\"search_notes\",\"args\":{\"query\":\"...\"}}. Output only that JSON, with no explanation.",
    "instruction_following": "Follow output-format instructions literally: if asked for exactly one word, or for JSON only, output only that and nothing else.",
    "summarization": "Read the entire text through to the final sentence before answering; the most important fact may appear at the end.",
    "multi_turn": "Use facts the user stated earlier in this conversation when answering.",
}

# Held out from tuning; used only for the final generalization check.
HOLDOUT_IDS = ["tool_readfile", "ret_pm0", "if_json", "sum_longtail"]


def read_json():
    try:
        with open(TUNED_JSON) as f:
            return json.load(f)
    except Exception:
        return {"inject_mode": "message", "correct_tool_calls": True, "prompt_suffix": "", "per_model_suffix": {}}


def write_json(cfg):
    with open(TUNED_JSON, "w") as f:
        json.dump(cfg, f, indent=2)


def pass_by_cap(records):
    caps = {}
    for r in records:
        c = r["capability"]
        caps.setdefault(c, [0, 0])
        caps[c][1] += 1
        if r.get("passed"):
            caps[c][0] += 1
    total = sum(v[0] for v in caps.values())
    return caps, total


def evaluate(model, ids, max_tokens, profile):
    recs = runner.run([model], 1, max_tokens, task_filter=set(ids), profile=profile)
    return pass_by_cap(recs)


def propose(caps, current_suffix):
    # most-failing capability not already covered by a block
    ranked = sorted(caps.items(), key=lambda kv: kv[1][0] - kv[1][1])  # (pass - total) ascending
    for cap, (p, t) in ranked:
        if p < t and cap in BLOCKS and BLOCKS[cap] not in current_suffix:
            return cap, BLOCKS[cap]
    return None, None


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", default="gemma3:4b-it-qat")
    ap.add_argument("--max-iters", type=int, default=3)
    ap.add_argument("--max-tokens", type=int, default=180)
    a = ap.parse_args()

    tune_ids = [t.id for t in TASKS if t.id not in HOLDOUT_IDS]

    # Start from a clean suffix (reproducible), keep inject-mode + correction on.
    cfg = read_json()
    cfg["prompt_suffix"] = ""
    write_json(cfg)

    print("[tuner] model=%s  tune_tasks=%d  holdout=%s" % (a.model, len(tune_ids), HOLDOUT_IDS), flush=True)
    caps, total = evaluate(a.model, tune_ids, a.max_tokens, profiles.get("tuned"))
    denom = sum(v[1] for v in caps.values())
    print("[tuner] iter0 TUNE=%d/%d %s" % (total, denom, {k: tuple(v) for k, v in caps.items()}), flush=True)

    for it in range(1, a.max_iters + 1):
        cfg = read_json()
        cap, block = propose(caps, cfg.get("prompt_suffix", ""))
        if not block:
            print("[tuner] iter%d: no proposal left — converged" % it, flush=True)
            break
        prev = cfg.get("prompt_suffix", "")
        cfg["prompt_suffix"] = (prev + "\n" + block).strip()
        write_json(cfg)
        print("[tuner] iter%d: propose block for '%s': %s..." % (it, cap, block[:60]), flush=True)

        new_caps, new_total = evaluate(a.model, tune_ids, a.max_tokens, profiles.get("tuned"))
        regress = any(new_caps.get(c, [0, 0])[0] < caps.get(c, [0, 0])[0] for c in caps)
        if new_total > total and not regress:
            print("[tuner] iter%d: ACCEPT (TUNE %d -> %d, no regression)" % (it, total, new_total), flush=True)
            caps, total = new_caps, new_total
        else:
            print("[tuner] iter%d: REJECT (TUNE %d -> %d, regress=%s) — revert" % (it, total, new_total, regress), flush=True)
            cfg["prompt_suffix"] = prev
            write_json(cfg)

    print("[tuner] --- holdout generalization (never tuned on) ---", flush=True)
    _, base_pass = evaluate(a.model, HOLDOUT_IDS, a.max_tokens, None)
    _, tuned_pass = evaluate(a.model, HOLDOUT_IDS, a.max_tokens, profiles.get("tuned"))
    print("[tuner] HOLDOUT baseline=%d/%d  tuned=%d/%d" % (base_pass, len(HOLDOUT_IDS), tuned_pass, len(HOLDOUT_IDS)), flush=True)
    print("[tuner] learned prompt_suffix:\n%s" % read_json().get("prompt_suffix", ""), flush=True)


if __name__ == "__main__":
    main()
