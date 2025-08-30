#!/usr/bin/env bash
set -euo pipefail

CERTS_DIR="./certs"

mkdir -p "$CERTS_DIR"

# Generate private key
openssl genrsa -out "$CERTS_DIR/server.key" 2048

# Generate self-signed certificate (valid for 365 days) with CN=localhost
openssl req -new -x509 \
  -key "$CERTS_DIR/server.key" \
  -out "$CERTS_DIR/server.crt" \
  -days 365 \
  -subj "/C=US/ST=CA/L=SanFrancisco/O=GoLB Dev/OU=Local/CN=localhost/subjectAltName=DNS:localhost,IP:127.0.0.1"

echo "Successfully generated certificates in $CERTS_DIR/"

