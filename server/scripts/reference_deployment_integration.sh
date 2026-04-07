#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
server_dir="$(cd "$script_dir/.." && pwd)"
reference_dir="$server_dir/deploy/reference"
compose_file="$reference_dir/compose.yaml"
log_root="${TOOLPLANE_REFERENCE_DEPLOYMENT_LOG_DIR:-$server_dir/../.tmp/reference-deployment-integration}"
python_bin="${PYTHON_BIN:-python3}"
verbose="${TOOLPLANE_REFERENCE_VERBOSE:-0}"

if ! command -v docker >/dev/null 2>&1; then
	echo "docker is required for reference_deployment_integration.sh" >&2
	exit 1
fi
if ! docker compose version >/dev/null 2>&1; then
	echo "docker compose is required for reference_deployment_integration.sh" >&2
	exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
	echo "openssl is required for reference_deployment_integration.sh" >&2
	exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
	echo "curl is required for reference_deployment_integration.sh" >&2
	exit 1
fi
if ! command -v "$python_bin" >/dev/null 2>&1; then
	echo "Python 3 is required for reference_deployment_integration.sh; set PYTHON_BIN to a Python 3 executable if needed" >&2
	exit 1
fi
if ! "$python_bin" - <<'PY'
import sys

sys.exit(0 if sys.version_info.major >= 3 else 1)
PY
then
	echo "reference_deployment_integration.sh requires Python 3; PYTHON_BIN resolved to an incompatible interpreter" >&2
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
	time.sleep(0.5)

sys.exit(1)
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
		time.sleep(0.5)
	else:
		probe.close()
		sys.exit(0)
	finally:
		probe.close()

sys.exit(1)
PY
}

wait_for_service_health() {
	local service="$1"
	local timeout_seconds="$2"
	local deadline=$((SECONDS + timeout_seconds))
	local container_id=""

	while [[ $SECONDS -lt $deadline ]]; do
		container_id="$(compose ps -q "$service" 2>/dev/null || true)"
		if [[ -n "$container_id" ]]; then
			local health_status
			health_status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null || true)"
			if [[ "$health_status" == "healthy" ]]; then
				return 0
			fi
		fi
		sleep 1
	done

	echo "service $service did not become healthy within ${timeout_seconds}s" >&2
	return 1
}

validate_metrics() {
	"$python_bin" - "$1" <<'PY'
import sys
import urllib.request

url = sys.argv[1]
with urllib.request.urlopen(url, timeout=5.0) as response:
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
with urllib.request.urlopen(url, timeout=5.0) as response:
	payload = json.loads(response.read().decode("utf-8"))

if payload.get("status") != "ok":
	raise SystemExit(f"unexpected /health status from {url}: {payload.get('status')}")

required_top_level = ["status", "circuit", "rateLimitRejects", "throttle", "timestamp"]
missing_top_level = [name for name in required_top_level if name not in payload]
if missing_top_level:
	raise SystemExit(f"missing /health fields from {url}: {', '.join(missing_top_level)}")
PY
}

extract_json_field() {
	"$python_bin" - "$1" "$2" <<'PY'
import json
import sys

payload = json.loads(sys.argv[1])
field_path = sys.argv[2].split('.')
value = payload
for part in field_path:
	if not isinstance(value, dict) or part not in value:
		raise SystemExit(f"missing field path: {sys.argv[2]}")
	value = value[part]

if isinstance(value, (dict, list)):
	print(json.dumps(value))
else:
	print(value)
PY
}

validate_session_lookup() {
	"$python_bin" - "$1" "$2" <<'PY'
import json
import sys

payload = json.loads(sys.argv[1])
expected_session_id = sys.argv[2]
session = payload.get("session") if isinstance(payload, dict) else None
if not isinstance(session, dict) and isinstance(payload, dict):
	session = payload
if not isinstance(session, dict):
	raise SystemExit("GetSession response did not include a session payload")
if session.get("id") != expected_session_id:
	raise SystemExit(f"GetSession returned {session.get('id')}, expected {expected_session_id}")
PY
}

