# The message-size ladder: a full replay window is readable everywhere

Date: 2026-09-16

The server retains an 8 MiB chunk window and accepts 16 MiB batch writes,
but everything downstream of it — both gateways and all three SDK
defaults — capped messages at 4 MiB. A provider that streamed a large
tool result had it stored correctly and then *unreadable*: `GetRequestChunks`
responses up to the window failed with ResourceExhausted at the gateway
dial and in every SDK, and the Python client wrapped the chunk read in
`try/except: pass`, so `streamResults` silently vanished from status
envelopes. One step higher, the comment claiming "a full max-size batch
fits" the server's own receive limit was wrong: the protobuf envelope
rides on top of the summed payload, so an exactly-full 16 MiB batch was
over the wire limit — rejected at the transport with no clear code.

- **Gateways** dial the backend with a `max-msg-size` default sized from
  the ladder (16 MiB batch + 1 MiB envelope headroom) instead of 4 MiB,
  in both `cmd/proxy` and `cmd/mcp-gateway`. Operators who pinned the
  flag keep their value.
- **SDKs** set explicit channel limits in the one options place per
  language: receive 9 MiB (window + headroom), send 17 MiB (batch +
  headroom) — Python `_channel_options`, TypeScript `channelOptions`,
  and the Go client's default call config.
- **Domain batch guard**: `AppendRequestChunks` now rejects a summed
  payload above `MaxChunkBatchPayloadBytes` (the transport ceiling minus
  64 KiB envelope headroom) with a clear `INVALID_ARGUMENT` naming the
  limit, instead of an opaque transport ResourceExhausted. Per-chunk
  bounds are unchanged.
- **Python no longer swallows chunk-read failures**: the status
  envelope's chunk enrichment logs a warning and attaches
  `streamResultsError` when the read fails, instead of silently dropping
  `streamResults`.
- **Conformance**: a new `request_recovery_full_window` fixture fills
  the entire 8 MiB window (16 × 512 KiB chunks) and reads it back in one
  `GetRequestChunks` response — proving the whole ladder (server send,
  gateway dial, client receive) end to end, on every transport's
  conformance run.

Notes: erratum for the 2026-09-11 release note's "a full max-size batch
fits" claim — it did not, for the envelope-margin reason above; that
batch boundary is now a documented, guarded limit.
