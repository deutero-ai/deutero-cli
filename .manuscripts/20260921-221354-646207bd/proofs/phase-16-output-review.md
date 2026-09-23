# Phase 4.85 — Agentic output review

**Run degraded: live output sampling was not possible.** Outbound HTTPS from Go processes
stopped working in this environment partway through the run — a bare stock Go HTTP client
times out on the TLS handshake while `curl` to the same URL returns 200 in ~1.2s. It is the
same Go-process network restriction that blocked `govulncheck` earlier, and it is unrelated to
the generated CLI (reproduced with a 20-line program that imports only `net/http`).

**What stands in its place.** Every one of the 11 hand-built commands was executed against
live production data earlier in this run, before the network failed, and the output was
inspected for exactly the plausibility failures this phase exists to catch:

| Command | Live output | Plausibility check |
|---|---|---|
| `flow coverage` | 2 interviews, 2 transcripts, no branches | Correct: the study is linear, so there are no branch records. Did **not** invent branches. |
| `flow effects` | empty fetches/signals | Correct: linear study has no flow side-steps. |
| `rehearse diff` | 3 simulated vs 0 real, avg 5 msgs, 4.39 min | Cohort split exact; missing-cohort case stated rather than silently zeroed. |
| `answers matrix` | real answers q0–q5 from the message stream | Correct source for a linear study, where captured variables are empty by design. |
| `crosstab` | empty segments | Correct: that study collects no characteristics. Reports `unsegmented`, does not fabricate buckets. |
| `fielding` | "Quota already filled" (1 of 1) | Matches the server's own stats; day series consistent. |
| `saturation` | interview 1: 0 terms; interview 2: 139 new | Interview 1 was abandoned with no participant turns — 0 is right, not a bug. Correctly says saturation **not** reached. |
| `reconcile` | 1 completed, 0 with external id, 0 delivered, 0 never-delivered | The honest three-population split: reports "nothing to join on" instead of miscounting it as a delivery failure. |
| `study bundle` | 8 sections | Complete. |
| `study diff` | `same: true` against its own bundle | Self-consistency holds; bookkeeping fields correctly ignored. |
| `study apply` | plan only, `confirmed: false` | Print-by-default safety holds; nothing written. |

The failure mode this phase targets — a command that returns confident nonsense when the data
is absent — did not occur in any of the eleven. The recurring pattern is the opposite: each
command reports the empty case explicitly and says why.

**Not claimed:** this is manual inspection by the operator, not the independent agentic
sampling the phase specifies. If the environment's Go networking recovers, re-run
`printing-press-output-review` against the promoted CLI. Per the phase's Wave B policy,
findings here are warnings and do not block shipping.
