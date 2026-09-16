# Machine re-register: reconcile in place, never mint twice

Date: 2026-09-16

Re-registering a machine (SDK reconnect, redeploy) used to tear its tool
set down before re-registering it: every tool row was hard-deleted, then
N ownership claims ran — all under the machine registry lock. Three
defects fell out of that shape, and a fourth lived in the identity path.

- **No delete-and-recreate gap**: the tool set now converges by
  reconciliation. Kept names upsert in place (tool IDs preserved — the
  upsert's same-owner branch from the RegisterTool contract change),
  names dropped from the new list are removed, and a name held by
  another live machine keeps its current owner (the claim conflicts and
  is logged, never deleted). Consumers no longer observe NOT_FOUND for a
  tool mid-re-register, and requests keyed by tool name stay valid
  across reconnects.
- **Tool IDs are stable**: the old delete-first flow forced the INSERT
  branch on every re-register, minting a fresh tool ID each time. Stable
  IDs keep request history, traces, and client-side references coherent
  across reconnects.
- **The registry lock no longer spans tool IO**: the machine lock guards
  identity and the registry only; the reconcile runs after it drops (the
  tool registry has its own lock). Heartbeats and machine lookups are no
  longer queued behind N tool transactions.
- **A cold replica cannot mint a second credential**: a same-ID
  re-register on a replica whose cache missed now reads the durable
  machine row and verifies the presented credential against it, instead
  of creating a fresh machine whose full upsert silently rotated the
  stored token hash and reset `created_at`. `SaveMachine` itself became
  first-registration-wins on identity fields (`token_hash` binds once —
  `BindMachineToken`'s compare-and-set is the only writer after that —
  and `created_at` never moves), on both backends, so no write path can
  clobber a credential any more. The legacy pre-token bind path also
  goes through the store's CAS, so two racing registrations cannot bind
  different credentials.

Notes: drain behavior is unchanged — draining still removes the
machine's tools before unregistering (that teardown is the point of
drain). The reconcile diff is store-first on store-backed replicas, so
it sees registrations other replicas made.
