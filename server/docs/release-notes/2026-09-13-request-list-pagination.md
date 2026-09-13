# 2026-09-13 — ListRequests pagination trailer

## Summary

`ListRequests` built its `ListPage` trailer and then dropped it: the
response carried only the request rows, so clients had no way to see
`next_page_token` or `total_size`. It also decided "is there another
page" with a page-length heuristic (`returned == page_size`), which
invents an extra empty page whenever the total is an exact multiple of
the page size.

## Changes

- The trailer is attached to every `ListRequestsResponse`.
- The emit condition is now `returned > 0 && offset+returned < total` —
  the exact form `ListUserSessions` already used — extracted into a
  shared `buildListPage` helper so both listings stay contract-equivalent.
- Conformance: new `request_list` fixture family
  (`request_list_pagination.json`) walked by both the Python and
  TypeScript harnesses over gRPC and HTTP. The conformance adapters
  gained page-envelope reads (`list_requests_page` / `listRequestsPage`)
  since the previous `list_requests` helpers flattened the page away.

## Tests

- Handler: 15 requests at page size 10 → page 1 has 10 rows, total 15,
  a token; page 2 has 5 rows, total 15, no token.
- Exact multiple: 10 requests at page size 10 → full page, no token.
- Conformance `request_list_pagination` passes over both transports in
  both SDK harnesses (Python 40 passed; TypeScript 46/46).

## Migration

- Clients that treated a missing trailer as "no pagination" now get real
  tokens; clients that guessed with `len(page) == page_size` can stop.
