# Reference Deployment

This document is the maintained production source of truth for running, bootstrapping, upgrading, draining, rolling back, and validating Toolplane.

## Blessed Topology

The blessed production path is a split `Postgres + migrate job + gRPC server + HTTP gateway` deployment.

- The gRPC server is the primary control-plane runtime.
- The HTTP gateway is a separately rolled compatibility edge.
- Postgres is the required durable store.
- A one-shot migrate step runs before server rollout.
- The gateway dials the server over gRPC TLS using an explicit CA bundle.

The maintained artifacts for this topology are:

- `server/deploy/reference/compose.yaml`
- `server/deploy/reference/.env.example`
- `server/deploy/reference/certs/README.md`
- `server/docs/release-gate.md`

## Secondary Paths

- `server/entrypoint.sh` is a convenience combined-container bootstrap, not the maintained production topology.
- `server/run.sh` and `server/start.sh` are local development helpers, not the production source of truth.
- Any deployment shape that differs from the split server-plus-gateway path below is a site-specific adaptation.

## Reference Services

| Service | Role | Published port | Health signal |
| --- | --- | --- | --- |
| `postgres` | durable store for sessions, API keys, machines, requests, and tasks | `5432` by default | `pg_isready` |
| `migrate` | one-shot schema readiness step | none | exits successfully after startup migrations |
| `server` | primary gRPC control plane plus `/metrics` | `9001`, `9102` by default | `/metrics` on `:9102` |
| `gateway` | HTTP compatibility edge and `/health` surface | `8080` by default | `/health` on `:8080` |
| `bootstrap-server` | one-time fixed-auth bridge for minting the first Postgres-backed admin key | none | `/metrics` on `:9102` |
| `bootstrap-gateway` | one-time HTTP edge for the bootstrap flow | `18080` by default | `/health` on `:8080` |

## First Deploy

Run these commands from `server/deploy/reference`.

1. Prepare the env file and set the published ports, Postgres credentials, allowed origins, and one-time bootstrap token.

```bash
cp .env.example .env
```

1. Place `ca.crt`, `server.crt`, and `server.key` under `server/deploy/reference/certs/`.

The reference certificate flow is documented in `server/deploy/reference/certs/README.md`.

1. Build the image, start Postgres, and run the one-shot migrate job.

```bash
docker compose build
docker compose up -d postgres
docker compose run --rm migrate
```

1. Start the one-time bootstrap profile, mint the first Postgres-backed admin key, and record the returned secret immediately.

```bash
docker compose --profile bootstrap up -d bootstrap-server bootstrap-gateway

SESSION_JSON=$(curl -sS \
  "http://127.0.0.1:${TOOLPLANE_BOOTSTRAP_HTTP_PUBLISHED_PORT:-18080}/api/CreateSession" \
  -H "Authorization: Bearer ${TOOLPLANE_BOOTSTRAP_FIXED_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"userId":"bootstrap-admin","name":"bootstrap-admin","description":"one-time admin bootstrap","namespace":"ops"}')

SESSION_ID=$(printf '%s' "$SESSION_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["session"]["id"])')

ADMIN_KEY_JSON=$(curl -sS \
  "http://127.0.0.1:${TOOLPLANE_BOOTSTRAP_HTTP_PUBLISHED_PORT:-18080}/api/CreateApiKey" \
  -H "Authorization: Bearer ${TOOLPLANE_BOOTSTRAP_FIXED_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"sessionId\":\"${SESSION_ID}\",\"name\":\"reference-admin\",\"capabilities\":[\"read\",\"execute\",\"admin\"]}")

ADMIN_API_KEY=$(printf '%s' "$ADMIN_KEY_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["key"])')
printf '%s\n' "$ADMIN_KEY_JSON"
```

`CreateApiKey` is the only supported place that returns the secret. `ListApiKeys` will not return it later.

1. Stop and remove the bootstrap services after the admin key is captured.

