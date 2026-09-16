# RegisterTool error contract: one sentinel, real codes

Date: 2026-09-16

Tool registration is an idempotent upsert-with-takeover, and its error
surface now says so.

- **The only registration conflict is ownership**: same-machine
  re-register and stale-owner takeover both update in place, preserve the
  tool ID, and return `OK`. A name held by another machine whose
  heartbeat is fresh is a *state conflict* and now surfaces as
  `FAILED_PRECONDITION` (with the owning machine named in the message) —
  not `ALREADY_EXISTS`, which the taxonomy reserves for create
  collisions.
- **The catch-all is gone**: previously every failure from
  `RegisterTool` — Postgres outage, persistence deadline, plain
  "machine not registered" — was mislabeled `ALREADY_EXISTS`. The
  handler now funnels through the standard taxonomy: infrastructure
  failures fail closed to `INTERNAL`, and a persistence deadline that
  fired inside the handler keeps `DEADLINE_EXCEEDED` (new taxonomy arm).
- **One sentinel on both backends**: the cache-only registry path and
  the store path now wrap the same `storage.ErrToolOwnershipConflict`
  (the cache path previously wrapped `ErrAlreadyExists`), so the machine
  re-register loop's conflict check matches on both storage modes. Two
  memory-store divergences from the Postgres claim were fixed in
  passing: a same-machine re-register no longer conflicts, and a prior
  owner missing from the registry is treated as takeover-able (both
  already were the Postgres behavior).
- **Python reconnect no longer swallows conflicts**: the reconnect walk
  string-matched `ALREADY_EXISTS` and silently continued — which, under
  the old catch-all, also hid infrastructure failures. Every
  registration failure now logs as a warning. Visible behavior change:
  a second live provider process registering a taken name errors loudly
  where it silently continued before — that is the point.
- TS/Go SDKs need no code change: their error classes are keyed by gRPC
  code, so the conflict remaps to the failed-precondition class
  automatically.

Notes: wire-shape compatible per the compatibility policy (behavior
tightening with a release note; gRPC drops the old handler's
response-plus-error dual return, so nothing on the wire depended on it).
Conformance: `tool_discovery.json` now pins same-machine re-register
keeping the tool ID and the rival-machine conflict code across gRPC,
HTTP, and MCP transports.
