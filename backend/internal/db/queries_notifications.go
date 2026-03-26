package db

import (
	"time"
)

func (db *DB) CreateNotification(n *Notification) error {
	_, err := db.Exec(`
		INSERT INTO notifications (recipient_did, actor_did, type, subject_uri, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, n.RecipientDID, n.ActorDID, n.Type, n.SubjectURI, n.CreatedAt)
	return err
}

func (db *DB) GetNotifications(recipientDID string, limit, offset int) ([]Notification, error) {
	rows, err := db.Query(`
		SELECT id, recipient_did, actor_did, type, subject_uri, created_at, read_at
		FROM notifications
		WHERE recipient_did = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, recipientDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.RecipientDID, &n.ActorDID, &n.Type, &n.SubjectURI, &n.CreatedAt, &n.ReadAt); err != nil {
			continue
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (db *DB) GetUnreadNotificationCount(recipientDID string) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM notifications WHERE recipient_did = $1 AND read_at IS NULL
	`, recipientDID).Scan(&count)
	return count, err
}

func (db *DB) MarkNotificationsRead(recipientDID string) error {
	_, err := db.Exec(`
		UPDATE notifications SET read_at = $1 WHERE recipient_did = $2 AND read_at IS NULL
	`, time.Now(), recipientDID)
	return err
}
