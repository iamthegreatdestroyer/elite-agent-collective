#!/usr/bin/env python3
"""Agent accuracy eval — grades whether answers are RIGHT, not whether they contain keywords.

WHY THIS EXISTS
    Every regex metric used on 2026-08-11 misled at least once. A marker count called a
    genuine improvement a regression, because truncated rambling hits more markers by volume
    than a correct one-line answer. Markers measure LENGTH and VOCABULARY. This measures the
    property we actually care about.

HOW IT AVOIDS THE SAME TRAP
    Questions are chosen so the correct answer is SHORT AND UNAMBIGUOUS — a number, a level,
    a yes/no, or an explicit "none". Grading then needs no judgement and no LLM judge (which
    would be the model marking its own work). If a question cannot be graded this way it does
    not belong in the set.

THREE OUTCOMES, NOT TWO
    correct  — asserts the right value
    wrong    — asserts a specific WRONG value (fabrication)
    abstain  — declines / says it does not know

    Collapsing wrong and abstain would hide the finding that matters. On 2026-08-11 the
    gateway said "AgentOS may be proprietary or fictional" (abstain, honest) while the
    retrieval path invented an "L5 admin" tier (wrong, confident). Those are not the same
    failure and must never score the same.

    A system that abstains is usable. A system that fabricates is worse than one that
    refuses, because it cannot be told apart from one that knows.
"""
import argparse, json, re, subprocess, sys, time

AGENT = "http://10.88.0.1:29080/agents/%s/invoke"
GATEWAY = "http://localhost:8000/v1/chat/completions"

# Each item: (id, question, right, wrong, note)
#   right : regex that must match for CORRECT
#   wrong : regex whose match means a WRONG assertion was made
# Answers are facts about THIS deployment, verifiable in the repo or on the live service.
CASES = [
    ("dispatch_sites",
     "How many agent dispatch sites does the authority ladder gate? Reply with just the number.",
     r"\b(6|six)\b", r"\b(2|3|4|5|7|8|9|two|three|four|five|seven|eight)\b",
     "six .Handle( call sites are guarded"),

    ("artifact_holders",
     "Which agents currently hold the artifact.write capability? If none hold it, say 'none'.",
     r"\bnone\b|\bnobody\b|\bno agents?\b|not granted|empty",
     r"\b(FORGE|APEX|SENTRY|VERTEX|STREAM|ATLAS|OMNISCIENT|SCRIBE)\b\s*(holds|has|is granted)",
     "the grant table is empty"),

    ("tier4_ceiling",
     "What is the authority ceiling for a tier 4 agent? Answer with one of: L0, L1, L2, L3, L4.",
     r"\bL2\b", r"\bL(0|1|3|4|5)\b",
     "tier 4 (OMNISCIENT) is capped at L2"),

    ("approval_executes",
     "Does approving an L2 proposal execute anything in this system? Answer yes or no, then one sentence.",
     r"\bno\b|does not execute|executes nothing|not execute",
     r"^\s*yes\b|\byes,? it (does|executes)",
     "approval records a decision; there is no tool surface"),

    ("echo_hash",
     "What hash function does the Echo chain use? Name it.",
     r"BLAKE2b", r"SHA-?256|SHA-?1\b|MD5|SHA-?512",
     "H(x) = BLAKE2b-256 per the primitive law"),

    ("failclosed_level",
     "Is the audit policy fail-open or fail-closed for L3 requests? Answer with one word.",
     r"fail-?closed", r"fail-?open",
     "fail-open below L3, fail-closed from L3 up"),

    ("echo_env",
     "Which environment variable enables the Echo audit chain? Give just the variable name.",
     r"EAC_ECHO_PATH", r"EAC_ECHO_ENABLED|ECHO_ENABLED|EAC_AUDIT|ENABLE_ECHO",
     "EAC_ECHO_PATH"),

    ("chain_contents",
     "Does the Echo chain store the text of user prompts? Answer yes or no.",
     r"\bno\b|never|does not (store|contain)|hashes only|metadata only",
     r"^\s*yes\b|\byes,? it (does|stores)",
     "hashes and metadata only — a canary test asserts this"),
]

