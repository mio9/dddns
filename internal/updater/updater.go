package updater

import (
	"context"
	"errors"
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

func findRecord(ctx context.Context, client *cloudflare.Client, zoneID string, record config.Record) (*dns.RecordResponse, error) {
	if record.ID != "" {
		found, err := client.DNS.Records.Get(ctx, record.ID, dns.RecordGetParams{
			ZoneID: cloudflare.F(zoneID),
		})
		if err != nil {
			return nil, fmt.Errorf("get DNS record: %w", err)
		}
		return found, nil
	}

	recordType := dns.RecordListParamsType(record.Type)
	records, err := client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(record.Name),
		}),
		Type: cloudflare.F(recordType),
	})
	if err != nil {
		return nil, fmt.Errorf("list DNS records: %w", err)
	}

	if len(records.Result) == 0 {
		return nil, fmt.Errorf("no %s record found for %q", record.Type, record.Name)
	}
	if len(records.Result) > 1 {
		return nil, fmt.Errorf("multiple %s records found for %q", record.Type, record.Name)
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

func updateRecord(ctx context.Context, client *cloudflare.Client, zoneID string, record config.Record, publicIP string) error {
	found, err := findRecord(ctx, client, zoneID, record)
	if err != nil {
		return err
	}

	if found.Content == publicIP {
		fmt.Printf("unchanged: %s already points to %s\n", found.Name, publicIP)
		return nil
	}

	updated, err := client.DNS.Records.Edit(ctx, found.ID, dns.RecordEditParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   editRecordBody(string(found.Type), publicIP),
	})
	if err != nil {
		return fmt.Errorf("update DNS record: %w", err)
	}

	fmt.Printf("updated: %s %s -> %s\n", updated.Name, found.Content, updated.Content)
	return nil
}

func Update(ctx context.Context, cfg *config.Config) error {
	publicIP, err := getPublicIP(ctx, cfg.IPProvider.URL)
	if err != nil {
		return err
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(cfg.Cloudflare.APIToken),
	)

	var updateErrors []error
	for _, record := range cfg.Records {
		if err := updateRecord(ctx, client, cfg.Cloudflare.ZoneID, record, publicIP); err != nil {
			updateErrors = append(updateErrors, err)
		}
	}

	return errors.Join(updateErrors...)
}
