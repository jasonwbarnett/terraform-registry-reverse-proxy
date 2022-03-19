# terraform-registry-reverse-proxy

This is used to proxy internal requests back to Terraform's Registry.

## What is this?

Hashicorp for some reason either hasn't prioritized or outright refuses to make
it easy to ingest Terraform providers and modules like you might traditionally
see in an Artifactory so that you don't have to cache the plugins and unzip them.
Instead just let the `terraform` cli natively fetch things on-demand in an
environment where direct internet access is not possible.

This application is intended to be put behind a web server, e.g. [NGINX][1], [Caddy][2],
[Apache][3], etc.

## Requirements

- web server binding on TCP/443
- ssl certificate(s) that is/are trusted by the client where `terraform` is
  being run
- dns record dedicated to terraform registry proxy (e.g. `terraform-registry.company.com`)


### Optionally

- artifact storage (e.g. Artifactory)
- dns record dedicated to hashicorp releases proxy (e.g. `hashicorp-releases.company.com`)

## Usage

Two possible usages:

1. Without external artifact storage
2. With external artifact storage

Read each section below for more details

### Without external artifact storage

```bash
./terraform-registry-reverse-proxy -registry-proxy-host terraform-registry.company.com \
                                   -release-proxy-host hashicorp-releases.company.com
```

### With external artifact storage

```bash
./terraform-registry-reverse-proxy -registry-proxy-host terraform-registry.company.com \
                                   -release-proxy-host artifactory.company.com \
                                   -release-proxy-path-prefix /artifactory/hashicorp-releases
```

[1]: https://nginx.org/en/
[2]: https://caddyserver.com/
[3]: https://httpd.apache.org/
