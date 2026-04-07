# Reference TLS Materials

The reference deployment mounts three files from this directory:

- `ca.crt` for the gateway's custom trust bundle
- `server.crt` for the gRPC server certificate
- `server.key` for the gRPC server private key

The reference compose stack expects the server certificate to be valid for `server`, `localhost`, and `127.0.0.1`.

Example OpenSSL flow:

```bash
cd server/deploy/reference/certs

openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 \
  -subj "/CN=Toolplane Reference CA" \
  -out ca.crt

openssl genrsa -out server.key 4096
openssl req -new -key server.key -subj "/CN=server" -out server.csr

cat > server.ext <<'EOF'
subjectAltName = DNS:server,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 825 -sha256 -extfile server.ext
```

Keep `ca.key` outside shared deployment volumes. The reference compose file mounts only `ca.crt`, `server.crt`, and `server.key`.
