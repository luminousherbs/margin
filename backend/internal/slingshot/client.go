package slingshot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL = "https://slingshot.microcosm.blue"
	DefaultTimeout = 5 * time.Second
	UserAgent      = "Margin (margin.at)"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func NewClientWithURL(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

type Identity struct {
	DID       string `json:"did"`
	Handle    string `json:"handle"`
	PDS       string `json:"pds"`
	SigningKey string `json:"signing_key"`
}

type Record struct {
	URI   string          `json:"uri"`
	CID   string          `json:"cid"`
	Value json.RawMessage `json:"value"`
}

type HydrationSource struct {
	Path  string `json:"path"`
	Shape string `json:"shape"`
}

type HydratePayload struct {
	XRPC                   string            `json:"xrpc"`
	AtprotoProxy           string            `json:"atproto_proxy"`
	Authorization          string            `json:"authorization,omitempty"`
	AtprotoAcceptLabelers  string            `json:"atproto_accept_labelers,omitempty"`
	Params                 any               `json:"params,omitempty"`
	HydrationSources       []HydrationSource `json:"hydration_sources"`
}

type HydrationResult struct {
	Status      string          `json:"status"`
	URI         string          `json:"uri,omitempty"`
	CID         string          `json:"cid,omitempty"`
	Value       json.RawMessage `json:"value,omitempty"`
	FollowUp    string          `json:"followUp,omitempty"`
	Reason      string          `json:"reason,omitempty"`
	ShouldRetry bool            `json:"shouldRetry,omitempty"`
}

type HydrateResponse struct {
	Output      json.RawMessage            `json:"output"`
	Records     map[string]HydrationResult `json:"records"`
	Identifiers map[string]HydrationResult `json:"identifiers"`
}

func (c *Client) get(ctx context.Context, endpoint string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		var xrpcErr struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if jsonErr := json.NewDecoder(resp.Body).Decode(&xrpcErr); jsonErr == nil && xrpcErr.Error != "" {
			return fmt.Errorf("%s: %s", xrpcErr.Error, xrpcErr.Message)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) ResolveIdentity(ctx context.Context, identifier string) (*Identity, error) {
	params := url.Values{}
	params.Set("identifier", identifier)
	endpoint := fmt.Sprintf("%s/xrpc/blue.microcosm.identity.resolveMiniDoc?%s", c.baseURL, params.Encode())

	var identity Identity
	if err := c.get(ctx, endpoint, &identity); err != nil {
		return nil, fmt.Errorf("identity not found for %s: %w", identifier, err)
	}
	return &identity, nil
}

func (c *Client) ResolveHandle(ctx context.Context, handle string) (string, error) {
	identity, err := c.ResolveIdentity(ctx, handle)
	if err != nil {
		return "", err
	}
	return identity.DID, nil
}

func (c *Client) ResolveDID(ctx context.Context, did string) (string, error) {
	identity, err := c.ResolveIdentity(ctx, did)
	if err != nil {
		return "", err
	}
	return identity.PDS, nil
}

func (c *Client) ResolveService(ctx context.Context, did, id, serviceType string) (string, error) {
	params := url.Values{}
	params.Set("did", did)
	params.Set("id", id)
	if serviceType != "" {
		params.Set("type", serviceType)
	}
	endpoint := fmt.Sprintf("%s/xrpc/com.bad-example.identity.resolveService?%s", c.baseURL, params.Encode())

	var result struct {
		Endpoint string `json:"endpoint"`
	}
	if err := c.get(ctx, endpoint, &result); err != nil {
		return "", fmt.Errorf("service not resolved for %s%s: %w", did, id, err)
	}
	return result.Endpoint, nil
}

func (c *Client) GetRecord(ctx context.Context, uri string) (*Record, error) {
	params := url.Values{}
	params.Set("at_uri", uri)
	endpoint := fmt.Sprintf("%s/xrpc/blue.microcosm.repo.getRecordByUri?%s", c.baseURL, params.Encode())

	var record Record
	if err := c.get(ctx, endpoint, &record); err != nil {
		return nil, fmt.Errorf("record not found %s: %w", uri, err)
	}
	return &record, nil
}

func (c *Client) GetRecordStandard(ctx context.Context, repo, collection, rkey string) (*Record, error) {
	params := url.Values{}
	params.Set("repo", repo)
	params.Set("collection", collection)
	params.Set("rkey", rkey)
	endpoint := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?%s", c.baseURL, params.Encode())

	var record Record
	if err := c.get(ctx, endpoint, &record); err != nil {
		return nil, fmt.Errorf("record not found %s/%s/%s: %w", repo, collection, rkey, err)
	}
	return &record, nil
}

func (c *Client) GetRecordByParts(ctx context.Context, repo, collection, rkey string) (*Record, error) {
	return c.GetRecord(ctx, fmt.Sprintf("at://%s/%s/%s", repo, collection, rkey))
}

func (c *Client) HydrateQueryResponse(ctx context.Context, payload HydratePayload) (*HydrateResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/xrpc/com.bad-example.proxy.hydrateQueryResponse", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var xrpcErr struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if jsonErr := json.NewDecoder(resp.Body).Decode(&xrpcErr); jsonErr == nil && xrpcErr.Error != "" {
			return nil, fmt.Errorf("%s: %s", xrpcErr.Error, xrpcErr.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result HydrateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
