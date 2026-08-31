# dddns

Small CLI that keeps DNS records in sync with your current public IP. Useful for home servers, NAS devices, or any host behind a dynamic IP.

## Requirements

- Go 1.26 or later (to build from source)
- A supported DNS provider account with records to update

## Supported providers

- **Cloudflare** — update existing DNS records via API token

## Build

```bash
go build -o dddns .
```

Or:

```bash
./build.sh
```

## Usage

```bash
dddns --config /path/to/config.yaml
# or
dddns -c /path/to/config.yaml
```

### Example output

```
updated: home.example.com 203.0.113.10 -> 203.0.113.42
```

When the IP is already correct:

```
unchanged: home.example.com already points to 203.0.113.42
```

## Configuration

Create a YAML file and pass it with `--config` / `-c`.

```yaml
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
| `ip-provider.url` | no | URL that returns your public IP as plain text. Default: `https://api.ipify.org` |
| `providers` | yes | List of DNS provider configurations |
| `providers[].type` | yes | Provider name. Supported: `cloudflare` |
| `providers[].zone_id` | yes* | Cloudflare zone ID |
| `providers[].api_token` | yes* | Cloudflare API token with DNS edit access |
| `providers[].records` | yes* | DNS records to update for this provider |
| `providers[].records[].name` | yes** | Full DNS record name (e.g. `home.example.com`) |
| `providers[].records[].id` | yes** | Cloudflare DNS record ID (alternative to `name`) |
| `providers[].records[].type` | no | Record type when using `name`. Default: `A`. Use `AAAA` for IPv6 |

\* Required for `cloudflare` providers.

\** Each record needs `id` **or** `name` (with optional `type`). Named records must already exist in the provider.

Public IP is fetched once per run and applied to all configured records across all providers.

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

## Scheduling

Run `dddns` on a timer so your DNS stays updated when your ISP changes your IP.

Example cron entry (every 5 minutes):

```cron
*/5 * * * * /usr/local/bin/dddns -c /etc/dddns/config.yaml >> /var/log/dddns.log 2>&1
```

## Exit codes

- `0` — success (updated or unchanged)
- `1` — error (config, network, or provider API failure)
