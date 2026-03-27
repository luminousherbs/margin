package verification

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"margin.at/internal/logger"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

var linkTagPattern = regexp.MustCompile(`<link[^>]+rel=["']site\.standard\.document["'][^>]+href=["']([^"']+)["'][^>]*/?>|<link[^>]+href=["']([^"']+)["'][^>]+rel=["']site\.standard\.document["'][^>]*/?>`)

var (
	verifyQueue = make(chan verifyTask, 50)
	recentMu    sync.RWMutex
	recentURIs  = make(map[string]time.Time)
)

var (
	domainMu      sync.Mutex
	domainActive  = make(map[string]int)
	domainMaxConc = 1
)

var rateLimiter = make(chan struct{}, 2)

func init() {
	for i := 0; i < cap(rateLimiter); i++ {
		rateLimiter <- struct{}{}
	}
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for range ticker.C {
			select {
			case rateLimiter <- struct{}{}:
			default:
			}
		}
	}()

	for i := 0; i < 3; i++ {
		go verifyWorker()
	}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			recentMu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for uri, t := range recentURIs {
				if t.Before(cutoff) {
					delete(recentURIs, uri)
				}
			}
			recentMu.Unlock()

			domainMu.Lock()
			for d, c := range domainActive {
				if c <= 0 {
					delete(domainActive, d)
				}
			}
			domainMu.Unlock()
		}
	}()
}

type verifyTask struct {
	url        string
	uri        string
	onVerified func(string)
	isDoc      bool
}

func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Host
}

func acquireDomain(domain string) bool {
	if domain == "" {
		return true
	}
	domainMu.Lock()
	defer domainMu.Unlock()
	if domainActive[domain] >= domainMaxConc {
		return false
	}
	domainActive[domain]++
	return true
}

func releaseDomain(domain string) {
	if domain == "" {
		return
	}
	domainMu.Lock()
	domainActive[domain]--
	if domainActive[domain] <= 0 {
		delete(domainActive, domain)
	}
	domainMu.Unlock()
}

func verifyWorker() {
	for task := range verifyQueue {
		<-rateLimiter

		domain := extractDomain(task.url)

		if !acquireDomain(domain) {
			continue
		}

		var err error
		if task.isDoc {
			err = VerifyDocument(task.url, task.uri)
		} else {
			err = VerifyPublication(task.url, task.uri)
		}

		releaseDomain(domain)

		if err != nil {
			continue
		}
		kind := "Publication"
		if task.isDoc {
			kind = "Document"
		}
		logger.Info("%s verified: %s", kind, task.uri)
		if task.onVerified != nil {
			task.onVerified(task.uri)
		}
	}
}

func isDuplicate(uri string) bool {
	recentMu.RLock()
	_, exists := recentURIs[uri]
	recentMu.RUnlock()
	if exists {
		return true
	}
	recentMu.Lock()
	recentURIs[uri] = time.Now()
	recentMu.Unlock()
	return false
}

func VerifyPublication(pubURL, expectedURI string) error {
	pubURL = strings.TrimRight(pubURL, "/")

	parsed, err := url.Parse(pubURL)
	if err != nil {
		return fmt.Errorf("invalid publication URL: %w", err)
	}

	wellKnownPath := "/.well-known/site.standard.publication"
	if parsed.Path != "" && parsed.Path != "/" {
		wellKnownPath += parsed.Path
	}
	wellKnownURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, wellKnownPath)

	req, err := http.NewRequest("GET", wellKnownURL, nil)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	req.Header.Set("User-Agent", "Margin/1.0 (Standard.site verification)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", wellKnownURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("well-known endpoint returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	returnedURI := strings.TrimSpace(string(body))
	if returnedURI != expectedURI {
		return fmt.Errorf("URI mismatch: expected %s, got %s", expectedURI, returnedURI)
	}

	return nil
}

func VerifyDocument(docURL, expectedURI string) error {
	req, err := http.NewRequest("GET", docURL, nil)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	req.Header.Set("User-Agent", "Margin/1.0 (Standard.site verification)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", docURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("document URL returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return fmt.Errorf("failed to read document: %w", err)
	}

	html := string(body)

	matches := linkTagPattern.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		href := m[1]
		if href == "" {
			href = m[2]
		}
		if strings.TrimSpace(href) == expectedURI {
			return nil
		}
	}

	return fmt.Errorf("no matching <link rel=\"site.standard.document\"> tag found for %s", expectedURI)
}

func VerifyPublicationAsync(pubURL, uri string, onVerified func(string)) {
	if isDuplicate(uri) {
		return
	}
	select {
	case verifyQueue <- verifyTask{url: pubURL, uri: uri, onVerified: onVerified, isDoc: false}:
	default:
		// Queue full — drop silently to protect network
	}
}

func VerifyDocumentAsync(docURL, uri string, onVerified func(string)) {
	if isDuplicate(uri) {
		return
	}
	select {
	case verifyQueue <- verifyTask{url: docURL, uri: uri, onVerified: onVerified, isDoc: true}:
	default:
		// Queue full — drop silently to protect network
	}
}
