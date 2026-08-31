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

type pendingUpdate struct {
	found    *dns.RecordResponse
	publicIP string
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
	var findErrors []error
	var patches []dns.BatchPatchUnionParam
	var pendingUpdates []pendingUpdate

	for _, record := range provider.records {
		found, err := provider.findRecord(ctx, record)
		if err != nil {
			findErrors = append(findErrors, err)
			continue
		}

		if found.Content == publicIP {
			fmt.Printf("unchanged: %s already points to %s\n", found.Name, publicIP)
			continue
		}

		patch, err := batchPatchForRecord(found, publicIP)
		if err != nil {
			findErrors = append(findErrors, err)
			continue
		}

		patches = append(patches, patch)
		pendingUpdates = append(pendingUpdates, pendingUpdate{
			found:    found,
			publicIP: publicIP,
		})
	}

	if len(patches) == 0 {
		return errors.Join(findErrors...)
	}

	_, err := provider.client.DNS.Records.Batch(ctx, dns.RecordBatchParams{
		ZoneID:  cloudflare.F(provider.zoneID),
		Patches: cloudflare.F(patches),
	})
	if err != nil {
		findErrors = append(findErrors, fmt.Errorf("batch update DNS records: %w", err))
		return errors.Join(findErrors...)
	}

	for _, update := range pendingUpdates {
		fmt.Printf("updated: %s %s -> %s\n", update.found.Name, update.found.Content, update.publicIP)
	}

	return errors.Join(findErrors...)
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

func batchPatchForRecord(found *dns.RecordResponse, publicIP string) (dns.BatchPatchUnionParam, error) {
	switch string(found.Type) {
	case "AAAA":
		return dns.BatchPatchAAAARecordParam{
			ID: cloudflare.F(found.ID),
			AAAARecordParam: dns.AAAARecordParam{
				Content: cloudflare.F(publicIP),
			},
		}, nil
	case "A":
		return dns.BatchPatchARecordParam{
			ID: cloudflare.F(found.ID),
			ARecordParam: dns.ARecordParam{
				Content: cloudflare.F(publicIP),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported record type %q for %s", found.Type, found.Name)
	}
}
