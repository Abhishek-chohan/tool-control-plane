# ListAuditEvents: the audit trail becomes readable

Date: 2026-09-17

The durable audit trail (attribution added 2026-09-16) was write-only:
operators could not read it back without querying the database directly.
A new additive RPC closes that gap:

- **`SessionsService/ListAuditEvents`** (admin capability) lists the
  trail newest-first, paged via the standard `ListPage` trailer, with
  optional filters: `session_id`, `actor_key_id`, and `event` (exact
  match). Ordering is `(created_at DESC, id DESC)` — the id tiebreaker
  keeps same-tick events stable across pages.
- **Attribution semantics are visible in the data**: caller-driven
  events carry `actor_key_id`; system-driven events (retention sweeps,
  dead-lettering) and pre-attribution rows carry an empty value — the
  trail is append-only, so no backfill exists. `details` renders as a
  JSON object string per the API's JSON-as-string convention.
- **Additive per the compatibility policy**: new messages
  (`AuditEvent`, `ListAuditEventsRequest`/`Response`), no changes to
  existing RPCs. Generated stubs ship for Go, Python, and TypeScript;
  SDK_MAP documents the row.

Notes: the trail is best-effort durable by design (asynchronous,
bounded writes — see `server/docs/observability.md`), so the listing
may lag a completed operation by milliseconds and drops under
saturation rather than blocking requests. Retention (30 days) bounds
what the listing can return.
