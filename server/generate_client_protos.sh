#!/usr/bin/env bash
#
# Syncs the single editable protobuf contract (server/proto/service.proto)
# into tracked client SDK folders and regenerates all checked-in stubs.
#
# Canonical invocation: cd server && make gen-proto-all
# See also: server/docs/proto_regeneration.md
#

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
server_dir="$repo_root/server"

copy_file() {
  local source_path="$1"
  local target_path="$2"

  mkdir -p "$(dirname "$target_path")"

  if [[ "$target_path" == *.proto ]]; then
    # Prepend a synced-copy header so the checked-in copy is visibly non-editable.
    {
      echo '// SYNCED COPY — DO NOT EDIT.'
      echo '// Source: server/proto/service.proto'
      echo '// Regenerate: cd server && make gen-proto-all'
      echo ''
      while IFS= read -r line || [[ -n "$line" ]]; do
        printf '%s\n' "${line%$'\r'}"
      done < "$source_path"
    } > "$target_path"
  else
    cp "$source_path" "$target_path"
  fi
}

require_command() {
  local command_name="$1"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Required command not found: $command_name" >&2
    exit 1
  fi
}

sync_client_proto_inputs() {
  echo "Syncing shared proto inputs into tracked client folders..."

  copy_file "$server_dir/proto/service.proto" "$repo_root/clients/go-client/proto/service.proto"
  copy_file "$server_dir/proto/service.proto" "$repo_root/clients/typescript-client/proto/service.proto"

  copy_file "$server_dir/google/api/annotations.proto" "$repo_root/clients/go-client/google/api/annotations.proto"
  copy_file "$server_dir/google/api/http.proto" "$repo_root/clients/go-client/google/api/http.proto"
  copy_file "$server_dir/google/api/annotations.proto" "$repo_root/clients/typescript-client/google/api/annotations.proto"
  copy_file "$server_dir/google/api/http.proto" "$repo_root/clients/typescript-client/google/api/http.proto"
}

generate_go_client_protos() {
  echo "Generating tracked Go client protos..."
  (
    cd "$repo_root/clients/go-client"
    protoc \
      --proto_path=. \
      --go_out=paths=source_relative:. \
      --go-grpc_out=paths=source_relative:. \
      proto/service.proto
  )
}

generate_python_client_protos() {
  echo "Generating tracked Python client protos..."
  (
    cd "$repo_root"
    python3 -m grpc_tools.protoc \
      --proto_path=server \
      --python_out=clients/python-client/toolplane \
      --grpc_python_out=clients/python-client/toolplane \
      server/proto/service.proto
    python3 server/scripts/fix_python_proto_imports.py clients/python-client/toolplane/proto/service_pb2_grpc.py
  )
}

generate_typescript_client_protos() {
  echo "Generating tracked TypeScript client protos..."
  (
    cd "$repo_root/clients/typescript-client"
    npm run generate:proto
  )
}

main() {
  require_command protoc
  require_command python3
  require_command npm

  sync_client_proto_inputs
  generate_go_client_protos
  generate_python_client_protos
  generate_typescript_client_protos

  echo "Tracked client proto artifacts regenerated successfully."
  echo "Canonical full regeneration command: cd server && make gen-proto-all"
}

main "$@"