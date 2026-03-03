package xrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"margin.at/internal/config"
	"margin.at/internal/logger"
	"margin.at/internal/slingshot"
)

var SlingshotClient = slingshot.NewClient()

var (
	didPattern  = regexp.MustCompile(`^did:[a-z]+:[a-zA-Z0-9._:%-]+$`)
	nsidPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(\.[a-zA-Z][a-zA-Z0-9]*)+$`)
	rkeyPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

type ATURI struct {
	DID        string
	Collection string
	RKey       string
}

func ParseATURI(uri string) (*ATURI, error) {
	if !strings.HasPrefix(uri, "at://") {
		return nil, fmt.Errorf("invalid AT-URI: must start with at://")
	}

	path := strings.TrimPrefix(uri, "at://")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		return nil, fmt.Errorf("invalid AT-URI: missing DID authority")
	}

	did := parts[0]
	if !didPattern.MatchString(did) {
		return nil, fmt.Errorf("invalid AT-URI: malformed DID %q", did)
	}

	result := &ATURI{DID: did}

	if len(parts) >= 2 && parts[1] != "" {
		collection := parts[1]
		if !nsidPattern.MatchString(collection) {
			return nil, fmt.Errorf("invalid AT-URI: malformed collection NSID %q", collection)
		}
		result.Collection = collection
	}

	if len(parts) >= 3 && parts[2] != "" {
		rkey := parts[2]
		if !rkeyPattern.MatchString(rkey) || strings.HasPrefix(rkey, ".") || strings.HasSuffix(rkey, ".") {
			return nil, fmt.Errorf("invalid AT-URI: malformed record key %q", rkey)
		}
		if len(rkey) > 512 {
			return nil, fmt.Errorf("invalid AT-URI: record key too long (max 512)")
		}
		result.RKey = rkey
	}

	if len(parts) > 3 {
		return nil, fmt.Errorf("invalid AT-URI: too many path segments")
	}

	return result, nil
}

func (a *ATURI) String() string {
	if a.Collection == "" {
		return fmt.Sprintf("at://%s", a.DID)
	}
	if a.RKey == "" {
		return fmt.Sprintf("at://%s/%s", a.DID, a.Collection)
	}
	return fmt.Sprintf("at://%s/%s/%s", a.DID, a.Collection, a.RKey)
}

func init() {
	logger.Info("Slingshot client initialized: %s", slingshot.DefaultBaseURL)
}

func ResolveDIDToPDS(did string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pds, err := SlingshotClient.ResolveDID(ctx, did); err == nil && pds != "" {
		return pds, nil
	}

	return resolveDIDToPDSDirect(did)
}

func resolveDIDToPDSDirect(did string) (string, error) {
	var docURL string
	if strings.HasPrefix(did, "did:plc:") {
		docURL = config.Get().PLCResolveURL(did)
	} else if strings.HasPrefix(did, "did:web:") {
		domain := strings.TrimPrefix(did, "did:web:")
		docURL = fmt.Sprintf("https://%s/.well-known/did.json", domain)
	} else {
		return "", nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(docURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch DID doc: %d", resp.StatusCode)
	}

	var doc struct {
		Service []struct {
			ID              string `json:"id"`
			Type            string `json:"type"`
			ServiceEndpoint string `json:"serviceEndpoint"`
		} `json:"service"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", err
	}

	for _, svc := range doc.Service {
		if svc.ID == "#atproto_pds" && svc.Type == "AtprotoPersonalDataServer" {
			return svc.ServiceEndpoint, nil
		}
	}
	for _, svc := range doc.Service {
		if svc.Type == "AtprotoPersonalDataServer" {
			return svc.ServiceEndpoint, nil
		}
	}
	return "", nil
}

func ResolveHandle(handle string) (string, error) {
	if strings.HasPrefix(handle, "did:") {
		return handle, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if did, err := SlingshotClient.ResolveHandle(ctx, handle); err == nil && did != "" {
		return did, nil
	}

	return resolveHandleDirect(handle)
}

func resolveHandleDirect(handle string) (string, error) {
	url := config.Get().BskyResolveHandleURL(handle)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to resolve handle: %d", resp.StatusCode)
	}

	var result struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.DID, nil
}
