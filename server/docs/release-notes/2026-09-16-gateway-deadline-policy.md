# The gateways own their deadline policy

Date: 2026-09-16

The HTTP gateway's unary deadline was one blanket 30-second default with
two holes: it silently truncated server-side waits (the fired deadline
converted to a normal in-flight 200 — no error anywhere), and any HTTP
caller could override it in either direction via the `Grpc-Timeout`
header, which grpc-gateway converts into the context deadline before the
dial interceptor runs.

- **Method-scoped deadlines**: the two wait entrypoints
  (`ToolService/InvokeTool`, `ToolService/ExecuteTool`) now dial with a
  wait backstop of 1h 30m — above the server's 3600s wait ceiling (which
  the server itself rejects beyond, per the wait-timeout change), so an
  honest max-length wait completes while a server bug still cannot hang
  the gateway connection. Every other unary keeps the 30-second
  fast-call default. The in-flight-200-on-expiry contract is unchanged.
- **`Grpc-Timeout` is stripped at the edge**: the proxy's root middleware
  removes the header before the gateway mux parses it, so the deadline
  policy is gateway-owned. HTTP callers bound themselves with their own
  client timeouts; an existing deadline still wins over the per-method
  default (the policy is a default, not an override). Stripping also
  protects streaming replays, which the header could otherwise kill
  mid-flight.
- **The MCP gateway mirrors the scoping** (same per-method policy in its
  backend dial). It has no grpc-gateway mux, so no header override exists
  there to strip.
- **Breaker interaction**: deliberate max-length waits no longer surface
  as 504s (they complete under the backstop), so they cannot trip the
  gateway-wide circuit breaker; legitimately slow fast-calls still fail
  at 30s and trip it as intended.

Notes: per-ingress numbers stay intentionally different and are now
documented policy — direct gRPC callers are bounded by the server's wait
ceiling, HTTP waits by the 1h30m backstop with 30s fast-calls, and the
MCP facade by its `--sync-timeout`. A single uniform cross-ingress
policy remains a non-goal.
