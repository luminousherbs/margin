package db

import (
	"fmt"
	"strings"
	"time"
)

func (db *DB) SaveEditHistory(uri, recordType, previousContent string, previousCID *string) error {
	_, err := db.Exec(`
		INSERT INTO edit_history (uri, record_type, previous_content, previous_cid, edited_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uri, recordType, previousContent, previousCID, time.Now())
	return err
}

func (db *DB) GetEditHistory(uri string) ([]EditHistory, error) {
	rows, err := db.Query(`
		SELECT id, uri, record_type, previous_content, previous_cid, edited_at
		FROM edit_history
		WHERE uri = $1
		ORDER BY edited_at DESC
	`, uri)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []EditHistory
	for rows.Next() {
		var h EditHistory
		if err := rows.Scan(&h.ID, &h.URI, &h.RecordType, &h.PreviousContent, &h.PreviousCID, &h.EditedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (db *DB) GetLatestEditTimes(uris []string) (map[string]time.Time, error) {
	if len(uris) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(uris))
	args := make([]interface{}, len(uris))
	for i, uri := range uris {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = uri
	}

	query := `
		SELECT uri, MAX(edited_at) as edited_at
		FROM edit_history
		WHERE uri IN (` + strings.Join(placeholders, ",") + `)
		GROUP BY uri
	`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var uri string
		var editedAt time.Time
		if err := rows.Scan(&uri, &editedAt); err != nil {
			continue
		}
		result[uri] = editedAt
	}

	return result, nil
}
