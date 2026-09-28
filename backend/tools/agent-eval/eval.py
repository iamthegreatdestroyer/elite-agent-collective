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
import argparse, json, os, re, subprocess, sys, time

AGENT = "http://10.88.0.1:29080/agents/%s/invoke"
GATEWAY = "http://localhost:8000/v1/chat/completions"

# The --path gateway model. It was hardcoded to "ling-mini-2.0", which NOTHING SERVES: the
# gateway does not error on an unknown model, it silently substitutes its default and returns
# HTTP 200 with a different name in the body. That one name silently broke the weekly synth for
# two months and eac-backend for weeks. (It is not even a typo - ling-mini-2.0 is a real 9.9 GB
# model, on INFER; the gateway routing table just never mapped it to the host that has it.)
# 🔴 THE EVAL WAS READING ITS OWN CACHED ANSWERS. It asks the SAME 8 fixed questions every
# night, and the gateway sits behind the Token Recycler, whose L1 layer is EXACT-match. Measured
# 2026-09-27: both eval questions returned x-cache: hit, x-cache-layer: l1-exact. So after the
# first run this stopped being a measurement and became a REPLAY - it would keep emitting a
# stable, plausible score long after the system under test had changed, in either direction.
#
# Cache-Control: no-store is NOT sufficient - measured: the gateway honours it for STORING only
# and still serves a hit on lookup. So each run salts its prompts, which changes the L1 exact
# key. Verified: salted + no-store -> x-cache: miss, and L2 semantic did not rescue the hit.
# no-store is sent as well so these one-off salted answers do not pollute the cache for anyone
# else. Set EAC_EVAL_SALT to reproduce a specific run.
RUN_SALT = os.environ.get("EAC_EVAL_SALT") or ("%06x" % (int(time.time()) & 0xffffff))

def _fresh(q):
    """The question, made unique to this run so an exact-match cache cannot answer it."""
    return "%s\n\n(eval run %s)" % (q, RUN_SALT)

GATEWAY_MODEL = os.environ.get("EAC_EVAL_GATEWAY_MODEL",
                               "hf.co/inclusionAI/Ling-mini-2.0-GGUF:Q4_K_M")

# Each item: (id, question, right, wrong, note)
#   right : regex that must match for CORRECT
#   wrong : regex whose match means a WRONG assertion was made
# Answers are facts about THIS deployment, verifiable in the repo or on the live service.
CASES = [
    ("dispatch_sites",
     "How many agent dispatch sites does the authority ladder gate? Reply with just the number.",
     r"\b(6|six)\b",
     # ANY count that is not six. Was a hand-listed 2-9 with no 1 and no 0, so when the model
     # answered "1" it matched neither pattern and was graded MISS -- a wrong answer recorded
     # as an abstention, three runs running. Bounded to 0-20 plus number-words deliberately:
     # a bare \\d+ would also match version strings and IDs inside a prose answer.
     r"\b(?!6\b)(?:1?[0-9]|20|zero|one|two|three|four|five|seven|eight|nine|ten)\b",
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
     r"BLAKE2b",
     # An ASSERTION of the wrong hash, not a mention of one. The 2026-08-11
     # re-run graded a fully correct answer "mixed" because it said BLAKE2b
     # "offers advantages over older algorithms like SHA-2 or MD5" — a true
     # comparative aside that the old pattern read as a competing claim.
     # Naming a rival algorithm while correctly rejecting it is what a good
     # answer looks like; the grader must not punish it.
     r"(?:uses?|using|is|employs?|based on|implemented with|computed with|via)\s+"
     r"(?:the\s+)?(?:SHA-?256|SHA-?1\b|MD5|SHA-?512|SHA-?3)",
     "H(x) = BLAKE2b-256 per the primitive law"),

    ("failclosed_level",
     "Is the audit policy fail-open or fail-closed for L3 requests? Answer with one word.",
     r"fail-?closed", r"fail-?open",
     "fail-open below L3, fail-closed from L3 up"),

    ("echo_env",
     "Which environment variable enables the Echo audit chain? Give just the variable name.",
     r"EAC_ECHO_PATH",
     # Same rule: the variable must be PUT FORWARD as the answer, not listed
     # among things it is not.
     # ANY env-var-shaped identifier that is not EAC_ECHO_PATH -- but ONLY where it is PUT
     # FORWARD as the answer. That assertion context is the original's good idea and a first
     # attempt at this fix discarded it, which graded "The variable is EAC_ECHO_PATH, not
     # ENABLE_ECHO" as mixed. The selftest caught that; it is why the selftest exists.
     # What WAS wrong: the hand-listed four guesses. The model invented ECHO_CHAIN_ENABLED and
     # ECHO_CHAIN_STATE, matched nothing, and scored MISS. A closed answer space (one exact
     # variable name) must be matched by EXCLUSION, never by guessing the wrong names.
     # (?-i:) because grade() searches with re.I, under which [A-Z] also matches lowercase
     # prose such as fail_closed.
     r"(?:is|are|use|uses|set|named|variable)\s+`?(?-i:(?!EAC_ECHO_PATH\b)[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+)`?"
     # (?:=\\S*)? because the model also answers in VAR=value form (observed:
     # ECHO_AUDIT_CHAIN=true), which the bare end-anchor rejected.
     r"|^\s*`?(?-i:(?!EAC_ECHO_PATH\b)[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+)`?(?:=\S*)?\s*$",
     "EAC_ECHO_PATH"),

    ("chain_contents",
     "Does the Echo chain store the text of user prompts? Answer yes or no.",
     r"\bno\b|never|does not (store|contain)|hashes only|metadata only",
     r"^\s*yes\b|\byes,? it (does|stores)",
     "hashes and metadata only — a canary test asserts this"),
]

