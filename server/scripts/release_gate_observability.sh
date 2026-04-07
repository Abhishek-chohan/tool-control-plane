#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
server_dir="$(cd "$script_dir/.." && pwd)"
log_dir="${TOOLPLANE_RELEASE_GATE_LOG_DIR:-$server_dir/../.tmp/release-gate-observability}"

python_bin="${PYTHON_BIN:-python}"
if ! command -v "$python_bin" >/dev/null 2>&1; then
	python_bin="python3"
fi
if ! command -v "$python_bin" >/dev/null 2>&1; then
	echo "python or python3 is required for release_gate_observability.sh" >&2
	exit 1
fi

find_free_port() {
	"$python_bin" - <<'PY'
import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.bind(("127.0.0.1", 0))
print(sock.getsockname()[1])
sock.close()
PY
}

wait_for_tcp() {
	"$python_bin" - "$1" "$2" "$3" <<'PY'
import socket
import sys
import time

host = sys.argv[1]
port = int(sys.argv[2])
deadline = time.time() + float(sys.argv[3])

while time.time() < deadline:
	probe = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	probe.settimeout(0.5)
	try:
		probe.connect((host, port))
	except OSError:
		time.sleep(0.2)
	else:
		probe.close()
		sys.exit(0)
	finally:
		probe.close()

sys.exit(1)
PY
}

wait_for_http_ok() {
	"$python_bin" - "$1" "$2" <<'PY'
import sys
import time
import urllib.error
import urllib.request

url = sys.argv[1]
deadline = time.time() + float(sys.argv[2])

while time.time() < deadline:
	try:
		with urllib.request.urlopen(url, timeout=1.0) as response:
			if response.status == 200:
				sys.exit(0)
	except (urllib.error.URLError, TimeoutError, ValueError):
		pass
	time.sleep(0.2)

sys.exit(1)
PY
}

validate_metrics() {
	"$python_bin" - "$1" <<'PY'
import sys
import urllib.request

url = sys.argv[1]
with urllib.request.urlopen(url, timeout=2.0) as response:
	body = response.read().decode("utf-8")

required_metrics = [
	"toolplane_request_queue_depth",
	"toolplane_request_inflight",
	"toolplane_request_dead_letter_current",
	"toolplane_request_requeues_total",
	"toolplane_request_dead_letters_total",
	"toolplane_machine_active",
	"toolplane_machine_draining",
	"toolplane_machine_inflight_load",
	"toolplane_task_pending",
	"toolplane_task_running",
	"toolplane_task_dead_letter_current",
	"toolplane_task_retries_total",
	"toolplane_task_dead_letters_total",
]

missing = [name for name in required_metrics if name not in body]
if missing:
	raise SystemExit(f"missing metrics from {url}: {', '.join(missing)}")
PY
}

validate_health() {
	"$python_bin" - "$1" <<'PY'
import json
import sys
import urllib.request

url = sys.argv[1]
with urllib.request.urlopen(url, timeout=2.0) as response:
	payload = json.loads(response.read().decode("utf-8"))

required_top_level = ["status", "circuit", "rateLimitRejects", "throttle", "timestamp"]
missing_top_level = [name for name in required_top_level if name not in payload]
if missing_top_level:
	raise SystemExit(f"missing /health fields from {url}: {', '.join(missing_top_level)}")

if payload["status"] != "ok":
	raise SystemExit(f"unexpected /health status from {url}: {payload['status']}")

circuit = payload["circuit"]
required_circuit = ["state", "inflight", "rejected", "accepted", "counts"]
missing_circuit = [name for name in required_circuit if name not in circuit]
if missing_circuit:
	raise SystemExit(f"missing circuit fields from {url}: {', '.join(missing_circuit)}")

counts = circuit["counts"]
required_counts = ["Requests", "TotalSuccesses", "TotalFailures", "ConsecutiveSuccesses", "ConsecutiveFailures"]
missing_counts = [name for name in required_counts if name not in counts]
if missing_counts:
	raise SystemExit(f"missing circuit count fields from {url}: {', '.join(missing_counts)}")

throttle = payload["throttle"]
required_throttle = ["total", "apiRate", "ipRate", "concurrency", "circuitOpen", "circuitProbe", "unknown"]
missing_throttle = [name for name in required_throttle if name not in throttle]
if missing_throttle:
	raise SystemExit(f"missing throttle fields from {url}: {', '.join(missing_throttle)}")
PY
}

if [[ -z "${TOOLPLANE_STORAGE_MODE:-}" ]]; then
	if [[ -n "${TOOLPLANE_DATABASE_URL:-}" ]]; then
		export TOOLPLANE_STORAGE_MODE=postgres
	else
		export TOOLPLANE_STORAGE_MODE=memory
	fi
fi

: "${TOOLPLANE_ENV_MODE:=development}"
: "${TOOLPLANE_AUTH_MODE:=fixed}"
: "${TOOLPLANE_AUTH_FIXED_API_KEY:=toolplane-conformance-fixture-key}"
: "${TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND:=1}"

export TOOLPLANE_ENV_MODE
export TOOLPLANE_AUTH_MODE
export TOOLPLANE_AUTH_FIXED_API_KEY
export TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND

grpc_port="${TOOLPLANE_RELEASE_GATE_GRPC_PORT:-$(find_free_port)}"
http_port="${TOOLPLANE_RELEASE_GATE_HTTP_PORT:-$(find_free_port)}"
metrics_port="${TOOLPLANE_RELEASE_GATE_METRICS_PORT:-$(find_free_port)}"
metrics_listen="127.0.0.1:${metrics_port}"

mkdir -p "$log_dir"
server_log="$log_dir/server.log"
proxy_log="$log_dir/proxy.log"

server_pid=""
proxy_pid=""

cleanup() {
	local exit_code=$?
	if [[ -n "$proxy_pid" ]] && kill -0 "$proxy_pid" >/dev/null 2>&1; then
		kill "$proxy_pid" >/dev/null 2>&1 || true
		wait "$proxy_pid" >/dev/null 2>&1 || true
	fi
	if [[ -n "$server_pid" ]] && kill -0 "$server_pid" >/dev/null 2>&1; then
		kill "$server_pid" >/dev/null 2>&1 || true
		wait "$server_pid" >/dev/null 2>&1 || true
	fi
	if [[ $exit_code -ne 0 ]]; then
		echo "release-gate observability logs:" >&2
		echo "  server: $server_log" >&2
		echo "  proxy:  $proxy_log" >&2
	fi
}
trap cleanup EXIT

cd "$server_dir"

go run ./cmd/server --port "$grpc_port" --metrics-listen "$metrics_listen" >"$server_log" 2>&1 &
server_pid=$!

if ! wait_for_tcp "127.0.0.1" "$grpc_port" 60; then
	echo "gRPC server did not become ready on port $grpc_port" >&2
	exit 1
fi

if ! wait_for_http_ok "http://127.0.0.1:${metrics_port}/metrics" 60; then
	echo "metrics endpoint did not become ready on port $metrics_port" >&2
	exit 1
fi

go run ./cmd/proxy --listen ":${http_port}" --backend "localhost:${grpc_port}" >"$proxy_log" 2>&1 &
proxy_pid=$!

if ! wait_for_http_ok "http://127.0.0.1:${http_port}/health" 60; then
	echo "proxy health endpoint did not become ready on port $http_port" >&2
	exit 1
fi

validate_metrics "http://127.0.0.1:${metrics_port}/metrics"
validate_health "http://127.0.0.1:${http_port}/health"

echo "release-gate observability check passed (grpc=${grpc_port} http=${http_port} metrics=${metrics_port})"
