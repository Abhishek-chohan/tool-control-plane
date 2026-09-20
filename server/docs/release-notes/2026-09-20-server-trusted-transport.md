# Core server trusted-transport declaration and served-transport metric

Date: 2026-09-20

The core gRPC server's production transport posture is now symmetric
with the gateways: TLS is required, and the one production-legal way to
serve plaintext is an explicit declaration that TLS terminates upstream.

- **`TOOLPLANE_SERVER_TRUSTED_TRANSPORT=1`** declares that a service
  mesh, sidecar, or terminating proxy terminates TLS in front of the
  server and the gRPC hop inside that boundary is intentionally
  plaintext. Without it, production boots without certificate and key
  files fail with an error naming both options — the TLS files
  (`--tls-cert-file`/`--tls-key-file`, or
  `TOOLPLANE_SERVER_TLS_CERT_FILE`/`TOOLPLANE_SERVER_TLS_KEY_FILE`) and
  the declaration. The declaration is environment-only, matching the
  gateways' `TOOLPLANE_TRUSTED_PROXY`, and development/test behavior is
  unchanged (plaintext remains the default outside production).
- **`toolplane_server_tls_enabled` gauge** on `/metrics`: `1` when the
  server terminates TLS itself, `0` when it serves plaintext. Serving
  plaintext is now visible to alerts instead of boot logs alone —
  pair it with `env="production"` in scrape-based alerting.
- The listening log line reports
  `transport=plaintext (upstream TLS terminator declared via TOOLPLANE_SERVER_TRUSTED_TRANSPORT)`
  when the declaration is what allowed a production plaintext boot.
- `toolplane doctor` enforces the same gate through `ValidateConfig`,
  so a mesh-terminated deployment validates without booting.

Notes: certificate and key files, when present, always win — the
declaration only relaxes the "no files in production" abort. The
gateways' own TLS gates are unchanged.
