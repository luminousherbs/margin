package db

import (
	"time"
)

func (db *DB) SaveSession(id, did, handle, accessToken, refreshToken, dpopKey string, expiresAt time.Time) error {
	_, err := db.Exec(db.Rebind(`
		INSERT INTO sessions (id, did, handle, access_token, refresh_token, dpop_key, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			dpop_key = excluded.dpop_key,
			expires_at = excluded.expires_at
	`), id, did, handle, accessToken, refreshToken, dpopKey, time.Now(), expiresAt)
	return err
}

func (db *DB) GetSession(id string) (did, handle, accessToken, refreshToken, dpopKey string, err error) {
	err = db.QueryRow(db.Rebind(`
		SELECT did, handle, access_token, refresh_token, COALESCE(dpop_key, '')
		FROM sessions
		WHERE id = ? AND expires_at > ?
	`), id, time.Now()).Scan(&did, &handle, &accessToken, &refreshToken, &dpopKey)
	return
}

func (db *DB) DeleteSession(id string) error {
	_, err := db.Exec(db.Rebind(`DELETE FROM sessions WHERE id = ?`), id)
	return err
}

func (db *DB) DeleteExpiredSessions() error {
	_, err := db.Exec(db.Rebind(`DELETE FROM sessions WHERE expires_at <= ?`), time.Now())
	return err
}
