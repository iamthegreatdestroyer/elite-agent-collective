# Validate the GRADER before trusting it on real answers.
# A grader that cannot return every verdict is not a grader.
import importlib.util, sys
spec = importlib.util.spec_from_file_location("ev", r"eval.py")
ev = importlib.util.module_from_spec(spec); sys.argv=["x"]; spec.loader.exec_module(ev)

C = {c[0]: c for c in ev.CASES}
T = [
    # (case, synthetic answer, expected verdict)
    ("dispatch_sites", "Six.", "correct"),
    ("dispatch_sites", "It gates four dispatch sites.", "wrong"),
    ("dispatch_sites", "I don't have specific information about AgentOS.", "abstain"),
    ("dispatch_sites", "The ladder is a hierarchical control structure.", "miss"),
    ("tier4_ceiling", "L2", "correct"),
    ("tier4_ceiling", "A tier 4 agent is capped at L3.", "wrong"),
    ("tier4_ceiling", "Levels run L0 to L4; tier 4 is capped at L2.", "mixed"),
    ("artifact_holders", "None — the grant table is empty.", "correct"),
    ("artifact_holders", "FORGE holds artifact.write.", "wrong"),
    ("echo_hash", "It uses BLAKE2b-256.", "correct"),
    ("echo_hash", "It uses SHA-256.", "wrong"),
    ("failclosed_level", "fail-closed", "correct"),
    ("failclosed_level", "fail-open", "wrong"),
    ("approval_executes", "No — approval records a decision and executes nothing.", "correct"),
    ("chain_contents", "No, it stores hashes only.", "correct"),
    ("echo_env", "EAC_ECHO_PATH", "correct"),
    ("echo_env", "ENABLE_ECHO", "wrong"),
    ("dispatch_sites", "", "error"),
]
bad = 0
for cid, ans, want in T:
    _, _, right, wrong, _ = C[cid]
    got = ev.grade(ans, right, wrong)
    ok = got == want
    if not ok: bad += 1
    print("%-4s %-16s want=%-8s got=%-8s  %r" % ("ok" if ok else "FAIL", cid, want, got, ans[:44]))
print()
print("grader self-test: %d/%d  -> %s" % (len(T)-bad, len(T), "USABLE" if bad==0 else "BROKEN, do not run"))
