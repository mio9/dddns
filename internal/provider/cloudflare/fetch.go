package cloudflare

import (
	"context"
	"fmt"

	"mio9/dddns/internal/config"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
)

type Fetcher struct {
	client *Provider
}

func NewFetcher(cfg config.Provider) (*Fetcher, error) {
	provider, err := New(cfg)
	if err != nil {
		return nil, err
	}
	return &Fetcher{client: provider}, nil
}

func (fetcher *Fetcher) FetchMatchingARecords(ctx context.Context, publicIP string) ([]config.Record, error) {
	pager := fetcher.client.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(fetcher.client.zoneID),
		Type:   cloudflare.F(dns.RecordListParamsTypeA),
		Content: cloudflare.F(dns.RecordListParamsContent{
			Exact: cloudflare.F(publicIP),
		}),
	})

	var records []config.Record
	for pager.Next() {
		found := pager.Current()
		records = append(records, config.Record{
			ID:   found.ID,
			Name: found.Name,
			Type: "A",
		})
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("list DNS records: %w", err)
	}

	return records, nil
}