```bash
docker compose stop bootstrap-server bootstrap-gateway
docker compose rm -f bootstrap-server bootstrap-gateway
```

1. Start the production server and gateway.

```bash
docker compose up -d server gateway
```

1. Validate the production stack.

```bash
curl -sS "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/health" | python3 -m json.tool

curl -sS "http://127.0.0.1:${TOOLPLANE_METRICS_PUBLISHED_PORT:-9102}/metrics" | grep '^toolplane_'

curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/GetSession" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"sessionId\":\"${SESSION_ID}\"}" | python3 -m json.tool
```

For an automated rehearsal of the same bootstrap and post-deploy validation flow, run `cd server && make reference-deployment-integration`.

## Upgrade Order

Use this order for upgrades to the reference stack:

1. Build or pull the new image.
2. Run `docker compose run --rm migrate` against the target Postgres database.
3. Roll the `server` service first and wait for its `/metrics` endpoint to answer.
4. Roll the `gateway` service second and wait for `/health` to return `status=ok`.
5. Re-run the post-deploy validation above.
6. Run the authoritative release gate locally against the same Postgres database before declaring the rollout complete.

The reference artifact is single-instance for clarity. If your platform runs multiple replicas, keep the same order: one migrate step, then server replicas, then gateway replicas.

## Provider Drain And Safe Handoff

Drain is part of the rollout contract, not an optional runtime feature.

1. Start replacement providers for the target session before draining the old ones.
2. Call `DrainMachine` for each machine you are retiring.
3. Wait until `toolplane_machine_draining` returns to zero, the old machine disappears from `ListMachines`, and the corresponding `machine_drain_completed` event appears in the session trace stream.
4. Only then terminate the old provider process or host.

Example drain call through the gateway:

```bash
curl -sS \
  "http://127.0.0.1:${TOOLPLANE_HTTP_PUBLISHED_PORT:-8080}/api/DrainMachine" \
  -H "Authorization: Bearer ${ADMIN_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"sessionId":"your-session-id","machineId":"your-machine-id"}'
```

Pending requests are not reassigned automatically during drain. Bring replacement providers online before drain so the pending queue has another owner path.

## Rollback Contract

The reference rollback contract is intentionally narrow.

- There is no down-migration tool. The rollout path assumes additive, idempotent startup migrations.
- If a release problem is isolated to the HTTP compatibility layer or proxy throttling, roll back `gateway` first.
- If the problem is in the control-plane runtime, roll back `server` only to a version that is already known to operate against the migrated schema.
- After any rollback, re-run the same `/health`, `/metrics`, and authenticated session checks used after rollout.

## Release Gate Relationship

`server/docs/release-gate.md` is the authoritative runtime validation path for this deployment story.

- The gate validates the same split server-plus-gateway runtime shape and the same Postgres-backed storage contract.
- The gate intentionally does not validate one-time TLS certificate provisioning or the bootstrap fixed-auth bridge used to mint the first Postgres-backed admin key.
- Use `make reference-deployment-integration` when you want the repository to rehearse the reference TLS certificate wiring and one-time bootstrap bridge directly.
- Use the reference deployment checks above for stack bring-up, then use `make release-gate` to validate the maintained runtime slice against the same database.

Local release-gate reproduction against the reference database from the repository root:

```bash
cd server && TOOLPLANE_DATABASE_URL=postgres://toolplane:toolplane@localhost:5432/toolplane?sslmode=disable make release-gate
```

## Client Boundary

The reference deployment always secures the gateway's hop to the server with gRPC TLS.

- HTTP clients use the gateway on `:8080`.
- Direct gRPC clients that connect to `:9001` must trust the same CA bundle used by the gateway.
- The reference stack publishes `:9001` because the canonical API is still gRPC-first, but the operator source of truth for this initiative is the split control-plane deployment, not every client transport configuration detail.
