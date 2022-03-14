# Traefik-Pebble stack

This project run 2 docker containers, traefik and pebble. 
Pebble is a really light implementation of Let's Encrypt ACME protocol.
We're using zimosworld fork's which allow to keep the same CA between launches.

Basically, you run this stack, and boum you have a local https development TLD
with auto proxying and auto certificate generation 

## Installation

```shell
mkdir -p $HOME/.docker/traefik
git clone https://github.com/occitech/traefik-pebble-stack.git $HOME/.docker/traefik
cd $HOME/.docker/traefik

docker network create --subnet=172.16.0.0/16 traefik
docker-compose up -d
```

servers are set up to always restart, CA will generate only once, all other certificates are lost on restart

## DNSmasq

You can use the module DNSmasq of NetworkManager, but if you want DNSmasq running on its own you can follow steps bellow:

## Install

**With Ubuntu:**
```
sudo apt-get install dnsmasq
```

If `systemd-resolve` is running on your distrib (like with Ubuntu), you need to disable it since it runs on the port 53 like DNSmasq:
```
sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved
sudo systemctl mask systemd-resolved
```

### DNSmasq config

In `/etc/dnsmasq.conf` add this lines to the end:
```
resolv-file=/etc/dnsmasq-dns.conf
local=/test/
address=/test/172.10.0.10 
```

Create file in /etc/dnsmasq-dns.conf and add your DNS server:
```
nameserver 172.16.0.10
nameserver 1.1.1.1
```

Create the resolver for the added address:
```shell
sudo mkdir -v /etc/resolver && sudo bash -c 'echo "nameserver 172.16.0.10" > /etc/resolver/test'
```
>Note:
> In `/etc/resolver/test`, the `test` is the same string you put in `/etc/dnsmasq.conf` for `local` and `address`.

**On Ubuntu**, you'll need to edit `/etc/dhcp/dhclient.conf` and uncomment the following line, so the DHCP client will use the local dnsmasq forwarder.
```
prepend domain-name-servers 127.0.0.1;
```

Restart DNSmasq:
```shell
sudo systemctl restart dnsmasq
sudo systemctl enable dnsmasq
```

**On Ubuntu**, apply the new DHCP config:
```
sudo dhclient
```

Execute `sudo systemctl status dnsmasq.service` to check if DNSmasq service is active and running.

At this point you should be able to access to http://traefik.test/.

## Trusted CA

On most linux systems, you can add a trusted CA with this command: 
```shell
cd $HOME/.docker/traefik
certutil -d sql:$HOME/.pki/nssdb -A -t "CT,C,C" -n "Traefik Pebble" -i ca/cert.pem
```

### Mac

```shell
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./ca/cert.pem  
```

### Force browsers to trust the certificate

You can achieve the same goal by adding your certificate as a trusted authority in browsers like Firefox or Chrome.

#### Firefox
- Open the menu > Options > Privacy & Security
- Scroll to the bottom, click on the "View Certificates" button
- Select Authorities tab
- Click on Import button and select the `/.docker/traefik/ca/cert.pem` file.
- Check the checkboxes to trust this CA for website and email, and click Ok.
- Click Ok, again, once you're back to previous window.

#### Chrome
- Open the menu > Settings > Privacy and security > Security
- Scroll to the bottom, click on "Manage certificates"
- Select Authorities tab
- Click on Import button and select the `/.docker/traefik/ca/cert.pem` file.
- Check the checkboxes to trust this CA for website, email and software, and click Ok.


## Troubleshooting

- If Traefik not work, try to use `172.10.0.0` subnet (instead of `172.16.0.0`) 

> Note:
> You have to update `docker-compose.yml` file too and others file configured in DNSmasq part.

- If when you execute `sudo systemctl status dnsmasq.service`, you have an error indicating the port 53 is already in use:
  - double-check if `systemd-resolved` is correctly disables (see above...)
  - check the config of NetworkManager, maybe you already have the DNSmasq module of NetworkManager running.