generate_reference_certs() {
	local cert_dir="$1"
	mkdir -p "$cert_dir"

	openssl genrsa -out "$cert_dir/ca.key" 2048 >/dev/null 2>&1
	openssl req -x509 -new -nodes -key "$cert_dir/ca.key" -sha256 -days 2 \
		-subj "/CN=Toolplane Reference Test CA" \
		-out "$cert_dir/ca.crt" >/dev/null 2>&1

	openssl genrsa -out "$cert_dir/server.key" 2048 >/dev/null 2>&1
	openssl req -new -key "$cert_dir/server.key" -subj "/CN=server" -out "$cert_dir/server.csr" >/dev/null 2>&1

	cat >"$cert_dir/server.ext" <<'EOF'
subjectAltName = DNS:server,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

	openssl x509 -req -in "$cert_dir/server.csr" -CA "$cert_dir/ca.crt" -CAkey "$cert_dir/ca.key" -CAcreateserial \
		-out "$cert_dir/server.crt" -days 2 -sha256 -extfile "$cert_dir/server.ext" >/dev/null 2>&1
}

project_name="toolplane-reference-it-$$-$(date +%s)"
cert_subdir="./certs/${project_name}"
cert_dir="$reference_dir/certs/${project_name}"
env_file="$reference_dir/.env.${project_name}"
log_dir="$log_root/${project_name}"

postgres_port="${TOOLPLANE_REFERENCE_POSTGRES_PORT:-$(find_free_port)}"
grpc_port="${TOOLPLANE_REFERENCE_GRPC_PORT:-$(find_free_port)}"
http_port="${TOOLPLANE_REFERENCE_HTTP_PORT:-$(find_free_port)}"
metrics_port="${TOOLPLANE_REFERENCE_METRICS_PORT:-$(find_free_port)}"
bootstrap_http_port="${TOOLPLANE_REFERENCE_BOOTSTRAP_HTTP_PORT:-$(find_free_port)}"
bootstrap_token="${TOOLPLANE_REFERENCE_BOOTSTRAP_TOKEN:-toolplane-bootstrap-${project_name}}"
image_name="${TOOLPLANE_REFERENCE_IMAGE:-toolplane-reference:reference-deployment-integration}"
skip_build="${TOOLPLANE_REFERENCE_SKIP_BUILD:-auto}"

mkdir -p "$log_dir"
build_log="$log_dir/build.log"
run_log="$log_dir/run.log"

log_step() {
	echo "[reference-deployment] $*"
}

run_logged() {
	local log_file="$1"
	shift
	if [[ "$verbose" == "1" ]]; then
		"$@" 2>&1 | tee "$log_file"
	else
		"$@" >"$log_file" 2>&1
	fi
}

compose() {
	docker compose \
		--project-name "$project_name" \
		--env-file "$env_file" \
		-f "$compose_file" \
		"$@"
}

cleanup() {
	local exit_code=$?
	if [[ $exit_code -ne 0 ]]; then
		compose logs --no-color >"$log_dir/compose.log" 2>&1 || true
		echo "reference deployment integration logs: $log_dir" >&2
	fi
	compose down -v --remove-orphans >/dev/null 2>&1 || true
	rm -rf "$cert_dir" "$env_file"
}
trap cleanup EXIT

generate_reference_certs "$cert_dir"

cat >"$env_file" <<EOF
TOOLPLANE_IMAGE=${image_name}
POSTGRES_USER=toolplane
POSTGRES_PASSWORD=toolplane
POSTGRES_DB=toolplane
TOOLPLANE_DATABASE_URL=postgres://toolplane:toolplane@postgres:5432/toolplane?sslmode=disable
TOOLPLANE_POSTGRES_PUBLISHED_PORT=${postgres_port}
TOOLPLANE_GRPC_PUBLISHED_PORT=${grpc_port}
TOOLPLANE_HTTP_PUBLISHED_PORT=${http_port}
TOOLPLANE_METRICS_PUBLISHED_PORT=${metrics_port}
TOOLPLANE_BOOTSTRAP_HTTP_PUBLISHED_PORT=${bootstrap_http_port}
TOOLPLANE_SERVER_TLS_CERT_PATH=${cert_subdir}/server.crt
TOOLPLANE_SERVER_TLS_KEY_PATH=${cert_subdir}/server.key
TOOLPLANE_PROXY_BACKEND_TLS_CA_PATH=${cert_subdir}/ca.crt
TOOLPLANE_PROXY_ALLOWED_ORIGINS=https://example.invalid
TOOLPLANE_PROXY_BACKEND_TLS_SERVER_NAME=server
TOOLPLANE_BOOTSTRAP_FIXED_API_KEY=${bootstrap_token}
EOF

