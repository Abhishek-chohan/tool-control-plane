#!/bin/sh

# Convenience combined-container bootstrap.
# The maintained production topology is the split server-plus-gateway path in
# server/docs/reference-deployment.md.

: "${TOOLPLANE_METRICS_LISTEN:=:9102}"

# Start the gRPC server in the background
echo "Starting gRPC server on :9001 with metrics on ${TOOLPLANE_METRICS_LISTEN}..."
/app/server --trace-sessions --metrics-listen "${TOOLPLANE_METRICS_LISTEN}" &

# Start the gateway proxy in the foreground
# It will connect to the gRPC server started above
echo "Starting gateway proxy on :8080..."
/app/toolplane-gateway --listen :8080 --backend localhost:9001 

# Keep the container running
wait -n
