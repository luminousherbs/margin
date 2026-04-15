package db

import (
	"database/sql"
	"embed"
	"time"

	"github.com/pressly/goose/v3"
)

//go:embed migrations
var migrationsFS embed.FS

func (db *DB) Migrate() error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db.DB, "migrations")
}

func (db *DB) GetCursor(id string) (int64, error) {
	var cursor int64
	err := db.QueryRow("SELECT last_cursor FROM cursors WHERE id = $1", id).Scan(&cursor)
	switch err {
	case nil:
		return cursor, nil
	case sql.ErrNoRows:
		return 0, nil
	default:
		return 0, err
	}
}

func (db *DB) SetCursor(id string, cursor int64) error {
	_, err := db.Exec(`
		INSERT INTO cursors (id, last_cursor, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT(id) DO UPDATE SET
			last_cursor = EXCLUDED.last_cursor,
			updated_at = EXCLUDED.updated_at
	`, id, cursor, time.Now())
	return err
}
