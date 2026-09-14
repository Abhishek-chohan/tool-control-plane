# Reference TLS Materials

The reference deployment mounts these files from this directory:

- `ca.crt` — the CA trust bundle used by both gateways for backend TLS and by clients pinning the edge
- `server.crt` / `server.key` — the gRPC server certificate and key
- `gateway.crt` / `gateway.key` — client-facing TLS for `toolplane-gateway`
- `mcp-gateway.crt` / `mcp-gateway.key` — client-facing TLS for `toolplane-mcp-gateway`

SAN requirements: the server certificate must be valid for `server`, `localhost`, and `127.0.0.1`; each gateway certificate for its service name (`gateway` / `mcp-gateway`), `localhost`, and `127.0.0.1`.

Example OpenSSL flow:

```bash
cd server/deploy/reference/certs

openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 \
  -subj "/CN=Toolplane Reference CA" \
  -out ca.crt

# --- gRPC server ---
openssl genrsa -out server.key 4096
openssl req -new -key server.key -subj "/CN=server" -out server.csr

cat > server.ext <<'EOF'
subjectAltName = DNS:server,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 825 -sha256 -extfile server.ext

# --- Edge gateways (same flow, per service name) ---
for edge in gateway mcp-gateway; do
  openssl genrsa -out "$edge.key" 4096
  openssl req -new -key "$edge.key" -subj "/CN=$edge" -out "$edge.csr"

  cat > "$edge.ext" <<EOF
subjectAltName = DNS:$edge,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

  openssl x509 -req -in "$edge.csr" -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out "$edge.crt" -days 825 -sha256 -extfile "$edge.ext"
done
```

`scripts/reference_deployment_integration.sh` generates this full set automatically (with 2-day certs) for its throwaway project directory.

Keep `ca.key` outside shared deployment volumes. The reference compose file mounts `ca.crt`, the server pair, and both edge-gateway pairs.
