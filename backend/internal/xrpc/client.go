package xrpc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v4"
)

type Client struct {
	PDS         string
	AccessToken string
	DPoPKey     *ecdsa.PrivateKey
	DPoPNonce   string
}

func NewClient(pds, accessToken string, dpopKey *ecdsa.PrivateKey) *Client {
	return &Client{
		PDS:         pds,
		AccessToken: accessToken,
		DPoPKey:     dpopKey,
	}
}

func (c *Client) createDPoPProof(method, uri string) (string, error) {
	now := time.Now()
	jti := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, jti); err != nil {

		for i := range jti {
			jti[i] = byte(now.UnixNano() >> (i * 8))
		}
	}

	publicJWK := jose.JSONWebKey{
		Key:       &c.DPoPKey.PublicKey,
		Algorithm: string(jose.ES256),
	}

	ath := ""
	if c.AccessToken != "" {
		hash := sha256.Sum256([]byte(c.AccessToken))
		ath = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	claims := map[string]interface{}{
		"jti": base64.RawURLEncoding.EncodeToString(jti),
		"htm": method,
		"htu": uri,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	}
	if c.DPoPNonce != "" {
		claims["nonce"] = c.DPoPNonce
	}
	if ath != "" {
		claims["ath"] = ath
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: c.DPoPKey}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"typ": "dpop+jwt",
			"jwk": publicJWK,
		},
	})
	if err != nil {
		return "", err
	}

	claimsBytes, _ := json.Marshal(claims)
	sig, err := signer.Sign(claimsBytes)
	if err != nil {
		return "", err
	}

	return sig.CompactSerialize()
}

func (c *Client) Call(ctx context.Context, method, nsid string, input, output interface{}) error {
	url := fmt.Sprintf("%s/xrpc/%s", c.PDS, nsid)

	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		var reqBody io.Reader
		if input != nil {

			data, err := json.Marshal(input)
			if err != nil {
				return err
			}
			reqBody = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return err
		}

		if input != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		dpopProof, err := c.createDPoPProof(method, url)
		if err != nil {
			return fmt.Errorf("failed to create DPoP proof: %w", err)
		}

		req.Header.Set("Authorization", "DPoP "+c.AccessToken)
		req.Header.Set("DPoP", dpopProof)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if nonce := resp.Header.Get("DPoP-Nonce"); nonce != "" {
			c.DPoPNonce = nonce
		}

		if resp.StatusCode < 400 {
			if output != nil {
				return json.NewDecoder(resp.Body).Decode(output)
			}
			return nil
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		if resp.StatusCode == 401 && (bytes.Contains(bodyBytes, []byte("use_dpop_nonce")) || bytes.Contains(bodyBytes, []byte("UseDpopNonce"))) {
			continue
		}

		return fmt.Errorf("XRPC error %d: %s", resp.StatusCode, bodyStr)
	}

	return fmt.Errorf("XRPC failed after retries")
}

type CreateRecordInput struct {
	Repo       string      `json:"repo"`
	Collection string      `json:"collection"`
	RKey       string      `json:"rkey,omitempty"`
	Record     interface{} `json:"record"`
}

type CreateRecordOutput struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

func (c *Client) CreateRecord(ctx context.Context, repo, collection string, record interface{}) (*CreateRecordOutput, error) {
	input := CreateRecordInput{
		Repo:       repo,
		Collection: collection,
		Record:     record,
	}

	var output CreateRecordOutput
	err := c.Call(ctx, "POST", "com.atproto.repo.createRecord", input, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

type DeleteRecordInput struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	RKey       string `json:"rkey"`
}

func (c *Client) DeleteRecord(ctx context.Context, repo, collection, rkey string) error {
	input := DeleteRecordInput{
		Repo:       repo,
		Collection: collection,
		RKey:       rkey,
	}

	return c.Call(ctx, "POST", "com.atproto.repo.deleteRecord", input, nil)
}

func (c *Client) DeleteRecordByURI(ctx context.Context, uri string) error {
	parsed, err := ParseATURI(uri)
	if err != nil {
		return err
	}

	if parsed.Collection == "" || parsed.RKey == "" {
		return fmt.Errorf("invalid AT-URI: must include collection and rkey")
	}

	return c.DeleteRecord(ctx, parsed.DID, parsed.Collection, parsed.RKey)
}

type PutRecordInput struct {
	Repo       string      `json:"repo"`
	Collection string      `json:"collection"`
	RKey       string      `json:"rkey"`
	Record     interface{} `json:"record"`
}

type PutRecordOutput struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

func (c *Client) PutRecord(ctx context.Context, repo, collection, rkey string, record interface{}) (*PutRecordOutput, error) {
	input := PutRecordInput{
		Repo:       repo,
		Collection: collection,
		RKey:       rkey,
		Record:     record,
	}

	var output PutRecordOutput
	err := c.Call(ctx, "POST", "com.atproto.repo.putRecord", input, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

type GetRecordOutput struct {
	URI   string          `json:"uri"`
	CID   string          `json:"cid"`
	Value json.RawMessage `json:"value"`
}

func (c *Client) GetRecord(ctx context.Context, repo, collection, rkey string) (*GetRecordOutput, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		c.PDS, repo, collection, rkey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	dpopProof, err := c.createDPoPProof("GET", url)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "DPoP "+c.AccessToken)
	req.Header.Set("DPoP", dpopProof)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var output GetRecordOutput
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, err
	}

	return &output, nil
}

type ListRecordsRecord struct {
	URI   string          `json:"uri"`
	CID   string          `json:"cid"`
	Value json.RawMessage `json:"value"`
}

type ListRecordsOutput struct {
	Cursor  string              `json:"cursor"`
	Records []ListRecordsRecord `json:"records"`
}

func (c *Client) ListRecords(ctx context.Context, repo, collection string, limit int) (*ListRecordsOutput, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s&limit=%d",
		c.PDS, repo, collection, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	dpopProof, err := c.createDPoPProof("GET", url)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "DPoP "+c.AccessToken)
	req.Header.Set("DPoP", dpopProof)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var output ListRecordsOutput
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, err
	}

	return &output, nil
}

type ResolveHandleOutput struct {
	Did string `json:"did"`
}

func (c *Client) ResolveHandle(ctx context.Context, handle string) (string, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.identity.resolveHandle?handle=%s", c.PDS, handle)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("XRPC error %d", resp.StatusCode)
	}

	var output ResolveHandleOutput
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return "", err
	}

	return output.Did, nil
}

type UploadBlobOutput struct {
	Blob BlobRef `json:"blob"`
}

func (c *Client) UploadBlob(ctx context.Context, data []byte, contentType string) (*BlobRef, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.uploadBlob", c.PDS)

	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", contentType)

		dpopProof, err := c.createDPoPProof("POST", url)
		if err != nil {
			return nil, fmt.Errorf("failed to create DPoP proof: %w", err)
		}

		req.Header.Set("Authorization", "DPoP "+c.AccessToken)
		req.Header.Set("DPoP", dpopProof)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if nonce := resp.Header.Get("DPoP-Nonce"); nonce != "" {
			c.DPoPNonce = nonce
		}

		if resp.StatusCode < 400 {
			var output UploadBlobOutput
			if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
				return nil, err
			}
			return &output.Blob, nil
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 401 && (bytes.Contains(bodyBytes, []byte("use_dpop_nonce")) || bytes.Contains(bodyBytes, []byte("UseDpopNonce"))) {
			continue
		}

		return nil, fmt.Errorf("XRPC error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil, fmt.Errorf("upload blob failed after retries")
}
