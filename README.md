# Traefik 3.0 + Pebble local development stack

This project run 2 docker containers, traefik and pebble.

Pebble is a really light implementation of Let's Encrypt ACME protocol.
We use the podCloud fork [@podcloud/pebble-static-CA](https://github.com/podcloud/pebble-static-CA) that support a static CA root file.

Basically, you run this stack, and boum you have a local https development TLD : `.test`
With auto proxying and auto certificate generation with traefik and pebble

## Installation

```shell
mkdir -p $HOME/.docker/traefik
git clone https://github.com/PofMagicfingers/traefik-pebble-stack.git $HOME/.docker/traefik
cd $HOME/.docker/traefik

docker network create --subnet=172.16.0.0/16 traefik
docker-compose up -d
```

- Servers are set up to always restart
- CA will generate only once
- All intermediates certificates are lost on restart

## DNSmasq config

You can configure DNSmasq inside NetworkManager to resolve `.test` to the local traefik container.

```dnsmasq.conf
local=/test/
address=/test/172.16.0.250
```

then you can restart NetworkManager with `sudo systemctl restart NetworkManager`

** Note : On Windows you can use Technitium DNS Server. **

## Trusted CA

On most linux systems, you can add a trusted CA with this command :

```shell
cd $HOME/.docker/traefik
certutil -d sql:$HOME/.pki/nssdb -A -t "CT,C,C" -n "Traefik Pebble" -i ca/ca.crt
```
