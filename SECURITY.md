# Security Policy

## Reporting a vulnerability

Do not open public issues or pull requests for security problems. Use GitHub's private vulnerability reporting for this repository (Security tab → "Report a vulnerability"), which reaches the maintainers directly.

## Supported configuration

Only production mode is supported for deployments that handle untrusted traffic:

- `TOOLPLANE_STORAGE_MODE=postgres` and `TOOLPLANE_AUTH_MODE=postgres` (the server refuses memory storage and non-Postgres auth in production mode)
- gRPC TLS on the server (`TOOLPLANE_SERVER_TLS_CERT_FILE` / `TOOLPLANE_SERVER_TLS_KEY_FILE`)
- Explicit `TOOLPLANE_PROXY_ALLOWED_ORIGINS` on the HTTP gateway, and TLS on any gateway-to-backend hop that crosses a network

The development defaults (memory storage, a fixed shared API key, no TLS) are for local work and CI only.

## Notes

- API keys carry explicit capabilities (`read`/`invoke`/`provide`/`admin`) and are bound to their session; mint them with least privilege.
- The HTTP and MCP gateways forward caller credentials to the backend unchanged and enforce per-key and per-IP rate limits.
- See `server/docs/release-notes/2026-09-05-security-posture.md` for the security-posture change history.
