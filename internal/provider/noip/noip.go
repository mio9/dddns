package noip

import (
	"context"
	"errors"
	"fmt"

	"mio9/dddns/internal/config"
	noipclient "mio9/dddns/internal/noip"
)

type Provider struct {
	client   *noipclient.Client
	zoneName string
	records  []config.Record
}

func New(cfg config.Provider) (*Provider, error) {
	return &Provider{
		client:   noipclient.NewClient(cfg.APIKey),
		zoneName: cfg.ZoneName,
		records:  cfg.Records,
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

func (provider *Provider) updateRecord(ctx context.Context, record config.Record, publicIP string) error {
	relativeName, err := noipclient.RelativeRecordName(provider.zoneName, recordName(record))
	if err != nil {
		return err
	}

	displayName := noipclient.DisplayName(provider.zoneName, relativeName)
	rrset, err := provider.client.GetRRSet(ctx, provider.zoneName, relativeName, record.Type)
	if err != nil {
		return fmt.Errorf("get DNS record %s: %w", displayName, err)
	}

	if noipclient.RRSetHasValue(rrset, publicIP) {
		fmt.Printf("unchanged: %s already points to %s\n", displayName, publicIP)
		return nil
	}

	previousValue := noipclient.CurrentValue(rrset)
	if err := provider.client.ReplaceRdata(
		ctx,
		provider.zoneName,
		relativeName,
		record.Type,
		noipclient.ReplaceRdataValues(rrset, publicIP),
	); err != nil {
		return fmt.Errorf("update DNS record %s: %w", displayName, err)
	}

	fmt.Printf("updated: %s %s -> %s\n", displayName, previousValue, publicIP)
	return nil
}

func recordName(record config.Record) string {
	if record.Name != "" {
		return record.Name
	}
	return record.ID
}
