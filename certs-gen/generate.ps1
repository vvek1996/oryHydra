# Get absolute path of workspace root (parent of script root)
$WorkspaceRoot = (Get-Item $PSScriptRoot).Parent.FullName

# Run OpenSSL inside a temporary Docker container to generate certificates
docker run --rm -v "${WorkspaceRoot}:/export" alpine sh -c "
  apk add --no-cache openssl && \
  openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
    -keyout /export/certs/tls.key \
    -out /export/certs/tls.crt \
    -config /export/certs-gen/tls-conf.cnf
"

Write-Host "Success: Certificates generated in certs/ directory!" -ForegroundColor Green