ABSTAIN = re.compile(
    r"i (don'?t|do not) (have|know)|no (specific )?information|unable to (determine|find)|"
    r"not (enough|sufficient) (context|information)|cannot determine|unclear from|"
    r"would need (more|additional)|may refer to|if you (can )?provide", re.I)


def ask_agent(agent, q, timeout):
    body = json.dumps({"messages": [{"role": "user", "content": q}], "stream": False})
    p = subprocess.run(["curl", "-s", "-m", str(timeout), "-X", "POST", AGENT % agent,
                        "-H", "Content-Type: application/json", "-d", body],
                       capture_output=True, text=True)
    try:
        return json.loads(p.stdout)["choices"][0]["message"]["content"]
    except Exception:
        return ""


def ask_gateway(q, timeout):
    body = json.dumps({"model": "ling-mini-2.0", "max_tokens": 512,
                       "messages": [{"role": "user", "content": q}], "stream": False})
    p = subprocess.run(["curl", "-s", "-m", str(timeout), "-X", "POST", GATEWAY,
                        "-H", "Content-Type: application/json", "-d", body],
                       capture_output=True, text=True)
    try:
        return json.loads(p.stdout)["choices"][0]["message"]["content"]
    except Exception:
        return ""


def grade(ans, right, wrong):
    """CORRECT beats WRONG: an answer may mention other levels while asserting the right one.
    A pure abstention only counts when nothing was asserted either way."""
    if not ans.strip():
        return "error"
    r = re.search(right, ans, re.I) is not None
    w = re.search(wrong, ans, re.I) is not None
    if r and not w:
        return "correct"
    if r and w:
        return "mixed"      # states the right value AND a wrong one — not a pass
    if w:
        return "wrong"
    if ABSTAIN.search(ans):
        return "abstain"
    return "miss"           # said neither — off-topic or evasive without saying so


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--agent", default="APEX")
    ap.add_argument("--path", choices=["agent", "gateway", "both"], default="both")
    ap.add_argument("--timeout", type=int, default=180)
    ap.add_argument("--out", default="/tmp/agent_eval_answers.txt")
    a = ap.parse_args()

    paths = ["agent", "gateway"] if a.path == "both" else [a.path]
    tally = {p: {} for p in paths}
    fh = open(a.out, "w")

    print("%-18s | %-9s | %-9s | note" % ("case", "agent", "gateway"))
    print("-" * 76)
    for cid, q, right, wrong, note in CASES:
        row = {}
        for p in paths:
            t0 = time.time()
            ans = ask_agent(a.agent, q, a.timeout) if p == "agent" else ask_gateway(q, a.timeout)
            row[p] = grade(ans, right, wrong)
            tally[p][row[p]] = tally[p].get(row[p], 0) + 1
            fh.write("### %s [%s] -> %s (%.0fs)\n%s\n\n" % (cid, p, row[p], time.time() - t0, ans))
        print("%-18s | %-9s | %-9s | %s" % (cid, row.get("agent", "-"), row.get("gateway", "-"), note))

    fh.close()
    print("-" * 76)
    for p in paths:
        t = tally[p]
        n = sum(t.values())
        print("%-8s correct=%d  wrong=%d  mixed=%d  abstain=%d  miss=%d  error=%d   (n=%d)" % (
            p, t.get("correct", 0), t.get("wrong", 0), t.get("mixed", 0),
            t.get("abstain", 0), t.get("miss", 0), t.get("error", 0), n))
    print("\nanswers: %s" % a.out)
    print("NOTE: wrong != abstain. A system that declines is usable; one that fabricates")
    print("      cannot be told apart from one that knows.")


if __name__ == "__main__":
    main()
