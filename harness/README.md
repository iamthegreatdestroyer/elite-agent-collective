# harness — tune the harness, not the model

A small, dependency-free (stdlib-only) toolkit that applies the LangChain /
NVIDIA *"Tuning the harness, not the model: a Nemotron 3 Ultra playbook"*
methodology to the Sigma ecosystem's own small, self-hosted models. Every model
call goes through the Ryzanstein gateway (`EAC_GATEWAY_URL`, default
`http://localhost:8000`). No OpenAI/Meta; CPU-only.

The premise the ecosystem is built on — *small local model + a good harness ≈
frontier behavior at a fraction of the cost* — is exactly the playbook's thesis.
This makes that pursuit **measured** instead of intuition-driven.

## The five pieces

| file | purpose |
|---|---|
| `eval/tasks.py` | per-capability tasks (tool-calling, retrieval/grounding, multi-turn, summarization, instruction-following) with **deterministic** checkers (no LLM-as-judge) |
| `eval/runner.py` | scores models per-capability through the gateway; model-outer loop (each model loads once); writes a JSON + markdown scoreboard |
| `middleware.py` | **tool-call validation + one-shot correction** re-prompt; `continuation_notice` (ReadFileContinuationNotice pattern) |
| `profiles.py` | a per-model harness **profile**: point-of-need vs system inject-mode + single-purpose prompt suffix + the middleware. Loads its editable knobs from `profile_tuned.json` |
| `tune.py` | the **ralph loop**: evaluate → propose one single-purpose block → re-evaluate → keep only on improvement-without-regression. Writes ONLY `profile_tuned.json`; a held-out task split validates generalization |
| `selftest.py` | fast no-network checks for the deterministic bits |

## Run it

```bash
# baseline per-capability scoreboard
python3 -m harness.eval.runner --models gemma3:4b-it-qat,phi4-mini,granite4:1b

# with the tuned harness profile (middleware + inject-mode + learned blocks)
python3 -m harness.eval.runner --models granite4:1b --profile tuned

# automated tuner (scoped write to profile_tuned.json only)
python3 -m harness.tune --model granite4:1b

# logic checks, no models
python3 -m harness.selftest
```

## Baseline result (2026-07-11, corrected checker)

| model | tool-calling | retrieval | multi-turn | summ | instr | overall |
|---|---|---|---|---|---|---|
| **phi4-mini** | 3/3 | 3/3 | 2/2 | 2/2 | 2/2 | **12/12** |
| gemma3:4b-it-qat | 2/3 | 3/3 | 2/2 | 2/2 | 2/2 | 11/12 |
| granite4:1b | 1/3 | 3/3 | 2/2 | 2/2 | 2/2 | 10/12 |

Findings, per-capability (the playbook's point — a per-capability score tells a
harness problem from a model one):
- **phi4-mini is the strongest small all-rounder** — worth considering beyond
  its current content-farm role.
- **granite** picks the right tool but flattens args
  (`{"tool":"read_file","path":...}` instead of nesting under `args`) — a
  harness-fixable schema issue.
- The first run also caught a bug in *this suite's* checker (a curly apostrophe
  in "don't" wasn't matched) — fixed; it's why you measure.

## Harness result (ralph loop on granite4:1b)

The tuned profile's **tool-call correction middleware** detects a missing `args`
object and re-prompts once. Result, no weights changed:

- granite tool-calling **1/3 → 3/3**, overall **10/12 → 12/12**.
- Held-out `tool_readfile` (never seen during tuning) went **fail → pass** —
  the fix **generalized**. The tuner converged with an empty prompt suffix
  because the middleware alone sufficed (an honest outcome — not every gap needs
  a prompt block).

## Env knobs
`EAC_GATEWAY_URL` (gateway base). The EAC retriever also honors
`EAC_RETRIEVER_INJECT=message|system` (default `message`, point-of-need).
