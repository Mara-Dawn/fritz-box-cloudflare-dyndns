# DynDNS Middleware for FRITZ!Box and CloudFlare

This containerized server application lets you update your CloudFlare DNS records with your FRITZ!Box router.

To develop this application I was inspired by [cloudflare-dyndns](https://github.com/L480/cloudflare-dyndns/), which sadly lacked a few features important to me.
I took the chance to write my own version as my first Go project to learn the language.

By default a DNS entry for the zone only will be created as well (e.g. `example.com` for `www.example.com`)

## Features
* Update your IPv4 and IPv6 DNS entries dynamically, right when your router notices them change.
* Create DNS entries if they do not exist yet.
* Specify multiple sub domains to update/create

## Getting started

### Create a Cloudflare API token

Create a [Cloudflare API token](https://dash.cloudflare.com/profile/api-tokens) with **read permissions** for the scope `Zone.Zone` and **edit permissions** for the scope `Zone.DNS`.

### Option 1: Use my free hosted instance

Use the server instance hosted by me. Just replace the update URL below with this one, no need to download or run anything yourself:

```
https://ddns.mara.cafe/?token=<pass>&records=www,test,home&zone=example.com&ipv4=<ipaddr>&ipv6=<ip6addr>
```

### Option 2: Run hosted docker image

Use the prebuilt docker image. Adjust the ports as needed.

```bash
docker run -e WEB_PORT=8070 -p 8070:8070 ghcr.io/mara-dawn/fritz-box-cloudflare-dyndns:latest
```

### Option 3: Run with docker compose

Download the latest release .zip, extract it and run the following command:

```bash
docker compose up --build
```
To change the port, edit the `WEB_PORT` env variable and the port bindings inside the `docker-compose.yml`. 

### Configure your FRITZ!Box

Navigate to Internet > Permit Access > DynDNS in your router web interface.

#### Update URL
```https://your.host.here:8070/?token=<pass>&records=www,test,home&zone=example.com&ipv4=<ipaddr>&ipv6=<ip6addr>```

Replace the URL parameter `records` and `zone` with your domain name.
If required you can omit either the `ipv4` or `ipv6` URL parameter.
Multiple records are separated by a `,`.
Leave the parts in `< >` as is, they will be filled out by your router automatically.

#### Domain Name
```example.com```

The FQDN from the URL parameter `record` and `zone`.

#### Username

Does not matter, you can enter whaterver you want.

#### Password

Your ClourFlare API token goes here.
