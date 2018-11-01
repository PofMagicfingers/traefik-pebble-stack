#!/bin/sh

rm /var/run/acme.json

echo "Waiting for certificate"
while [ ! -f "/usr/local/share/ca-certificates/cert.pem" ]; do
    sleep 5
done;
update-ca-certificates

exec /entrypoint.sh "$@"
