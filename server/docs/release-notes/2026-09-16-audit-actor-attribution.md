# Audit events name the acting key; the dead sessions.api_key column is gone

Date: 2026-09-16

The durable audit trail recorded *what* happened but never *who* did it:
`audit_events` had no actor column, `CreateApiKey` attributed creation to
the session owner instead of the authenticated principal (an admin-for-the-
session key minting keys showed up as the owner's doing), and the
session-wide kill switch (`InvalidateSession`) — the one trail you read
during an incident — recorded no actor at all.

- **`actor_key_id` column** (nullable): caller-driven audited events —
  API-key creation and revocation, session deletion and bulk deletion, and
  the kill switch — now name the acting API key. System-driven events
  (retention sweeps, dead-lettering) carry the empty value, as do rows that
  predate attribution: the trail is append-only, so no backfill exists or
  is possible. Machine events continue to attribute through `machine_id`.
- **`CreateApiKey` attribution fixed**: the key's `created_by` and the
  audit actor come from the authenticated principal (user identity when
  present, the key id otherwise) — not the session owner.
- **The kill switch is attributable**: `InvalidateSession` records the
  admin key that pulled the switch.
- **The audit contract is documented** in `server/docs/observability.md`:
  what is recorded, actor semantics, delivery guarantees (asynchronous,
  bounded, best-effort durable), and retention.
- **`sessions.api_key` dropped**: the legacy session-lock column was
  written (always empty) by every session upsert and read by nothing; the
  startup retirement drain and the model field went with it. Pre-1.0 with
  no tagged releases, the compatibility policy pins the wire, not internal
  storage shape.

Notes: the actor rides the existing trace pipeline (metadata at the record
site, graduated to the column by the audit recorder), so attribution costs
nothing on the request path and survives the async write queue.
