# dddns

Small CLI that keeps Cloudflare DNS records in sync with your current public IP. Useful for home servers, NAS devices, or any host behind a dynamic IP.

## Requirements

- Go 1.26 or later (to build from source)
- A Cloudflare account with a zone and DNS record to update
- A Cloudflare API token with permission to edit DNS records in that zone

## Build

```bash
go build -o dddns ./cmd/dddns
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
cloudflare:
  zone_id: "your-zone-id"
  # api_token: "your-api-token"  # optional if CLOUDFLARE_API_TOKEN is set

records:
  - name: "home.example.com"
    type: "A"   # optional, defaults to A
  - name: "vpn.example.com"

ip-provider:
  url: "https://api.ipify.org"  # optional
```

### Config reference

| Field | Required | Description |
|-------|----------|-------------|
| `cloudflare.zone_id` | yes | Cloudflare zone ID for the domain |
| `cloudflare.api_token` | yes* | API token with DNS edit access for the zone |
| `records` | yes | List of DNS records to update |
| `records[].name` | yes** | Full DNS record name (e.g. `home.example.com`) |
| `records[].id` | yes** | Cloudflare DNS record ID (alternative to `records[].name`) |
| `records[].type` | no | Record type to match when using `records[].name`. Default: `A`. Use `AAAA` for IPv6 |
| `ip-provider.url` | no | URL that returns your public IP as plain text. Default: `https://api.ipify.org` |

\* Either set `cloudflare.api_token` in the config file or export the `CLOUDFLARE_API_TOKEN` environment variable.

\** Each record needs `records[].id` **or** `records[].name` (with optional `records[].type`). If you use `records[].name`, the record must exist already in Cloudflare.

### Example with record ID

If you already know the DNS record ID, you can skip the name lookup:

```yaml
cloudflare:
  zone_id: "your-zone-id"

records:
  - id: "dns-record-id"
```

### API token

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
- `1` — error (config, network, or Cloudflare API failure)
