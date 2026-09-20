# Reliability drills under load

Date: 2026-09-20

The load driver gained `--drill`, which injects the failure semantics
from the reliability drill matrix while the control plane is saturated
with agent-shaped traffic — the idle matrix proves the semantics;
drills under load prove they hold under contention.

- **`provider-kill` (the D1 semantics):** while agent-turn load runs,
  a provider dies *while holding a claimed request*. The orphaned claim
  must requeue after lease expiry and finish on a replacement machine:
  the drill registers the replacement (retrying until the dead owner's
  row is gone — tool ownership is exclusive), then asserts the
  store-backed reaper requeued the claim (Postgres), the held request
  reached terminal DONE, and nothing dead-lettered. Needs a ~45s+
  window: the 30s lease must expire first.
- **`drain-under-backlog` (the D4 semantics):** drain one provider
  mid-load. Asserts the drain RPC completes while saturated, reports
  drained within bounds, and nothing dead-lettered. The claim-blocking
  semantics themselves stay with the idle matrix's focused proofs.
- **`multi-instance-contention` (the D7 semantics):** consumers drive a
  second server replica sharing the Postgres store while providers
  serve from the first — the active-active shape under load. Asserts
  zero serialization-retry exhaustion. Skips cleanly without
  `TOOLPLANE_DATABASE_URL`.

Drill assertions ride in the JSON report (`drill.assertions`) and the
exit code reflects them. Fake providers also retry `UNAVAILABLE` calls
the way real provider SDKs do (three attempts, linear backoff), so
transient serialization contention under a requeue storm is absorbed
exactly as production absorbs it instead of failing the run.

Notes: `reliability-drills.md` gained a "Running Drills Under Load"
section mapping each drill to its matrix entry with the commands above.
