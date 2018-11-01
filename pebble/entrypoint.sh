#!/bin/sh

cd /var/pebble/certs/

if [ ! -f "ca/cert.pem" ]; then
    openssl genrsa -out ca/key.pem 4096
    openssl req -x509 -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=pebble CA" -new -nodes -key ca/key.pem -sha256 -days 36500 -out ca/cert.pem
fi

if [ ! -f "localhost/cert.pem" ]; then
    openssl genrsa -out localhost/key.pem 2048
    openssl req -new -sha256 -key localhost/key.pem \
        -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=pebble" -out localhost/request.csr
    openssl x509 -req -in localhost/request.csr \
        -CA ca/cert.pem -CAkey ca/key.pem \
        -CAcreateserial -out localhost/cert.pem -days 3650 -sha256
fi

cd -

exec "$@"
