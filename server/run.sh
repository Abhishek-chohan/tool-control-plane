#!/usr/bin/env bash
set -euo pipefail

# Local development bootstrap.
# Builds the server/gateway binaries and launches the stack with explicit
# env-driven defaults. Override any variable before running this script.
# See server/docs/local-development.md for the full env contract.

cd "$(dirname "$0")"

: "${TOOLPLANE_ENV_MODE:=development}"
: "${TOOLPLANE_AUTH_MODE:=fixed}"
: "${TOOLPLANE_AUTH_FIXED_API_KEY:=toolplane-conformance-fixture-key}"
: "${TOOLPLANE_STORAGE_MODE:=memory}"
: "${TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND:=1}"
: "${TOOLPLANE_PROXY_ALLOWED_ORIGINS:=http://localhost:3000,http://127.0.0.1:3000,http://localhost:8080}"
: "${TOOLPLANE_GRPC_PORT:=9001}"
: "${ToolplaneHTTP_PORT:=8080}"

export TOOLPLANE_ENV_MODE
export TOOLPLANE_AUTH_MODE
export TOOLPLANE_AUTH_FIXED_API_KEY
export TOOLPLANE_STORAGE_MODE
export TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND
export TOOLPLANE_PROXY_ALLOWED_ORIGINS
export TOOLPLANE_GRPC_PORT
export ToolplaneHTTP_PORT

echo "Building server and gateway..."
mkdir -p bin
go build -v -o bin/server ./cmd/server
go build -v -o bin/toolplane-gateway ./cmd/proxy

echo "Launching Toolplane local development stack with:"
echo "  TOOLPLANE_ENV_MODE=${TOOLPLANE_ENV_MODE}"
echo "  TOOLPLANE_AUTH_MODE=${TOOLPLANE_AUTH_MODE}"
echo "  TOOLPLANE_STORAGE_MODE=${TOOLPLANE_STORAGE_MODE}"
echo "  TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND=${TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND}"
echo "  TOOLPLANE_PROXY_ALLOWED_ORIGINS=${TOOLPLANE_PROXY_ALLOWED_ORIGINS}"
echo "  TOOLPLANE_GRPC_PORT=${TOOLPLANE_GRPC_PORT}"
echo "  ToolplaneHTTP_PORT=${ToolplaneHTTP_PORT}"
echo
echo "Use server/.env.example and server/docs/local-development.md to customize this bootstrap."

./start.sh