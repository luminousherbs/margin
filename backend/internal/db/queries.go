package db

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type EditHistory struct {
	ID              int       `json:"id"`
	URI             string    `json:"uri"`
	RecordType      string    `json:"recordType"`
	PreviousContent string    `json:"previousContent"`
	PreviousCID     *string   `json:"previousCid"`
	EditedAt        time.Time `json:"editedAt"`
}

func scanAnnotations(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]Annotation, error) {
	var annotations []Annotation
	for rows.Next() {
		var a Annotation
		if err := rows.Scan(&a.URI, &a.AuthorDID, &a.Motivation, &a.BodyValue, &a.BodyFormat, &a.BodyURI, &a.TargetSource, &a.TargetHash, &a.TargetTitle, &a.SelectorJSON, &a.TagsJSON, &a.CreatedAt, &a.IndexedAt, &a.CID); err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
	}
	return annotations, nil
}

func HashURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return HashString(rawURL)
	}

	host := strings.ToLower(parsed.Host)
	host = strings.TrimPrefix(host, "www.")

	normalized := host + parsed.Path
	if parsed.RawQuery != "" {
		normalized += "?" + parsed.RawQuery
	}
	normalized = strings.TrimSuffix(normalized, "/")

	return HashString(normalized)
}

func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (db *DB) GetAuthorByURI(uri string) (string, error) {
	var authorDID string
	err := db.QueryRow(`SELECT author_did FROM annotations WHERE uri = $1`, uri).Scan(&authorDID)
	if err == nil {
		return authorDID, nil
	}

	err = db.QueryRow(`SELECT author_did FROM highlights WHERE uri = $1`, uri).Scan(&authorDID)
	if err == nil {
		return authorDID, nil
	}

	err = db.QueryRow(`SELECT author_did FROM bookmarks WHERE uri = $1`, uri).Scan(&authorDID)
	if err == nil {
		return authorDID, nil
	}

	return "", fmt.Errorf("uri not found or no author")
}

func buildPlaceholders(n, startAt int) string {
	if n == 0 {
		return ""
	}
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", startAt+i)
	}
	return strings.Join(placeholders, ", ")
}
