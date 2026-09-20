# Soak mode and the nightly load workflow

Date: 2026-09-20

The load driver gained a soak mode and the repository its first
scheduled load job — capacity evidence that runs continuously instead
of on request.

- **`loadgen --soak`** samples the server's own runtime gauges
  (`go_goroutines`, `process_resident_memory_bytes`) every 15 seconds
  while the configured shape runs, then — after the fleet has drained —
  asserts bounded growth: goroutines within a fixed slack of the
  run-window floor, memory within 10% + 32MiB of the run-window peak.
  The thresholds are leak detectors, not precision budgets; assertions
  ride the report like drill assertions and the exit code reflects
  them. `make soak` (default 30 minutes, `SOAK_DURATION=...` to vary)
  wraps it and writes `bin/soak-report.json`.
- **Nightly `Load` workflow** (`.github/workflows/load.yml`): 03:30
  UTC against a fresh Postgres service container, 45-minute job
  timeout, `workflow_dispatch` inputs for duration and shape, the
  repository's first `concurrency:` group (an in-flight run is
  cancelled by a newer one), and the report uploaded as an artifact on
  every outcome (14-day retention — the report is the deliverable).
  Never a per-PR gate: soak jobs are flake-shaped by nature, and the
  per-PR signal stays with the correctness suites.

Notes: soak samples are taken from the server's `/metrics` endpoint, so
external-mode soaks work against any server that exposes it
(`--metrics-url`). A soak cannot ride a drill — drills mutate the
fleet, which would smear the growth baselines.
