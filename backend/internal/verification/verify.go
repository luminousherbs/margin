package verification

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"margin.at/internal/logger"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

var linkTagPattern = regexp.MustCompile(`<link[^>]+rel=["']site\.standard\.document["'][^>]+href=["']([^"']+)["'][^>]*/?>|<link[^>]+href=["']([^"']+)["'][^>]+rel=["']site\.standard\.document["'][^>]*/?>`)

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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
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
	go func() {
		if err := VerifyPublication(pubURL, uri); err != nil {
			return
		}
		logger.Info("Publication verified: %s", uri)
		if onVerified != nil {
			onVerified(uri)
		}
	}()
}

func VerifyDocumentAsync(docURL, uri string, onVerified func(string)) {
	go func() {
		if err := VerifyDocument(docURL, uri); err != nil {
			return
		}
		logger.Info("Document verified: %s", uri)
		if onVerified != nil {
			onVerified(uri)
		}
	}()
}
