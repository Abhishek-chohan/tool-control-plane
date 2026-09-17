# The toolplane unified command

Date: 2026-09-17

A new `toolplane` binary unifies the control plane and its client verbs
behind one entry point:

- **`toolplane serve`** runs the gRPC control plane with the same flags
  (`--port`, `--metrics-listen`, `--migrate-only`, `--tls-cert-file`,
  `--tls-key-file`, `--trace-sessions`), the same `TOOLPLANE_*`
  environment contract, and the same dev-posture banner and production
  gates as the standalone `toolplane-server`. Either entry point can be
  swapped without re-learning configuration; `--version` reports the
  build.
- **Exit-code contract**: client verbs exit with a stable code mapped
  from the failure's gRPC status — 2 invalid input, 3 not found,
  4 permission denied, 5 failed precondition, 6 resource exhausted,
  7 unavailable, 8 deadline exceeded, 9 cancelled, 1 anything else —
  so scripts branch on semantics without parsing output. The mapping is
  pinned by tests.
- **Completions**: `toolplane completion bash|zsh|fish|powershell`
  generates shell completion from the command tree.

The standalone binaries are unchanged and remain the reference
deployment's entry points. Internally the server lifecycle moved to an
importable package (`internal/server`) with signal handling separated
from the run loop, which is what lets the CLI embed it and lets tests
boot a real server in-process; `cmd/server/auth` moved to
`internal/auth` to match (import path change only — no behavior change).

Notes: new Go dependencies `github.com/spf13/cobra` (command tree,
generated completions, consistent help across the growing verb surface)
and its transitive `pflag`. Client verbs (`invoke`, session/machine/key
administration, and the rest) land in subsequent changes; `serve` is the
foundation they register against.
