package db

import (
	"time"
)

func (db *DB) SaveSession(id, did, handle, accessToken, refreshToken, dpopKey string, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO sessions (id, did, handle, access_token, refresh_token, dpop_key, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT(id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			dpop_key = EXCLUDED.dpop_key,
			expires_at = EXCLUDED.expires_at
	`, id, did, handle, accessToken, refreshToken, dpopKey, time.Now(), expiresAt)
	return err
}

func (db *DB) GetSession(id string) (did, handle, accessToken, refreshToken, dpopKey string, err error) {
	err = db.QueryRow(`
		SELECT did, handle, access_token, refresh_token, COALESCE(dpop_key, '')
		FROM sessions
		WHERE id = $1 AND expires_at > $2
	`, id, time.Now()).Scan(&did, &handle, &accessToken, &refreshToken, &dpopKey)
	return
}

func (db *DB) DeleteSession(id string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (db *DB) DeleteExpiredSessions() error {
	_, err := db.Exec(`DELETE FROM sessions WHERE expires_at <= $1`, time.Now())
	return err
}
