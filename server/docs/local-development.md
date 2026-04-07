# Local Development Bootstrap

This document describes the supported local-development startup path.

## Goal

Use one explicit development contract for auth, storage, and proxy behavior instead of relying on embedded secrets or silent fallbacks.

## Supported Development Mode

The supported local path uses:

- `TOOLPLANE_ENV_MODE=development`
- `TOOLPLANE_AUTH_MODE=fixed`
- `TOOLPLANE_AUTH_FIXED_API_KEY=toolplane-conformance-fixture-key`
- `TOOLPLANE_STORAGE_MODE=memory`
- `TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND=1`

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

## Production-Like Startup

For production-oriented testing, replace the development auth and storage settings explicitly.

For local Postgres-backed API key auth:

```bash
export TOOLPLANE_ENV_MODE=production
export TOOLPLANE_AUTH_MODE=postgres
export TOOLPLANE_STORAGE_MODE=postgres
export TOOLPLANE_DATABASE_URL=postgres://username:password@localhost:5432/toolplane?sslmode=disable
export TOOLPLANE_PROXY_ALLOWED_ORIGINS=https://your-app.example.com
export TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND=0
```

If the database does not have any API keys yet, bootstrap one against the same Postgres database first. The simplest path is to start once in fixed auth mode, call `SessionsService.CreateApiKey`, record the returned secret immediately, then restart with `TOOLPLANE_AUTH_MODE=postgres`. `ListApiKeys` is metadata-only and will not return the full secret later.

The server now fails fast when required auth or storage configuration is missing, and the proxy rejects insecure backend dialing and wildcard origins in production mode.

For production-oriented proxy policy controls, start the gateway with explicit rate limits rather than leaving the defaults disabled. Example:

```bash
./bin/toolplane-gateway --listen :8080 --backend localhost:9001 --api-rate 25 --api-burst 50 --ip-rate 100 --ip-burst 200
```

The gateway's `/health` endpoint reports circuit-breaker and throttle counters so operators can confirm those controls are active without reproducing failures locally.

## Notes

- `TOOLPLANE_STORAGE_MODE=memory` is an explicit development-mode choice, not a silent fallback.
- `TOOLPLANE_ENV_MODE=production` now requires Postgres-backed storage. In-memory mode remains development or test only.
- `TOOLPLANE_AUTH_MODE=postgres` validates against the server-managed `api_keys` records already loaded into the session service from Postgres-backed storage.
- Postgres-backed API keys are verified by stored hash, carry explicit `read` or `execute` or `admin` capabilities, and remain session-scoped credentials.
- The HTTP gateway remains a compatibility surface over the canonical gRPC contract.
- If you use the Python or TypeScript conformance harnesses with auto-boot enabled, they now set the same development-mode env contract automatically.
