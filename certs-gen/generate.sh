#!/usr/bin/env bash
set -e

# Get absolute path of workspace root (parent of script directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(dirname "$SCRIPT_DIR")"

# Run OpenSSL inside a temporary Docker container to generate certificates
docker run --rm -v "$WORKSPACE_ROOT:/export" alpine sh -c "
  apk add --no-cache openssl && \
  openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
    -keyout /export/certs/tls.key \
    -out /export/certs/tls.crt \
    -config /export/certs-gen/tls-conf.cnf
"

echo "Success: Certificates generated in certs/ directory!"
