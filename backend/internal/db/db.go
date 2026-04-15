package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"margin.at/internal/domain"
)

type DB struct {
	*sql.DB
}

type (
	Note             = domain.Note
	Annotation       = domain.Annotation
	Selector         = domain.Selector
	Highlight        = domain.Highlight
	Bookmark         = domain.Bookmark
	Reply            = domain.Reply
	Like             = domain.Like
	Collection       = domain.Collection
	CollectionItem   = domain.CollectionItem
	Notification     = domain.Notification
	APIKey           = domain.APIKey
	Profile          = domain.Profile
	Preferences      = domain.Preferences
	Block            = domain.Block
	Mute             = domain.Mute
	ModerationReport = domain.ModerationReport
	ModerationAction = domain.ModerationAction
	ContentLabel     = domain.ContentLabel
)

func New(dsn string) (*DB, error) {
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		return nil, fmt.Errorf("only PostgreSQL is supported; DSN must start with postgres:// or postgresql://")
	}

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{DB: sqlDB}, nil
}

func (db *DB) Close() error { return db.DB.Close() }

func ParseSelector(selectorJSON *string) (*Selector, error) {
	if selectorJSON == nil || *selectorJSON == "" {
		return nil, nil
	}
	var s Selector
	if err := json.Unmarshal([]byte(*selectorJSON), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func ParseTags(tagsJSON *string) ([]string, error) {
	if tagsJSON == nil || *tagsJSON == "" {
		return nil, nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(*tagsJSON), &tags); err != nil {
		return nil, err
	}
	return tags, nil
}
