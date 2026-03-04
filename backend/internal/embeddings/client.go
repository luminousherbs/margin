package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"margin.at/internal/logger"
)

const (
	Model          = "text-embedding-3-small"
	Dimensions     = 1536
	MaxTokens      = 8191
	MaxInputChars  = 8000
	BatchSize      = 64
	openAIEndpoint = "https://api.openai.com/v1/embeddings"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	mu         sync.Mutex
}

type embeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type embeddingResponse struct {
	Data  []embeddingData `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type embeddingData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

func NewClient() *Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Info("OPENAI_API_KEY not set — embedding generation will be disabled")
	}
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) IsEnabled() bool {
	return c.apiKey != ""
}

func (c *Client) Embed(text string) ([]float32, error) {
	results, err := c.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return results[0], nil
}

func (c *Client) EmbedBatch(texts []string) ([][]float32, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	truncated := make([]string, len(texts))
	for i, t := range texts {
		t = truncateText(t, MaxInputChars)
		if strings.TrimSpace(t) == "" {
			t = " "
		}
		truncated[i] = t
	}

	results := make([][]float32, len(texts))

	for start := 0; start < len(truncated); start += BatchSize {
		end := start + BatchSize
		if end > len(truncated) {
			end = len(truncated)
		}
		batch := truncated[start:end]

		embeddings, err := c.callAPI(batch)
		if err != nil {
			return nil, fmt.Errorf("embedding batch %d-%d failed: %w", start, end, err)
		}

		for _, emb := range embeddings {
			idx := start + emb.Index
			if idx < len(results) {
				results[idx] = emb.Embedding
			}
		}
	}

	return results, nil
}

func (c *Client) callAPI(inputs []string) ([]embeddingData, error) {
	reqBody := embeddingRequest{
		Model: Model,
		Input: inputs,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openAIEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result embeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s", result.Error.Message)
	}

	return result.Data, nil
}

func truncateText(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars]
}

func BuildAnnotationText(bodyValue, selectorJSON, targetTitle, tagsJSON *string) string {
	var parts []string

	if selectorJSON != nil && *selectorJSON != "" {
		var selector struct {
			Exact  string `json:"exact"`
			Prefix string `json:"prefix"`
			Suffix string `json:"suffix"`
		}
		if err := json.Unmarshal([]byte(*selectorJSON), &selector); err == nil && selector.Exact != "" {
			parts = append(parts, selector.Exact)
		}
	}

	if bodyValue != nil && *bodyValue != "" {
		parts = append(parts, *bodyValue)
	}

	if targetTitle != nil && *targetTitle != "" {
		parts = append(parts, *targetTitle)
	}

	if tagsJSON != nil && *tagsJSON != "" {
		var tags []string
		if err := json.Unmarshal([]byte(*tagsJSON), &tags); err == nil && len(tags) > 0 {
			parts = append(parts, strings.Join(tags, ", "))
		}
	}

	return strings.Join(parts, " | ")
}

func BuildDocumentText(title, description, textContent string, tags []string) string {
	var parts []string

	parts = append(parts, title)

	if len(tags) > 0 {
		parts = append(parts, strings.Join(tags, ", "))
	}

	if textContent != "" {
		parts = append(parts, textContent)
	} else if description != "" {
		parts = append(parts, description)
	}

	return strings.Join(parts, " | ")
}
