package cloudflare

import (
	"context"
	"errors"
	"fmt"

	"mio9/dddns/internal/config"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
	"github.com/cloudflare/cloudflare-go/v5/option"
)

type Provider struct {
	client  *cloudflare.Client
	zoneID  string
	records []config.Record
}

func New(cfg config.Provider) (*Provider, error) {
	return &Provider{
		client: cloudflare.NewClient(
			option.WithAPIToken(cfg.APIToken),
		),
		zoneID:  cfg.ZoneID,
		records: cfg.Records,
	}, nil
}

func (provider *Provider) Update(ctx context.Context, publicIP string) error {
	var updateErrors []error
	for _, record := range provider.records {
		if err := provider.updateRecord(ctx, record, publicIP); err != nil {
			updateErrors = append(updateErrors, err)
		}
	}
	return errors.Join(updateErrors...)
}

func (provider *Provider) findRecord(ctx context.Context, record config.Record) (*dns.RecordResponse, error) {
	if record.ID != "" {
		found, err := provider.client.DNS.Records.Get(ctx, record.ID, dns.RecordGetParams{
			ZoneID: cloudflare.F(provider.zoneID),
		})
		if err != nil {
			return nil, fmt.Errorf("get DNS record: %w", err)
		}
		return found, nil
	}

	recordType := dns.RecordListParamsType(record.Type)
	records, err := provider.client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(provider.zoneID),
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

func (provider *Provider) updateRecord(ctx context.Context, record config.Record, publicIP string) error {
	found, err := provider.findRecord(ctx, record)
	if err != nil {
		return err
	}

	if found.Content == publicIP {
		fmt.Printf("unchanged: %s already points to %s\n", found.Name, publicIP)
		return nil
	}

	updated, err := provider.client.DNS.Records.Edit(ctx, found.ID, dns.RecordEditParams{
		ZoneID: cloudflare.F(provider.zoneID),
		Body:   editRecordBody(string(found.Type), publicIP),
	})
	if err != nil {
		return fmt.Errorf("update DNS record: %w", err)
	}

	fmt.Printf("updated: %s %s -> %s\n", updated.Name, found.Content, updated.Content)
	return nil
}