log_step "logs: $log_dir"

if [[ "$skip_build" == "auto" ]]; then
	if docker image inspect "$image_name" >/dev/null 2>&1; then
		skip_build=1
	else
		skip_build=0
	fi
fi

if [[ "$skip_build" == "1" ]] && ! docker image inspect "$image_name" >/dev/null 2>&1; then
	log_step "cached image $image_name not found; building instead"
	skip_build=0
fi

if [[ "$skip_build" != "1" ]]; then
	log_step "building image $image_name"
	run_logged "$build_log" docker build -t "$image_name" "$server_dir"
	log_step "build complete"
else
	log_step "reusing cached image $image_name"
fi

log_step "starting postgres"
compose up -d postgres
wait_for_service_health postgres 120
	log_step "running migrations"
compose run --rm migrate >/dev/null

log_step "starting bootstrap services"
compose --profile bootstrap up -d bootstrap-server bootstrap-gateway

if ! wait_for_http_ok "http://127.0.0.1:${bootstrap_http_port}/health" 120; then
	echo "bootstrap gateway did not become ready on port ${bootstrap_http_port}" >&2
	exit 1
fi

session_json="$(curl -fsS \
	"http://127.0.0.1:${bootstrap_http_port}/api/CreateSession" \
	-H "Authorization: Bearer ${bootstrap_token}" \
	-H 'Content-Type: application/json' \
	-d '{"userId":"bootstrap-admin","name":"bootstrap-admin","description":"one-time admin bootstrap","namespace":"ops"}')"
session_id="$(extract_json_field "$session_json" 'session.id')"

admin_key_json="$(curl -fsS \
	"http://127.0.0.1:${bootstrap_http_port}/api/CreateApiKey" \
	-H "Authorization: Bearer ${bootstrap_token}" \
	-H 'Content-Type: application/json' \
	-d "{\"sessionId\":\"${session_id}\",\"name\":\"reference-admin\",\"capabilities\":[\"read\",\"execute\",\"admin\"]}")"
admin_api_key="$(extract_json_field "$admin_key_json" 'key')"

if [[ -z "$admin_api_key" ]]; then
	echo "bootstrap CreateApiKey response did not return a key" >&2
	exit 1
fi

log_step "stopping bootstrap services"
compose stop bootstrap-server bootstrap-gateway >/dev/null
compose rm -f bootstrap-server bootstrap-gateway >/dev/null

log_step "starting production services"
compose up -d server gateway

if ! wait_for_tcp 127.0.0.1 "$grpc_port" 120; then
	echo "gRPC server did not become ready on port ${grpc_port}" >&2
	exit 1
fi
if ! wait_for_http_ok "http://127.0.0.1:${metrics_port}/metrics" 120; then
	echo "metrics endpoint did not become ready on port ${metrics_port}" >&2
	exit 1
fi
if ! wait_for_http_ok "http://127.0.0.1:${http_port}/health" 120; then
	echo "gateway health endpoint did not become ready on port ${http_port}" >&2
	exit 1
fi

validate_metrics "http://127.0.0.1:${metrics_port}/metrics"
validate_health "http://127.0.0.1:${http_port}/health"

get_session_json="$(curl -fsS \
	"http://127.0.0.1:${http_port}/api/GetSession" \
	-H "Authorization: Bearer ${admin_api_key}" \
	-H 'Content-Type: application/json' \
	-d "{\"sessionId\":\"${session_id}\"}")"
validate_session_lookup "$get_session_json" "$session_id"

cat >"$run_log" <<EOF
postgres=${postgres_port}
grpc=${grpc_port}
http=${http_port}
metrics=${metrics_port}
bootstrap_http=${bootstrap_http_port}
session_id=${session_id}
EOF

echo "reference deployment integration check passed (postgres=${postgres_port} grpc=${grpc_port} http=${http_port} metrics=${metrics_port} bootstrap_http=${bootstrap_http_port}; logs=${log_dir})"