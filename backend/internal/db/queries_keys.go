package db

import (
	"time"
)

func (db *DB) CreateAPIKey(key *APIKey) error {
	_, err := db.Exec(`
		INSERT INTO api_keys (id, owner_did, name, key_hash, created_at, uri, cid)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			key_hash = EXCLUDED.key_hash,
			uri = EXCLUDED.uri,
			cid = EXCLUDED.cid
	`, key.ID, key.OwnerDID, key.Name, key.KeyHash, key.CreatedAt, key.URI, key.CID)
	return err
}

func (db *DB) GetAPIKeysByOwner(ownerDID string) ([]APIKey, error) {
	rows, err := db.Query(`
		SELECT id, owner_did, name, key_hash, created_at, last_used_at
		FROM api_keys
		WHERE owner_did = $1
		ORDER BY created_at DESC
	`, ownerDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.OwnerDID, &k.Name, &k.KeyHash, &k.CreatedAt, &k.LastUsedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (db *DB) GetAPIKeyByHash(keyHash string) (*APIKey, error) {
	var k APIKey
	err := db.QueryRow(`
		SELECT id, owner_did, name, key_hash, created_at, last_used_at
		FROM api_keys
		WHERE key_hash = $1
	`, keyHash).Scan(&k.ID, &k.OwnerDID, &k.Name, &k.KeyHash, &k.CreatedAt, &k.LastUsedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (db *DB) UpdateAPIKeyLastUsed(id string) error {
	_, err := db.Exec(`UPDATE api_keys SET last_used_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}
