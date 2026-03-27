package constellation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://constellation.microcosm.blue"
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

type CountResponse struct {
	Total int `json:"total"`
}

type Link struct {
	URI        string `json:"uri"`
	Collection string `json:"collection"`
	DID        string `json:"did"`
	Path       string `json:"path"`
}

type LinksResponse struct {
	Links  []Link `json:"links"`
	Cursor string `json:"cursor,omitempty"`
}

func (c *Client) getBacklinksCount(ctx context.Context, subject, source string) (int, error) {
	params := url.Values{}
	params.Set("subject", subject)
	params.Set("source", source)

	endpoint := fmt.Sprintf("%s/xrpc/blue.microcosm.links.getBacklinksCount?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var countResp CountResponse
	if err := json.NewDecoder(resp.Body).Decode(&countResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return countResp.Total, nil
}

type BacklinksResponse struct {
	Backlinks []struct {
		URI string `json:"uri"`
		DID string `json:"did"`
	} `json:"backlinks"`
	Cursor string `json:"cursor,omitempty"`
}

func (c *Client) getBacklinks(ctx context.Context, subject, source string, limit int) (*BacklinksResponse, error) {
	params := url.Values{}
	params.Set("subject", subject)
	params.Set("source", source)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	endpoint := fmt.Sprintf("%s/xrpc/blue.microcosm.links.getBacklinks?%s", c.baseURL, params.Encode())

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

	var result BacklinksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetLikeCount(ctx context.Context, subjectURI string) (int, error) {
	return c.getBacklinksCount(ctx, subjectURI, "at.margin.like:subject.uri")
}

func (c *Client) GetReplyCount(ctx context.Context, rootURI string) (int, error) {
	return c.getBacklinksCount(ctx, rootURI, "at.margin.reply:root.uri")
}

type CountsResult struct {
	LikeCount  int
	ReplyCount int
}

func (c *Client) GetCountsBatch(ctx context.Context, uris []string) (map[string]CountsResult, error) {
	if len(uris) == 0 {
		return map[string]CountsResult{}, nil
	}

	results := make(map[string]CountsResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 10)

	for _, uri := range uris {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			likeCount, _ := c.GetLikeCount(ctx, u)
			replyCount, _ := c.GetReplyCount(ctx, u)

			mu.Lock()
			results[u] = CountsResult{
				LikeCount:  likeCount,
				ReplyCount: replyCount,
			}
			mu.Unlock()
		}(uri)
	}

	wg.Wait()
	return results, nil
}

func (c *Client) GetAnnotationsForURL(ctx context.Context, targetURL string) ([]Link, error) {
	resp, err := c.getBacklinks(ctx, targetURL, "at.margin.annotation:target.source", 100)
	if err != nil {
		return nil, err
	}
	links := make([]Link, len(resp.Backlinks))
	for i, bl := range resp.Backlinks {
		links[i] = Link{URI: bl.URI, DID: bl.DID, Collection: "at.margin.annotation", Path: ".target.source"}
	}
	return links, nil
}

func (c *Client) GetHighlightsForURL(ctx context.Context, targetURL string) ([]Link, error) {
	resp, err := c.getBacklinks(ctx, targetURL, "at.margin.highlight:target.source", 100)
	if err != nil {
		return nil, err
	}
	links := make([]Link, len(resp.Backlinks))
	for i, bl := range resp.Backlinks {
		links[i] = Link{URI: bl.URI, DID: bl.DID, Collection: "at.margin.highlight", Path: ".target.source"}
	}
	return links, nil
}

func (c *Client) GetBookmarksForURL(ctx context.Context, targetURL string) ([]Link, error) {
	resp, err := c.getBacklinks(ctx, targetURL, "at.margin.bookmark:source", 100)
	if err != nil {
		return nil, err
	}
	links := make([]Link, len(resp.Backlinks))
	for i, bl := range resp.Backlinks {
		links[i] = Link{URI: bl.URI, DID: bl.DID, Collection: "at.margin.bookmark", Path: ".source"}
	}
	return links, nil
}

func (c *Client) GetAllItemsForURL(ctx context.Context, targetURL string) (annotations, highlights, bookmarks []Link, err error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	wg.Add(3)

	go func() {
		defer wg.Done()
		links, e := c.GetAnnotationsForURL(ctx, targetURL)
		mu.Lock()
		defer mu.Unlock()
		if e != nil {
			errs = append(errs, e)
		} else {
			annotations = links
		}
	}()

	go func() {
		defer wg.Done()
		links, e := c.GetHighlightsForURL(ctx, targetURL)
		mu.Lock()
		defer mu.Unlock()
		if e != nil {
			errs = append(errs, e)
		} else {
			highlights = links
		}
	}()

	go func() {
		defer wg.Done()
		links, e := c.GetBookmarksForURL(ctx, targetURL)
		mu.Lock()
		defer mu.Unlock()
		if e != nil {
			errs = append(errs, e)
		} else {
			bookmarks = links
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		return annotations, highlights, bookmarks, errs[0]
	}

	return annotations, highlights, bookmarks, nil
}

func (c *Client) GetLikers(ctx context.Context, subjectURI string) ([]string, error) {
	resp, err := c.getBacklinks(ctx, subjectURI, "at.margin.like:subject.uri", 100)
	if err != nil {
		return nil, err
	}
	dids := make([]string, len(resp.Backlinks))
	for i, bl := range resp.Backlinks {
		dids[i] = bl.DID
	}
	return dids, nil
}
