# Release Note — Chunk Storage Redesign (2026-09-11)

Branch: `feat/chunk-storage`

## What changed

### Append-only chunk table

Stream chunks previously lived in a JSONB array on the `requests` row: every
`AppendRequestChunks` call rewrote the whole row — including a JSONB array
that grows with the stream — inside the fenced serializable transaction. A
100-chunk stream meant 100 full-row rewrites of ever-growing payloads, and
the request row ballooned to hundreds of KiB for text streams.

Chunks now live in a dedicated append-only `request_chunks` table
(`request_id`, `seq`, `chunk`, `created_at`; primary key on request+seq,
`ON DELETE CASCADE` with the request):

- `AppendRequestChunksFenced` keeps its lease fence and sequence
  bookkeeping, but writes payloads as row inserts and updates only the
  request row's `stream_start_seq` / `next_stream_seq` / `updated_at` — no
  whole-row rewrite, no JSONB growth.
- Window reads go through the new `Storer.GetRequestChunksByRequest`;
  store-backed requests serve their chunk window from the table (including
  chunks appended by other replicas, which the local cache never mirrored).
- Trimmed window entries are deleted in the same transaction, so the table
  does not accumulate dead rows.
- The `requests.stream_results` JSONB column is frozen (legacy rows keep
  their data) and no longer read on the hot path.

### Byte caps

- `MaxRequestChunkBytes` (512 KiB): `AppendRequestChunks` rejects a chunk
  above the cap with `INVALID_ARGUMENT` before touching state.
- `maxRequestStreamWindowBytes` (8 MiB): the retained window is bounded by
  total payload as well as count (100); when the byte bound is exceeded,
  oldest chunks are trimmed and `StartSeq` advances, so resume semantics
  see the true window start.
- Together with the 32-chunks-per-RPC practical batch, worst-case memory
  per request window drops from ~400 MiB (the old 100 × 4 MiB) to ~8 MiB.

### Signal-driven streaming

`StreamExecuteTool` and `ResumeStream` polled a snapshot every 200 ms. They
now subscribe to the request's append signal and deliver chunks within
microseconds of the append; a 2-second fallback timer guards against a
missed broadcast.

### Signals-map leak fixed

The request signal map (`sync.Map` keyed by request ID) never deleted
entries: one leaked entry per request for the life of the process.
Terminal transitions (submit resolution/rejection, cancel, provider-driven
done/failure, and stream/resume handlers observing terminal) now release
the entry; late subscribers recreate it transparently.

### Server transport bounds

The gRPC server previously ran on implicit defaults. It now sets
`MaxRecvMsgSize`/`MaxSendMsgSize` to `MaxChunkBatchBytes` (32 × 512 KiB =
16 MiB, sized for a full max-size chunk batch per RPC) and a keepalive
enforcement policy (MinTime 10s, PermitWithoutStream).

## Compatibility

- Proto unchanged. `Request.stream_results` remains on the wire for
  compatibility but is no longer populated; consumers read chunk windows
  via `GetRequestChunks` / `ResumeStream` (the Python HTTP client's status
  call already enriched from the chunks endpoint, and the gRPC client now
  does the same, so poll-based fallbacks are unaffected).
- Postgres deployments run the `request_chunks` migration on startup;
  legacy `stream_results` data is left in place and simply unused.
- Client call patterns need no changes.

## Known limits

- Legacy rows' pre-migration chunks stay in `stream_results` and are not
  migrated into the table; their windows read as empty from the table.
  Streams are ephemeral, so this affects only requests alive across the
  upgrade.
- The 2-second fallback timer means a missed broadcast delays (does not
  lose) delivery.

## Verification

- `pkg/storage/chunks_test.go` (both stores): fenced append sequences,
  table-backed window read, second-append continuation, stale-epoch
  rejection, model-level byte-window trim.
- `pkg/service/chunk_storage_test.go`: cross-replica window visibility
  (replica B reads chunks it never mirrored), oversized-chunk rejection
  leaving state untouched, signal-map release on terminal, signal wake
  latency well under the old poll interval.
- Full server suite under `-race` (memory + Postgres): zero races;
  golangci-lint, vet, gofmt clean; Python lint + full conformance
  (grpc/http/mcp, multi-instance); TypeScript unit 37 + conformance 44/44.
