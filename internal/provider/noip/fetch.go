package noip

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mio9/dddns/internal/config"
	noipclient "mio9/dddns/internal/noip"
)

type Fetcher struct {
	client   *noipclient.Client
	zoneName string
}

func NewFetcher(cfg config.Provider) (*Fetcher, error) {
	return &Fetcher{
		client:   noipclient.NewClient(cfg.APIKey),
		zoneName: cfg.ZoneName,
	}, nil
}

func (fetcher *Fetcher) FetchMatchingARecords(ctx context.Context, publicIP string) ([]config.Record, error) {
	names, err := fetcher.client.ListNamesInZone(ctx, fetcher.zoneName)
	if err != nil {
		return nil, err
	}

	var records []config.Record
	var fetchErrors []error
	for _, name := range names {
		rrset, err := fetcher.client.GetRRSet(ctx, fetcher.zoneName, name, "A")
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				continue
			}
			fetchErrors = append(fetchErrors, fmt.Errorf("get A record for %q: %w", name, err))
			continue
		}
		if !noipclient.RRSetHasValue(rrset, publicIP) {
			continue
		}
		records = append(records, config.Record{
			Name: noipclient.DisplayName(fetcher.zoneName, name),
			Type: "A",
		})
	}

	return records, errors.Join(fetchErrors...)
}
