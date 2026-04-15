package postgres

import (
	"context"
	"time"
)

type SessionRepository struct {
	db DB
}

func NewSessionRepository(db DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) GetSession(ctx context.Context, id string) (did, handle, accessToken, refreshToken, dpopKey string, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT did, handle, access_token, refresh_token, COALESCE(dpop_key, '')
		FROM sessions
		WHERE id = $1 AND expires_at > $2
	`, id, time.Now()).Scan(&did, &handle, &accessToken, &refreshToken, &dpopKey)
	return
}
