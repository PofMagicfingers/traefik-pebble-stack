#!/bin/bash

cd /var/pebble/certs/

if [ ! -f "ca/ca.crt" ]; then
	openssl genrsa -out ca/ca.key 4096
	openssl req -x509 -subj "/C=US/ST=CA/O=Pebble Static CA, Inc./CN=pebble static CA" -new -nodes -key ca/ca.key -sha256 -days 36500 -out ca/ca.crt
fi

if [ ! -f "server/server.crt" ]; then
	openssl genrsa -out server/server.key 2048
	openssl req -new -sha256 -key server/server.key \
		-subj "/C=US/ST=CA/O=Pebble Static CA, Inc./CN=pebble" -out server/server.csr
	openssl x509 -req -in server/server.csr \
		-extfile <(printf "subjectAltName=DNS:pebble") \
		-CA ca/ca.crt -CAkey ca/ca.key \
		-CAcreateserial -out server/server.crt \
		-days 3650 -sha256
fi

cd -

exec "$@"
