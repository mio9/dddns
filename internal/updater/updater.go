package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"mio9/dddns/internal/config"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
	"github.com/cloudflare/cloudflare-go/v5/option"
)

func getPublicIP(ctx context.Context, checkURL string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return "", fmt.Errorf("create IP check request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("fetch public IP: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch public IP: unexpected status %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read public IP response: %w", err)
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return "", fmt.Errorf("public IP response empty")
	}

	return ip, nil
}

func findRecord(ctx context.Context, client *cloudflare.Client, cfg *config.Config) (*dns.RecordResponse, error) {
	if cfg.Record.ID != "" {
		record, err := client.DNS.Records.Get(ctx, cfg.Record.ID, dns.RecordGetParams{
			ZoneID: cloudflare.F(cfg.Cloudflare.ZoneID),
		})
		if err != nil {
			return nil, fmt.Errorf("get DNS record: %w", err)
		}
		return record, nil
	}

	recordType := dns.RecordListParamsType(cfg.Record.Type)
	records, err := client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(cfg.Cloudflare.ZoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(cfg.Record.Name),
		}),
		Type: cloudflare.F(recordType),
	})
	if err != nil {
		return nil, fmt.Errorf("list DNS records: %w", err)
	}

	if len(records.Result) == 0 {
		return nil, fmt.Errorf("no %s record found for %q", cfg.Record.Type, cfg.Record.Name)
	}
	if len(records.Result) > 1 {
		return nil, fmt.Errorf("multiple %s records found for %q", cfg.Record.Type, cfg.Record.Name)
	}

	return &records.Result[0], nil
}

func editRecordBody(recordType string, ip string) dns.RecordEditParamsBodyUnion {
	switch recordType {
	case "AAAA":
		return dns.AAAARecordParam{
			Content: cloudflare.F(ip),
		}
	default:
		return dns.ARecordParam{
			Content: cloudflare.F(ip),
		}
	}
}

func Update(ctx context.Context, cfg *config.Config) error {
	publicIP, err := getPublicIP(ctx, cfg.IPProvider.URL)
	if err != nil {
		return err
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(cfg.Cloudflare.APIToken),
	)

	record, err := findRecord(ctx, client, cfg)
	if err != nil {
		return err
	}

	if record.Content == publicIP {
		fmt.Printf("unchanged: %s already points to %s\n", record.Name, publicIP)
		return nil
	}

	updated, err := client.DNS.Records.Edit(ctx, record.ID, dns.RecordEditParams{
		ZoneID: cloudflare.F(cfg.Cloudflare.ZoneID),
		Body:   editRecordBody(cfg.Record.Type, publicIP),
	})
	if err != nil {
		return fmt.Errorf("update DNS record: %w", err)
	}

	fmt.Printf("updated: %s %s -> %s\n", updated.Name, record.Content, updated.Content)
	return nil
}
