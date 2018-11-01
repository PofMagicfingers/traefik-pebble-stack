# Traefik-Pebble stack

This project run 2 docker containers, traefik and pebble. 
Pebble is a really light implementation of Let's Encrypt ACME protocol.
We're using zimosworld fork's which allow to keep the same CA between launches.

Basically, you run this stack, and boum you have a local https development TLD
with auto proxying and auto certificate generation 

# Installation

```shell
mkdir -p $HOME/.docker/traefik
git clone https://github.com/PofMagicfingers/traefik-pebble-stack.git $HOME/.docker/traefik
cd $HOME/.docker/traefik

docker network create --subnet=172.10.0.0/16 traefik
docker-compose up -d
```

servers are set up to always restart, CA will generate only once, all other certificates are lost on restart

# DNSmasq config
```
local=/test/
address=/test/172.10.0.10 
```

# Trusted CA

On most linux systems, you can add a trusted CA with this command : 
```shell
cd $HOME/.docker/traefik
certutil -d sql:$HOME/.pki/nssdb -A -t "CT,C,C" -n "Traefik Pebble" -i ca/cert.pem
```
