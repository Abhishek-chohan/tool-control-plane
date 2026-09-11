#!/usr/bin/env bash
# End-to-end quickstart demo: the README "First Offload Path" as one command.
#
# Boots an in-memory Toolplane server, runs example_client.py (the provider:
# registers tools and serves them), then example_user.py (the consumer:
# discovers tools, submits work, polls results) against the session the
# provider printed. Tears everything down on exit.
#
# Usage:  cd server && make demo        (or run this script directly)
# Needs:  Go (for the server) and python3 with the client requirements.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
server_dir="$script_dir/.."
client_dir="$server_dir/../clients/python-client"
demo_dir="$server_dir/.tmp-demo"
api_key="demo-key"
grpc_port="${TOOLPLANE_DEMO_PORT:-9001}"

mkdir -p "$demo_dir"
server_log="$demo_dir/server.log"
provider_log="$demo_dir/provider.log"
server_bin="$demo_dir/toolplane-server"

cleanup() {
	[[ -n "${provider_pid:-}" ]] && kill "$provider_pid" 2>/dev/null || true
	[[ -n "${server_pid:-}" ]] && kill "$server_pid" 2>/dev/null || true
	wait 2>/dev/null || true
}
trap cleanup EXIT

echo "==> Building in-memory server..."
(cd "$server_dir" && go build -o "$server_bin" ./cmd/server)

echo "==> Checking python client dependencies..."
if ! python3 -c "import grpc, google.api.annotations_pb2" 2>/dev/null; then
	echo "The demo needs the Python client requirements. Run:" >&2
	echo "  pip install -r $client_dir/requirements.txt" >&2
	exit 1
fi

echo "==> Starting server on :$grpc_port (log: $server_log)..."
TOOLPLANE_AUTH_MODE=fixed \
TOOLPLANE_AUTH_FIXED_API_KEY="$api_key" \
TOOLPLANE_STORAGE_MODE=memory \
	"$server_bin" --port "$grpc_port" --metrics-listen "" >"$server_log" 2>&1 &
server_pid=$!

for _ in $(seq 1 60); do
	if python3 -c "import socket; socket.create_connection(('127.0.0.1', $grpc_port), timeout=0.5).close()" 2>/dev/null; then
		break
	fi
	if ! kill -0 "$server_pid" 2>/dev/null; then
		echo "server exited during startup:" >&2
		tail -20 "$server_log" >&2
		exit 1
	fi
	sleep 0.5
done

echo "==> Starting provider (example_client.py; log: $provider_log)..."
(
	cd "$client_dir"
	PYTHONUNBUFFERED=1 \
	TOOLPLANE_SERVER_HOST=127.0.0.1 TOOLPLANE_SERVER_PORT="$grpc_port" TOOLPLANE_API_KEY="$api_key" \
		python3 example_client.py
) >"$provider_log" 2>&1 &
provider_pid=$!

echo "==> Waiting for the provider to print its session ID..."
session_id=""
for _ in $(seq 1 60); do
	# grep exits 1 on no match; tolerate it inside set -e.
	session_id="$(grep -oE 'TOOLPLANE_SESSION_ID=[A-Za-z0-9-]+' "$provider_log" 2>/dev/null | head -1 | cut -d= -f2 || true)"
	[[ -n "$session_id" ]] && break
	if ! kill -0 "$provider_pid" 2>/dev/null; then
		echo "provider exited before printing a session ID:" >&2
		tail -20 "$provider_log" >&2
		exit 1
	fi
	sleep 0.5
done
if [[ -z "$session_id" ]]; then
	echo "timed out waiting for the provider session ID:" >&2
	tail -20 "$provider_log" >&2
	exit 1
fi
echo "    session: $session_id"

# The session ID is printed before tool registration and runtime start;
# wait for the runtime-ready marker so the consumer cannot race the
# provider and find no tools.
echo "==> Waiting for the provider runtime to report ready..."
for _ in $(seq 1 60); do
	if grep -q "ready to claim requests" "$provider_log" 2>/dev/null; then
		break
	fi
	if ! kill -0 "$provider_pid" 2>/dev/null; then
		echo "provider exited before its runtime was ready:" >&2
		tail -20 "$provider_log" >&2
		exit 1
	fi
	sleep 0.5
done

echo "==> Running consumer (example_user.py)..."
cd "$client_dir"
TOOLPLANE_SERVER_HOST=127.0.0.1 TOOLPLANE_SERVER_PORT="$grpc_port" TOOLPLANE_API_KEY="$api_key" \
TOOLPLANE_SESSION_ID="$session_id" \
	python3 example_user.py

echo ""
echo "Demo complete. Server log: $server_log; provider log: $provider_log."
