# Local Development Bootstrap

This document describes the supported local-development startup path.

For the maintained production topology, bootstrap flow, migration order, drain behavior, rollback contract, and validation path, use `server/docs/reference-deployment.md`.

## Goal

Use one explicit development contract for auth, storage, and proxy behavior instead of relying on embedded secrets or silent fallbacks.

## Supported Development Mode

The supported local path uses:

- `TOOLPLANE_ENV_MODE=development`
- `TOOLPLANE_AUTH_MODE=fixed`
- `TOOLPLANE_AUTH_FIXED_API_KEY=toolplane-conformance-fixture-key`
- `TOOLPLANE_STORAGE_MODE=memory`
- `TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND=1`
- `TOOLPLANE_METRICS_LISTEN=127.0.0.1:9102`

The fixed API key above is an intentionally non-secret local fixture value. It is suitable for local development and CI only.

## Quick Start

From `server/` in a bash shell:

```bash
set -a
source .env.example
set +a

./run.sh
```

That path builds the server and gateway, exports the explicit development settings, and starts the local stack through `start.sh`.

The supported local bootstrap keeps metrics on loopback at `127.0.0.1:9102` so Prometheus-style scraping uses a stable address without exposing the endpoint beyond the host.

## Production Reference Pointer

If you want a production-oriented local stack, do not extend `run.sh` or `start.sh` into a pseudo-production shape.

Use `server/docs/reference-deployment.md` and `server/deploy/reference/compose.yaml` instead. That path is the maintained operator contract for:

- split server and gateway processes
- Postgres-backed storage and auth
- explicit migration ordering
- gRPC TLS on the server-to-gateway hop
- one-time API-key bootstrap
- rollout, drain, rollback, and validation

## Notes

- `TOOLPLANE_STORAGE_MODE=memory` is an explicit development-mode choice, not a silent fallback.
- `TOOLPLANE_ENV_MODE=production` now requires Postgres-backed storage. In-memory mode remains development or test only.
- `TOOLPLANE_AUTH_MODE=postgres` validates against the server-managed `api_keys` records already loaded into the session service from Postgres-backed storage.
- Postgres-backed API keys are verified by stored hash, carry explicit `read` or `execute` or `admin` capabilities, and remain session-scoped credentials.
- The HTTP gateway remains a compatibility surface over the canonical gRPC contract.
- `server/entrypoint.sh` remains a convenience combined-container bootstrap, not the maintained production topology.
- If you use the Python or TypeScript conformance harnesses with auto-boot enabled, they now set the same development-mode env contract automatically.
