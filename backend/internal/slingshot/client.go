package slingshot

import (
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
	DID    string `json:"did"`
	Handle string `json:"handle"`
	PDS    string `json:"pds"`
}

type Record struct {
	URI   string          `json:"uri"`
	CID   string          `json:"cid"`
	Value json.RawMessage `json:"value"`
}

func (c *Client) ResolveIdentity(ctx context.Context, identifier string) (*Identity, error) {
	params := url.Values{}
	params.Set("identifier", identifier)

	endpoint := fmt.Sprintf("%s/xrpc/blue.microcosm.identity.resolveMiniDoc?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("identity not found: %s", identifier)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &identity, nil
}

func (c *Client) GetRecord(ctx context.Context, uri string) (*Record, error) {
	params := url.Values{}
	params.Set("at_uri", uri)

	endpoint := fmt.Sprintf("%s/xrpc/blue.microcosm.repo.getRecordByUri?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("record not found: %s", uri)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var record Record
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &record, nil
}

func (c *Client) GetRecordByParts(ctx context.Context, repo, collection, rkey string) (*Record, error) {
	uri := fmt.Sprintf("at://%s/%s/%s", repo, collection, rkey)
	return c.GetRecord(ctx, uri)
}

type ListRecordsResponse struct {
	Records []Record `json:"records"`
	Cursor  string   `json:"cursor,omitempty"`
}

func (c *Client) ListRecords(ctx context.Context, repo, collection string, limit int, cursor string) (*ListRecordsResponse, error) {
	params := url.Values{}
	params.Set("repo", repo)
	params.Set("collection", collection)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	endpoint := fmt.Sprintf("%s/records?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var listResp ListRecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &listResp, nil
}

func (c *Client) ResolveDID(ctx context.Context, did string) (string, error) {
	identity, err := c.ResolveIdentity(ctx, did)
	if err != nil {
		return "", err
	}
	return identity.PDS, nil
}

func (c *Client) ResolveHandle(ctx context.Context, handle string) (string, error) {
	identity, err := c.ResolveIdentity(ctx, handle)
	if err != nil {
		return "", err
	}
	return identity.DID, nil
}
