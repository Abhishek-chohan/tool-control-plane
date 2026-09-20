# Error taxonomy completion: every handler funnels through one translation

Date: 2026-09-20

The last handler-local `status.Errorf` calls in the gRPC adapter are
gone: list, session, machine, and task handlers now route every failure
through `statusFromDomainError` like the rest of the surface, so a
condition surfaces with the same code regardless of which RPC produced
it, and unmapped failures fail closed to INTERNAL by design rather than
by happenstance.

Wire-visible changes, all narrow:

- **`ListAuditEvents` page-encode failures move INVALID_ARGUMENT →
  INTERNAL.** Encoding a response page is a server-side operation; the
  other list handlers already reported it as INTERNAL, and caller-input
  codes are reserved for caller input. Malformed *request* tokens
  remain INVALID_ARGUMENT everywhere (the page-token codec now carries
  the `ErrInvalidArgument` sentinel, so all three decode sites get it
  from one place).
- **Cancellation keeps its code.** `errors.Is(err, context.Canceled)`
  now maps to CANCELED like `context.DeadlineExceeded` already mapped
  to DEADLINE_EXCEEDED — previously both surfaced as INTERNAL from
  handlers that passed raw store/service errors through. In practice:
  `DrainMachine` aborted by its context reports CANCELED/DEADLINE_EXCEEDED
  instead of INTERNAL, and any list handler sitting on a store call when
  a caller hangs up reports CANCELED.
- **`CreateSession` collisions** keep bare ALREADY_EXISTS with no
  payload (the arm attaches nothing; the existence-oracle guard is
  unchanged) — the message now reads
  `failed to create session: already exists: …`.
- Ordinary list/stat/invalidate failures keep INTERNAL and keep their
  `failed to <action>: …` messages; the change is the route, not the
  wire.

Notes: the CLI exit-code contract picks up the code changes above
automatically (deadline → 8, canceled → 9, instead of the generic 1);
SDK typed-error classes follow their code tables the same way. The
`internal/auth` interceptors remain intentionally outside the taxonomy:
they deny before handlers run and carry per-method policy context.
