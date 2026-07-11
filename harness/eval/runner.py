"""Eval runner: score local models per-capability through the gateway.

Loops model-outer / task-inner so each model is loaded into the single-GPU-less
ollama slot exactly once (model switches are the expensive part on the box).

Usage (from the EAC repo root):
  python3 -m harness.eval.runner --models gemma3:4b-it-qat,phi4-mini,granite4:1b
  python3 -m harness.eval.runner --models granite4:1b --tasks tool_weather --label smoke

A harness profile (Increment 2/3) can be applied with --profile <name>; without
it, the baseline task messages are used verbatim.
"""
import argparse
import json
import os
import time

from harness.gateway import chat, content_of
from harness.eval.tasks import TASKS, capabilities

DEFAULT_MODELS = ["gemma3:4b-it-qat", "phi4-mini", "granite4:1b"]


def _load_profile(name):
    if not name:
        return None
    from harness import profiles  # lazy: profiles.py arrives with Increment 2
    return profiles.get(name)


def run(models, runs, max_tokens, task_filter=None, cap_filter=None, profile=None):
    tasks = TASKS
    if task_filter:
        tasks = [t for t in tasks if t.id in task_filter]
    if cap_filter:
        tasks = [t for t in tasks if t.capability in cap_filter]

    records = []
    for model in models:  # outer: pay each model-load once
        for t in tasks:
            messages = profile.apply(t, model) if profile else t.messages
            for r in range(runs):
                rec = {"model": model, "task": t.id, "capability": t.capability, "run": r}
                try:
                    msg, finish, dt = chat(messages, model, max_tokens=max_tokens, temperature=0.0)
                    text = content_of(msg)
                    if profile:
                        text, msg = profile.post(t, model, text, msg, messages=messages, max_tokens=max_tokens)
                    passed = bool(t.check(text, msg))
                    rec.update(passed=passed, finish=finish, latency=round(dt, 2), response=text[:800])
                except Exception as e:  # never let one call crash the sweep
                    rec.update(passed=False, error=str(e)[:200], latency=None, response="")
                records.append(rec)
                print("  %-20s %-14s run%d -> %s (%ss)" % (
                    model, t.id, r, "PASS" if rec.get("passed") else "fail", rec.get("latency")), flush=True)
    return records


def summarize(records, models):
    caps = capabilities()
    summary = {}
    for m in models:
        summary[m] = {}
        for c in caps:
            sub = [r for r in records if r["model"] == m and r["capability"] == c]
            summary[m][c] = {"pass": sum(1 for r in sub if r.get("passed")), "total": len(sub)}
        allsub = [r for r in records if r["model"] == m]
        summary[m]["_overall"] = {"pass": sum(1 for r in allsub if r.get("passed")), "total": len(allsub)}
    return summary, caps


def markdown(summary, caps, models):
    cols = caps + ["_overall"]
    lines = ["| model | " + " | ".join(c.replace("_", " ") for c in cols) + " |",
             "|" + "|".join(["---"] * (len(cols) + 1)) + "|"]
    for m in models:
        cells = ["%d/%d" % (summary[m][c]["pass"], summary[m][c]["total"]) for c in cols]
        lines.append("| %s | %s |" % (m, " | ".join(cells)))
    return "\n".join(lines)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--models", default=",".join(DEFAULT_MODELS))
    ap.add_argument("--runs", type=int, default=1)
    ap.add_argument("--max-tokens", type=int, default=200)
    ap.add_argument("--tasks", default="")
    ap.add_argument("--caps", default="")
    ap.add_argument("--profile", default="")
    ap.add_argument("--out", default="harness/eval/scoreboard.json")
    ap.add_argument("--label", default="baseline")
    a = ap.parse_args()

    models = [m.strip() for m in a.models.split(",") if m.strip()]
    tf = set(x.strip() for x in a.tasks.split(",") if x.strip()) or None
    cf = set(x.strip() for x in a.caps.split(",") if x.strip()) or None
    profile = _load_profile(a.profile)

    t0 = time.time()
    records = run(models, a.runs, a.max_tokens, tf, cf, profile)
    summary, caps = summarize(records, models)
    md = markdown(summary, caps, models)
    out = {
        "label": a.label, "profile": a.profile or None, "models": models,
        "runs": a.runs, "max_tokens": a.max_tokens, "elapsed_s": round(time.time() - t0, 1),
        "summary": summary, "records": records,
    }
    if os.path.dirname(a.out):
        os.makedirs(os.path.dirname(a.out), exist_ok=True)
    with open(a.out, "w") as f:
        json.dump(out, f, indent=2)

    print("\n=== SCOREBOARD (%s%s) ===" % (a.label, "/" + a.profile if a.profile else ""))
    print(md)
    print("\nwrote %s  (%.1fs)" % (a.out, out["elapsed_s"]))


if __name__ == "__main__":
    main()
