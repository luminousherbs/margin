package db

import (
	"database/sql"
	"fmt"
	"strings"
)

func (db *DB) GetProfile(did string) (*Profile, error) {
	var p Profile
	err := db.QueryRow(
		`SELECT uri, author_did, display_name, avatar, bio, website, links_json, created_at, indexed_at
		 FROM profiles WHERE author_did = $1`, did,
	).Scan(&p.URI, &p.AuthorDID, &p.DisplayName, &p.Avatar, &p.Bio, &p.Website, &p.LinksJSON, &p.CreatedAt, &p.IndexedAt)
	switch err {
	case nil:
		return &p, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (db *DB) GetProfilesByDIDs(dids []string) (map[string]*Profile, error) {
	if len(dids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(dids))
	args := make([]interface{}, len(dids))
	for i, did := range dids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = did
	}

	rows, err := db.Query(
		`SELECT uri, author_did, display_name, bio, avatar, website, links_json, created_at, indexed_at
		 FROM profiles WHERE author_did IN (`+strings.Join(placeholders, ",")+`)`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make(map[string]*Profile)
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.URI, &p.AuthorDID, &p.DisplayName, &p.Bio, &p.Avatar, &p.Website, &p.LinksJSON, &p.CreatedAt, &p.IndexedAt); err != nil {
			continue
		}
		profiles[p.AuthorDID] = &p
	}
	return profiles, rows.Err()
}

func (db *DB) UpsertProfile(p *Profile) error {
	_, err := db.Exec(`
		INSERT INTO profiles (uri, author_did, display_name, avatar, bio, website, links_json, created_at, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(uri) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			avatar       = EXCLUDED.avatar,
			bio          = EXCLUDED.bio,
			website      = EXCLUDED.website,
			links_json   = EXCLUDED.links_json,
			indexed_at   = EXCLUDED.indexed_at
	`, p.URI, p.AuthorDID, p.DisplayName, p.Avatar, p.Bio, p.Website, p.LinksJSON, p.CreatedAt, p.IndexedAt)
	return err
}

func (db *DB) DeleteProfile(uri string) error {
	_, err := db.Exec("DELETE FROM profiles WHERE uri = $1", uri)
	return err
}

func (db *DB) GetPreferences(did string) (*Preferences, error) {
	var p Preferences
	err := db.QueryRow(
		`SELECT uri, author_did, external_link_skipped_hostnames, subscribed_labelers,
		        label_preferences, disable_external_link_warning, enable_community_bookmarks,
		        created_at, indexed_at, cid
		 FROM preferences WHERE author_did = $1`, did,
	).Scan(
		&p.URI, &p.AuthorDID, &p.ExternalLinkSkippedHostnames, &p.SubscribedLabelers,
		&p.LabelPreferences, &p.DisableExternalLinkWarning, &p.EnableCommunityBookmarks,
		&p.CreatedAt, &p.IndexedAt, &p.CID,
	)
	switch err {
	case nil:
		return &p, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (db *DB) UpsertPreferences(p *Preferences) error {
	_, err := db.Exec(`
		INSERT INTO preferences (
			uri, author_did, external_link_skipped_hostnames, subscribed_labelers,
			label_preferences, disable_external_link_warning, enable_community_bookmarks,
			created_at, indexed_at, cid
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT(uri) DO UPDATE SET
			external_link_skipped_hostnames = EXCLUDED.external_link_skipped_hostnames,
			subscribed_labelers             = EXCLUDED.subscribed_labelers,
			label_preferences               = EXCLUDED.label_preferences,
			disable_external_link_warning   = EXCLUDED.disable_external_link_warning,
			enable_community_bookmarks      = EXCLUDED.enable_community_bookmarks,
			indexed_at                      = EXCLUDED.indexed_at,
			cid                             = EXCLUDED.cid
	`, p.URI, p.AuthorDID, p.ExternalLinkSkippedHostnames, p.SubscribedLabelers,
		p.LabelPreferences, p.DisableExternalLinkWarning, p.EnableCommunityBookmarks,
		p.CreatedAt, p.IndexedAt, p.CID)
	return err
}

func (db *DB) DeletePreferences(uri string) error {
	_, err := db.Exec("DELETE FROM preferences WHERE uri = $1", uri)
	return err
}

func (db *DB) GetPreferenceURIs(did string) ([]string, error) {
	rows, err := db.Query(
		"SELECT uri FROM preferences WHERE author_did = $1 AND uri IS NOT NULL AND uri != ''", did,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var uris []string
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}
	return uris, rows.Err()
}

func (db *DB) DeleteAPIKey(id, ownerDID string) (string, error) {
	var uri string
	err := db.QueryRow("SELECT uri FROM api_keys WHERE id = $1 AND owner_did = $2", id, ownerDID).Scan(&uri)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	_, err = db.Exec("DELETE FROM api_keys WHERE id = $1 AND owner_did = $2", id, ownerDID)
	return uri, err
}

func (db *DB) DeleteAPIKeyByURI(uri string) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE uri = $1", uri)
	return err
}

func (db *DB) GetAPIKeyURIs(ownerDID string) ([]string, error) {
	rows, err := db.Query(
		"SELECT uri FROM api_keys WHERE owner_did = $1 AND uri IS NOT NULL AND uri != ''", ownerDID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var uris []string
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}
	return uris, rows.Err()
}
