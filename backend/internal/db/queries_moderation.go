package db

import (
	"fmt"
	"strings"
	"time"
)

func (db *DB) CreateBlock(actorDID, subjectDID string) error {
	_, err := db.Exec(`
		INSERT INTO blocks (actor_did, subject_did, created_at) VALUES ($1, $2, $3)
		ON CONFLICT(actor_did, subject_did) DO NOTHING
	`, actorDID, subjectDID, time.Now())
	return err
}

func (db *DB) DeleteBlock(actorDID, subjectDID string) error {
	_, err := db.Exec(`DELETE FROM blocks WHERE actor_did = $1 AND subject_did = $2`, actorDID, subjectDID)
	return err
}

func (db *DB) GetBlocks(actorDID string) ([]Block, error) {
	rows, err := db.Query(`SELECT id, actor_did, subject_did, created_at FROM blocks WHERE actor_did = $1 ORDER BY created_at DESC`, actorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var b Block
		if err := rows.Scan(&b.ID, &b.ActorDID, &b.SubjectDID, &b.CreatedAt); err != nil {
			continue
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (db *DB) IsBlocked(actorDID, subjectDID string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM blocks WHERE actor_did = $1 AND subject_did = $2)`, actorDID, subjectDID).Scan(&exists)
	return exists, err
}

func (db *DB) IsBlockedEither(did1, did2 string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM blocks WHERE (actor_did = $1 AND subject_did = $2) OR (actor_did = $2 AND subject_did = $1))`, did1, did2).Scan(&exists)
	return exists, err
}

func (db *DB) GetBlockedDIDs(actorDID string) ([]string, error) {
	rows, err := db.Query(`SELECT subject_did FROM blocks WHERE actor_did = $1`, actorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dids []string
	for rows.Next() {
		var did string
		if err := rows.Scan(&did); err != nil {
			continue
		}
		dids = append(dids, did)
	}
	return dids, nil
}

func (db *DB) GetBlockedByDIDs(actorDID string) ([]string, error) {
	rows, err := db.Query(`SELECT actor_did FROM blocks WHERE subject_did = $1`, actorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dids []string
	for rows.Next() {
		var did string
		if err := rows.Scan(&did); err != nil {
			continue
		}
		dids = append(dids, did)
	}
	return dids, nil
}

func (db *DB) CreateMute(actorDID, subjectDID string) error {
	_, err := db.Exec(`
		INSERT INTO mutes (actor_did, subject_did, created_at) VALUES ($1, $2, $3)
		ON CONFLICT(actor_did, subject_did) DO NOTHING
	`, actorDID, subjectDID, time.Now())
	return err
}

func (db *DB) DeleteMute(actorDID, subjectDID string) error {
	_, err := db.Exec(`DELETE FROM mutes WHERE actor_did = $1 AND subject_did = $2`, actorDID, subjectDID)
	return err
}

func (db *DB) GetMutes(actorDID string) ([]Mute, error) {
	rows, err := db.Query(`SELECT id, actor_did, subject_did, created_at FROM mutes WHERE actor_did = $1 ORDER BY created_at DESC`, actorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mutes []Mute
	for rows.Next() {
		var m Mute
		if err := rows.Scan(&m.ID, &m.ActorDID, &m.SubjectDID, &m.CreatedAt); err != nil {
			continue
		}
		mutes = append(mutes, m)
	}
	return mutes, nil
}

func (db *DB) IsMuted(actorDID, subjectDID string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM mutes WHERE actor_did = $1 AND subject_did = $2)`, actorDID, subjectDID).Scan(&exists)
	return exists, err
}

func (db *DB) GetMutedDIDs(actorDID string) ([]string, error) {
	rows, err := db.Query(`SELECT subject_did FROM mutes WHERE actor_did = $1`, actorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dids []string
	for rows.Next() {
		var did string
		if err := rows.Scan(&did); err != nil {
			continue
		}
		dids = append(dids, did)
	}
	return dids, nil
}

func (db *DB) GetAllHiddenDIDs(actorDID string) (map[string]bool, error) {
	hidden := make(map[string]bool)
	if actorDID == "" {
		return hidden, nil
	}

	blocked, err := db.GetBlockedDIDs(actorDID)
	if err != nil {
		return hidden, err
	}
	for _, did := range blocked {
		hidden[did] = true
	}

	blockedBy, err := db.GetBlockedByDIDs(actorDID)
	if err != nil {
		return hidden, err
	}
	for _, did := range blockedBy {
		hidden[did] = true
	}

	muted, err := db.GetMutedDIDs(actorDID)
	if err != nil {
		return hidden, err
	}
	for _, did := range muted {
		hidden[did] = true
	}

	return hidden, nil
}

func (db *DB) GetViewerRelationship(viewerDID, subjectDID string) (blocked bool, muted bool, blockedBy bool, err error) {
	if viewerDID == "" || subjectDID == "" {
		return false, false, false, nil
	}

	blocked, err = db.IsBlocked(viewerDID, subjectDID)
	if err != nil {
		return
	}

	muted, err = db.IsMuted(viewerDID, subjectDID)
	if err != nil {
		return
	}

	blockedBy, err = db.IsBlocked(subjectDID, viewerDID)
	return
}

func (db *DB) CreateReport(reporterDID, subjectDID string, subjectURI *string, reasonType string, reasonText *string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO moderation_reports (reporter_did, subject_did, subject_uri, reason_type, reason_text, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6)
		RETURNING id
	`, reporterDID, subjectDID, subjectURI, reasonType, reasonText, time.Now()).Scan(&id)
	return id, err
}

func (db *DB) GetReports(status string, limit, offset int) ([]ModerationReport, error) {
	query := `SELECT id, reporter_did, subject_did, subject_uri, reason_type, reason_text, status, created_at, resolved_at, resolved_by
		FROM moderation_reports`
	args := []interface{}{}
	paramIdx := 1

	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
		paramIdx = 2
	}

	query += ` ORDER BY created_at DESC LIMIT $` + itoa(paramIdx) + ` OFFSET $` + itoa(paramIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []ModerationReport
	for rows.Next() {
		var r ModerationReport
		if err := rows.Scan(&r.ID, &r.ReporterDID, &r.SubjectDID, &r.SubjectURI, &r.ReasonType, &r.ReasonText, &r.Status, &r.CreatedAt, &r.ResolvedAt, &r.ResolvedBy); err != nil {
			continue
		}
		reports = append(reports, r)
	}
	return reports, nil
}

func (db *DB) GetReport(id int) (*ModerationReport, error) {
	var r ModerationReport
	err := db.QueryRow(`SELECT id, reporter_did, subject_did, subject_uri, reason_type, reason_text, status, created_at, resolved_at, resolved_by FROM moderation_reports WHERE id = $1`, id).Scan(
		&r.ID, &r.ReporterDID, &r.SubjectDID, &r.SubjectURI, &r.ReasonType, &r.ReasonText, &r.Status, &r.CreatedAt, &r.ResolvedAt, &r.ResolvedBy,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) ResolveReport(id int, resolvedBy string, status string) error {
	_, err := db.Exec(`UPDATE moderation_reports SET status = $1, resolved_at = $2, resolved_by = $3 WHERE id = $4`, status, time.Now(), resolvedBy, id)
	return err
}

func (db *DB) CreateModerationAction(reportID int, actorDID, action string, comment *string) error {
	_, err := db.Exec(`INSERT INTO moderation_actions (report_id, actor_did, action, comment, created_at) VALUES ($1, $2, $3, $4, $5)`, reportID, actorDID, action, comment, time.Now())
	return err
}

func (db *DB) GetReportActions(reportID int) ([]ModerationAction, error) {
	rows, err := db.Query(`SELECT id, report_id, actor_did, action, comment, created_at FROM moderation_actions WHERE report_id = $1 ORDER BY created_at DESC`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []ModerationAction
	for rows.Next() {
		var a ModerationAction
		if err := rows.Scan(&a.ID, &a.ReportID, &a.ActorDID, &a.Action, &a.Comment, &a.CreatedAt); err != nil {
			continue
		}
		actions = append(actions, a)
	}
	return actions, nil
}

func (db *DB) GetReportCount(status string) (int, error) {
	query := `SELECT COUNT(*) FROM moderation_reports`
	args := []interface{}{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (db *DB) CreateContentLabel(src, uri, val, createdBy string) error {
	_, err := db.Exec(`INSERT INTO content_labels (src, uri, val, neg, created_by, created_at) VALUES ($1, $2, $3, 0, $4, $5)`, src, uri, val, createdBy, time.Now())
	return err
}

func (db *DB) SyncSelfLabels(authorDID, uri string, labels []string) error {
	_, err := db.Exec(`DELETE FROM content_labels WHERE src = $1 AND uri = $2 AND created_by = $3`, authorDID, uri, authorDID)
	if err != nil {
		return err
	}
	for _, val := range labels {
		if err := db.CreateContentLabel(authorDID, uri, val, authorDID); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) NegateContentLabel(id int) error {
	_, err := db.Exec(`UPDATE content_labels SET neg = 1 WHERE id = $1`, id)
	return err
}

func (db *DB) DeleteContentLabel(id int) error {
	_, err := db.Exec(`DELETE FROM content_labels WHERE id = $1`, id)
	return err
}

func (db *DB) GetContentLabelsForURIs(uris []string, labelerDIDs []string) (map[string][]ContentLabel, error) {
	result := make(map[string][]ContentLabel)
	if len(uris) == 0 {
		return result, nil
	}

	query := `SELECT id, src, uri, val, neg, created_by, created_at FROM content_labels
		WHERE uri = ANY($1) AND neg = 0`
	args := []interface{}{pqStringArray(uris)}

	if len(labelerDIDs) > 0 {
		query += ` AND src = ANY($2)`
		args = append(args, pqStringArray(labelerDIDs))
	}

	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var l ContentLabel
		if err := rows.Scan(&l.ID, &l.Src, &l.URI, &l.Val, &l.Neg, &l.CreatedBy, &l.CreatedAt); err != nil {
			continue
		}
		result[l.URI] = append(result[l.URI], l)
	}
	return result, nil
}

func (db *DB) GetContentLabelsForDIDs(dids []string, labelerDIDs []string) (map[string][]ContentLabel, error) {
	result := make(map[string][]ContentLabel)
	if len(dids) == 0 {
		return result, nil
	}

	query := `SELECT id, src, uri, val, neg, created_by, created_at FROM content_labels
		WHERE uri = ANY($1) AND neg = 0`
	args := []interface{}{pqStringArray(dids)}

	if len(labelerDIDs) > 0 {
		query += ` AND src = ANY($2)`
		args = append(args, pqStringArray(labelerDIDs))
	}

	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var l ContentLabel
		if err := rows.Scan(&l.ID, &l.Src, &l.URI, &l.Val, &l.Neg, &l.CreatedBy, &l.CreatedAt); err != nil {
			continue
		}
		result[l.URI] = append(result[l.URI], l)
	}
	return result, nil
}

func (db *DB) GetAllContentLabels(limit, offset int) ([]ContentLabel, error) {
	rows, err := db.Query(`SELECT id, src, uri, val, neg, created_by, created_at FROM content_labels ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []ContentLabel
	for rows.Next() {
		var l ContentLabel
		if err := rows.Scan(&l.ID, &l.Src, &l.URI, &l.Val, &l.Neg, &l.CreatedBy, &l.CreatedAt); err != nil {
			continue
		}
		labels = append(labels, l)
	}
	return labels, nil
}

func itoa(i int) string {
	return strings.Repeat("", 0) + fmt.Sprintf("%d", i)
}
