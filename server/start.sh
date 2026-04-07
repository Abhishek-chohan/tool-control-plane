#!/bin/bash
set -euo pipefail

# Add Go bin to PATH
export PATH=$PATH:/home/runner/go/bin

GRPC_PORT="${TOOLPLANE_GRPC_PORT:-9001}"
HTTP_PORT="${ToolplaneHTTP_PORT:-8080}"
METRICS_LISTEN="${TOOLPLANE_METRICS_LISTEN:-127.0.0.1:9102}"

# Start the gRPC server in the background
echo "Starting gRPC server on localhost:${GRPC_PORT} with metrics on ${METRICS_LISTEN}..."
./bin/server --port "${GRPC_PORT}" --metrics-listen "${METRICS_LISTEN}" &
SERVER_PID=$!

# Wait a bit for server to start
sleep 2

# Start the gateway proxy in the foreground
echo "Starting HTTP gateway on localhost:${HTTP_PORT}..."
./bin/toolplane-gateway --listen "localhost:${HTTP_PORT}" --backend "localhost:${GRPC_PORT}"

# Cleanup: If gateway exits, kill the server too
kill $SERVER_PID 2>/dev/null
