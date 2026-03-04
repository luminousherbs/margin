package standardsite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	client = &http.Client{Timeout: 10 * time.Second}

	pubCache   = make(map[string]string)
	pubCacheMu sync.RWMutex
)

func ResolveCanonicalURL(site, path, canonicalURL string) string {
	if canonicalURL != "" && strings.HasPrefix(canonicalURL, "https://") {
		return canonicalURL
	}

	if strings.HasPrefix(site, "at://") {
		pubURL := resolvePublicationURL(site)
		if pubURL != "" {
			base := strings.TrimRight(pubURL, "/")
			if path != "" {
				return base + "/" + strings.TrimLeft(path, "/")
			}
			return base
		}
		return ""
	}

	base := strings.TrimRight(site, "/")
	if path != "" {
		return base + "/" + strings.TrimLeft(path, "/")
	}
	return base
}

func resolvePublicationURL(atURI string) string {
	pubCacheMu.RLock()
	if url, ok := pubCache[atURI]; ok {
		pubCacheMu.RUnlock()
		return url
	}
	pubCacheMu.RUnlock()

	parts := strings.SplitN(strings.TrimPrefix(atURI, "at://"), "/", 3)
	if len(parts) != 3 {
		return ""
	}
	did, collection, rkey := parts[0], parts[1], parts[2]

	pdsHost := resolvePDS(did)
	if pdsHost == "" {
		return ""
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pdsHost, did, collection, rkey)

	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ""
	}

	var result struct {
		Value struct {
			URL string `json:"url"`
		} `json:"value"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Value.URL == "" {
		return ""
	}

	pubCacheMu.Lock()
	pubCache[atURI] = result.Value.URL
	pubCacheMu.Unlock()

	return result.Value.URL
}

func resolvePDS(did string) string {
	var url string
	if strings.HasPrefix(did, "did:plc:") {
		url = "https://plc.directory/" + did
	} else if strings.HasPrefix(did, "did:web:") {
		domain := strings.TrimPrefix(did, "did:web:")
		url = "https://" + domain + "/.well-known/did.json"
	} else {
		return ""
	}

	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return ""
	}

	var doc struct {
		Service []struct {
			ID              string `json:"id"`
			Type            string `json:"type"`
			ServiceEndpoint string `json:"serviceEndpoint"`
		} `json:"service"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return ""
	}

	for _, svc := range doc.Service {
		if svc.ID == "#atproto_pds" && svc.Type == "AtprotoPersonalDataServer" {
			return strings.TrimRight(svc.ServiceEndpoint, "/")
		}
	}
	return ""
}
