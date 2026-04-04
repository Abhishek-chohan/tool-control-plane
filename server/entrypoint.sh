#!/bin/sh

# Start the gRPC server in the background
echo "Starting gRPC server on :9001..."
/app/server --trace-sessions &

# Start the gateway proxy in the foreground
# It will connect to the gRPC server started above
echo "Starting gateway proxy on :8080..."
/app/toolplane-gateway --listen :8080 --backend localhost:9001 

# Keep the container running
wait -n
