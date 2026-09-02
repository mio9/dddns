# dddns

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/mio9/dddns/release.yml)
![](banner.jpg)

DD(どこでも)DDNS => DDDNS

Any integratable DNS service as Dynamic DNS service.


## Requirements

- A supported DNS provider account with records to update
- Go 1.26 or later (to build from source and develop locally)

## Currently Supported providers

- **Cloudflare** — update existing DNS records via API token and official Cloudflare Go client.
- **No-IP** — update existing DNS records via the [No-IP REST API](https://developer.noip.com/docs/getting-started-with-the-no-ip-api)
- (Suggest more services to support by creating an issue or make a PR!)

## Installation

### Homebrew

```bash
brew tap mio9/tap
brew install mio9/tap/dddns
```

### Manual

```bash
curl -LsSf https://github.com/mio9/dddns/releases/latest/download/dddns-linux-amd64.tar.gz | tar -xzf -
chmod +x dddns
sudo mv dddns /usr/local/bin/ # Optional, if you want to install it globally, move to $PATH instead if you wanted to
```

## Usage

```bash
dddns --config /path/to/config.yaml
# or
dddns -c /path/to/config.yaml
```


## Configuration

Create a YAML file and pass it with `--config` / `-c`.

```yaml
update-interval: "5m"  # optional; blocking timer mode when set

ip-provider:
  url: "https://api.ipify.org"  # optional

providers:
  - type: cloudflare
    zone_id: "your-zone-id"
    # api_token: "your-api-token"  # optional if CLOUDFLARE_API_TOKEN is set
    records:
      - name: "home.example.com"
        type: "A"   # optional, defaults to A
      - name: "vpn.example.com"
```

### Config reference

| Field | Required | Description |
|-------|----------|-------------|
| `update-interval` | no | Periodic update interval (e.g. `30s`, `5m`, `1h`, `1h30m`). When set, dddns runs as a blocking process until stopped with Ctrl+C or SIGTERM |
| `ip-provider.url` | no | URL that returns your public IP as plain text. Default: `https://api.ipify.org` |
| `providers` | yes | List of DNS provider configurations |
| `providers[].type` | yes | Provider name. Supported: `cloudflare`, `no-ip` |
| `providers[].zone_id` | yes* | Cloudflare zone ID |
| `providers[].zone_name` | yes** | No-IP zone name (e.g. `example.com`) |
| `providers[].api_token` | yes* | Cloudflare API token with DNS edit access |
| `providers[].api_key` | yes** | No-IP API key with DNS edit access |
| `providers[].records` | yes* / yes** | DNS records to update for this provider |
| `providers[].records[].name` | yes*** | DNS record name. Cloudflare: full FQDN. No-IP: FQDN or name relative to `zone_name` |
| `providers[].records[].id` | yes*** | Cloudflare DNS record ID (alternative to `name`) |
| `providers[].records[].type` | no | Record type when using `name`. Default: `A`. Use `AAAA` for IPv6 |

\* Required for `cloudflare` providers.

\** Required for `no-ip` providers.

\*** Each record needs `id` **or** `name` (with optional `type`). Named records must already exist in the provider.

Public IP is fetched once per update cycle and applied to all configured records across all providers. When the public IP matches the local cache from the last successful sync, DNS providers are not contacted.

### Timer mode

Set `update-interval` to run dddns as a long-lived process instead of a one-shot cron job:

```yaml
update-interval: "5m"
```

- Runs the first update immediately on start
- Repeats every configured interval
- Logs update errors to stderr and keeps running
- Stop with Ctrl+C or SIGTERM (exit code 0)

Supported units match Go duration syntax: `s`, `m`, `h`, and combinations such as `1h30m`.

### Example with record ID

```yaml
providers:
  - type: cloudflare
    zone_id: "your-zone-id"
    records:
      - id: "dns-record-id"
```

### Cloudflare API token

Create a token in the [Cloudflare dashboard](https://dash.cloudflare.com/profile/api-tokens) with **Edit DNS** permission scoped to the target zone.

To avoid storing the token in the config file:

```bash
export CLOUDFLARE_API_TOKEN="your-token"
dddns -c config.yaml
```

### No-IP example

```yaml
providers:
  - type: no-ip
    zone_name: "example.com"
    # api_key: "your-api-key"  # optional if NOIP_API_KEY is set
    records:
      - name: "home.example.com"
        type: "A"
      - name: "vpn"              # relative to zone_name
        type: "A"
```

Create an API key in the [No-IP dashboard](https://www.noip.com/members/dns/records.php) with permission to edit DNS records in the target zone.

To avoid storing the key in the config file:

```bash
export NOIP_API_KEY="your-api-key"
dddns -c config.yaml
```

## Scheduling

Run `dddns` on a timer so your DNS stays updated when your ISP changes your IP.

**Timer mode (recommended for a single host):** set `update-interval` in the config and run dddns in the foreground or under a process supervisor:

```yaml
update-interval: "5m"
```

**Cron:** for one-shot mode, omit `update-interval` and schedule with cron:

```cron
*/5 * * * * /usr/local/bin/dddns -c /etc/dddns/config.yaml >> /var/log/dddns.log 2>&1
```

## Exit codes

- `0` — success (updated or unchanged), or timer mode stopped by signal
- `1` — error in one-shot mode (config, network, or provider API failure)


## Local Development Build

```bash
go build -o dddns .
```

Or:

```bash
./build.sh v1.0.0
```