# 2026-08-26 onward the agent path returned this persona scaffold instead of an answer,
# because its configured model (ling-mini-2.0) is not served by its configured endpoint
# (localhost:8000 = ryzanstein-api, which offers only ryzanstein-bitnet-7b). The scaffold is
# confident, well-formed, and says nothing -- and it was being GRADED. A stub is not a wrong
# answer; it means the system under test never ran.
STUB = re.compile(
    r"(My approach follows these principles|"
    r"\d\.\s*DECOMPOSE\s*.{0,3}\s*Break problem into atomic components|"
    r"Let me analyze your request)", re.I | re.S)

ABSTAIN = re.compile(
    r"i (don'?t|do not) (have|know)|no (specific )?information|unable to (determine|find)|"
    r"not (enough|sufficient) (context|information)|cannot determine|unclear from|"
    r"would need (more|additional)|may refer to|if you (can )?provide", re.I)


def ask_agent(agent, q, timeout):
    body = json.dumps({"messages": [{"role": "user", "content": _fresh(q)}], "stream": False})
    p = subprocess.run(["curl", "-s", "-m", str(timeout), "-X", "POST", AGENT % agent,
                        "-H", "Content-Type: application/json",
                        "-H", "Cache-Control: no-store", "-d", body],
                       capture_output=True, text=True)
    try:
        return json.loads(p.stdout)["choices"][0]["message"]["content"]
    except Exception:
        return ""


def ask_gateway(q, timeout):
    body = json.dumps({"model": GATEWAY_MODEL, "max_tokens": 512,
                       "messages": [{"role": "user", "content": _fresh(q)}], "stream": False})
    p = subprocess.run(["curl", "-s", "-m", str(timeout), "-X", "POST", GATEWAY,
                        "-H", "Content-Type: application/json",
                        "-H", "Cache-Control: no-store", "-d", body],
                       capture_output=True, text=True)
    try:
        resp = json.loads(p.stdout)
    except Exception:
        return ""
    # A SUBSTITUTED MODEL STILL ANSWERS, AND THE ANSWER LOOKS FINE. The only way to notice is to
    # check the name that came back. Scoring a model you did not ask for is not a weaker result;
    # it is a different experiment wearing this one's label.
    served = resp.get("model", "")
    if served and served != GATEWAY_MODEL:
        raise SystemExit(
            "FATAL: asked the gateway for %r but it served %r. Every score below would describe "
            "the wrong model. Set EAC_EVAL_GATEWAY_MODEL, or fix the gateway routing."
            % (GATEWAY_MODEL, served))
    try:
        return resp["choices"][0]["message"]["content"]
    except Exception:
        return ""


def grade(ans, right, wrong):
    """CORRECT beats WRONG: an answer may mention other levels while asserting the right one.
    A pure abstention only counts when nothing was asserted either way."""
    if not ans.strip():
        return "error"
    if STUB.search(ans):
        # The backend answered with boilerplate. Grading it would invent a quality signal
        # out of an infrastructure failure -- which is exactly what happened here for the
        # eleven nights from 2026-08-26.
        return "stub"
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
        print("%-8s correct=%d  wrong=%d  mixed=%d  abstain=%d  miss=%d  error=%d  stub=%d   (n=%d)" % (
            p, t.get("correct", 0), t.get("wrong", 0), t.get("mixed", 0),
            t.get("abstain", 0), t.get("miss", 0), t.get("error", 0), t.get("stub", 0), n))
        if t.get("stub", 0):
            print("%-8s ^^ NOT A QUALITY RESULT: %d of %d answers were the persona scaffold, so"
                  % ("", t["stub"], n))
            print("%-8s    the model never ran. Check OLLAMA_URL/OLLAMA_MODEL on eac-backend"
                  % "")
            print("%-8s    and that the endpoint actually serves that model (/v1/models)." % "")
    print("\nanswers: %s" % a.out)
    print("NOTE: wrong != abstain. A system that declines is usable; one that fabricates")
    print("      cannot be told apart from one that knows.")
    print("      stub != wrong.  A stub means the system under test never ran at all; a score")
    print("      computed over stubs is not a measurement of anything.")


if __name__ == "__main__":
    main()
