# Fritz.Box CloudFlare DynDNS Server

This containerized server application lets you update your CloudFlare DNS records with your FRITZ!Box router.

To develop this application I was inspired by [cloudflare-dyndns](https://github.com/L480/cloudflare-dyndns/), which sadly lacked a few features important to me.
I took the chance to write my own version as my first Go project to learn the language.

To change the port, edit the `WEB_PORT` env variable inside the `docker-compose.yml`. By default a DNS entry for the zone only will be created as well (e.g. `example.com` for `www.example.com`)

## Features
* Update your IPv4 and IPv6 DNS entries dynamically, right when your router notices them change.
* Create DNS entries if they do not exist yet.
* Specify multiple sub domains to update/create

## Getting started

### Create a Cloudflare API token

Create a [Cloudflare API token](https://dash.cloudflare.com/profile/api-tokens) with **read permissions** for the scope `Zone.Zone` and **edit permissions** for the scope `Zone.DNS`.

### Run with Docker Compose

Clone the repository and run the following command:

```bash
docker compose up
```

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
