# Same-tick listings are deterministic; pagination comments tell the truth

Date: 2026-09-16

`ListRequests` sorted by `CreatedAt` alone with an unstable sort over map
iteration — requests created in the same clock tick (common: batch
submission under a provider loop) could reshuffle between pages and between
identical calls. The store's `ORDER BY created_at` had the same gap: rows
with equal timestamps came back in whatever order the scan produced.

- **Deterministic order**: the store orders by `(created_at, id)` and the
  service sorts with the same tiebreak (`sort.SliceStable`), matching the
  shape `ListUserSessions` already had. Pages are now reproducible for
  same-tick creations on both storage modes.
- **Honest pagination comments**: the v1 token is a reversible
  base64-wrapped offset, not "deliberately non-enumerable" — it is
  trivially decodable and carries no authority (a forged token at worst
  yields an empty or repeated page). And the offset addresses the *live*
  ordering, not a snapshot: rows that appear or leave the filtered set
  mid-walk shift positions, so pages may skip or repeat under mutation.
  The false claims in `page_token.go` and both listing handlers now state
  this contract.

Notes: snapshot/isolated pagination (status+ID cursors) remains v2
territory — the offset tradeoff is now documented instead of denied. The
unpaginated listing surfaces (`ListSessions`/`ListTasks`/`ListTools`)
are unbounded by design and stay a separate follow-up.
