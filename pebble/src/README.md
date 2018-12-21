# Pebble

The original Pebble is a miniature version of [Boulder](https://github.com/letsencrypt/boulder), not suited for use as a production CA.

See the original project [here](https://github.com/letsencrypt/pebble)

The goal of this fork from [zimosworld](https://github.com/zimosworld) is to provide a internal certificate authority (that skips verification) for a team of developers,
 where you have one self signed root certificate that all your team's computers trust.

## !!! WARNING !!!

This fork of Pebble is **NOT INTENDED FOR PRODUCTION USE** and assumes you have at a minimum a basic understanding of how SSL works.

## Example Scenario

You have a team of developers using Docker that all work on more than one project (could even be one project with many services).

In the above situation you could use [Traefik](https://traefik.io/) as a reserve proxy, routing requests to the different containers.
You could also then setup a pebble server using this fork and configure [Traefik](https://traefik.io/) to use it to automatically generate SSL Certificates using your self signed root certificate.

## Install

Pebble will include the self signed root certificate from `/var/pebble/certs/ca/cert.pem` and the private key from `/var/pebble/certs/ca/key.pem`.

### Manual

1. [Set up Go](https://golang.org/doc/install) and your `$GOPATH`
2. `go get -u github.com/zimosworld/pebble/...`
3. `cd $GOPATH/src/github.com/zimosworld/pebble && go install ./...`
4. `PEBBLE_WFE_NONCEREJECT=0 PEBBLE_VA_ALWAYS_VALID=1 PEBBLE_VA_NOSLEEP=1 pebble -config /var/pebble/config/pebble-config.json`
