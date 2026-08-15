#!/bin/sh
# Generates a self-signed TLS cert on first boot if none exists yet, then
# hands off to nginx. Once a real domain + Let's Encrypt cert is in place,
# drop this and mount the real fullchain.pem/privkey.pem instead.
set -e

CERT_DIR=/etc/nginx/certs
CN=${SSL_CN:-localhost}

if [ ! -f "$CERT_DIR/fullchain.pem" ] || [ ! -f "$CERT_DIR/privkey.pem" ]; then
    echo "==> No certificate found, generating a self-signed one for CN=$CN"
    apk add --no-cache openssl >/dev/null
    mkdir -p "$CERT_DIR"
    openssl req -x509 -nodes -days 825 -newkey rsa:2048 \
        -keyout "$CERT_DIR/privkey.pem" \
        -out "$CERT_DIR/fullchain.pem" \
        -subj "/CN=$CN" \
        -addext "subjectAltName=DNS:$CN,IP:127.0.0.1"
fi

exec nginx -g "daemon off;"
