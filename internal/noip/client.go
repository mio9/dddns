package noip

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://api.noip.com/v1"

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type Rdata struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

type RRSet struct {
	Name    string  `json:"name"`
	DNSType string  `json:"dns_type"`
	Rdata   []Rdata `json:"rdata"`
}

type rrSetResponse struct {
	Data RRSet `json:"data"`
}

type errorBody struct {
	Errors []struct {
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

func (client *Client) GetRRSet(ctx context.Context, zoneName, name, dnsType string) (*RRSet, error) {
	requestPath := client.recordPath(zoneName, name, dnsType)

	var response rrSetResponse
	if err := client.do(ctx, http.MethodGet, requestPath, nil, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (client *Client) ReplaceRdata(ctx context.Context, zoneName, name, dnsType string, rdata []Rdata) error {
	requestPath := client.recordPath(zoneName, name, dnsType) + "/rdata"
	return client.do(ctx, http.MethodPut, requestPath, rdata, nil)
}

func (client *Client) recordPath(zoneName, name, dnsType string) string {
	segments := []string{
		"/dns/records",
		zoneName,
		name,
		"rrsets",
		dnsType,
	}
	escaped := make([]string, len(segments))
	for index, segment := range segments {
		escaped[index] = url.PathEscape(segment)
	}
	return strings.Join(escaped, "/")
}

func (client *Client) do(ctx context.Context, method, requestPath string, requestBody any, responseBody any) error {
	requestURL := strings.TrimSuffix(client.baseURL, "/") + requestPath

	var bodyReader io.Reader
	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Accept", "application/json")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, requestPath, err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if responseBody != nil && len(responseBytes) > 0 {
			if err := json.Unmarshal(responseBytes, responseBody); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}
		return nil
	}

	return decodeAPIError(response.StatusCode, responseBytes)
}

func decodeAPIError(statusCode int, responseBytes []byte) error {
	var body errorBody
	if err := json.Unmarshal(responseBytes, &body); err == nil && len(body.Errors) > 0 {
		firstError := body.Errors[0]
		if firstError.Detail != "" {
			return fmt.Errorf("no-ip API error %d: %s (%s)", statusCode, firstError.Title, firstError.Detail)
		}
		return fmt.Errorf("no-ip API error %d: %s", statusCode, firstError.Title)
	}

	return fmt.Errorf("no-ip API error %d: %s", statusCode, strings.TrimSpace(string(responseBytes)))
}
