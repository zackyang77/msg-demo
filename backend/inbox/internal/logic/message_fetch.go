package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/pineapple/msg-demo/backend/inbox/internal/types"
)

func fetchPersonalMessage(ctx context.Context, db *sql.DB, id int64) (*types.Message, error) {
	var (
		msg       types.Message
		title     sql.NullString
		readAt    sql.NullTime
		createdAt time.Time
	)

	query := `
SELECT id, sender_id, receiver_id, title, content, is_read, read_at, created_at
FROM direct_messages WHERE id = ?`

	err := db.QueryRowContext(ctx, query, id).Scan(
		&msg.Id,
		&msg.SenderId,
		&msg.ReceiverId,
		&title,
		&msg.Content,
		&msg.IsRead,
		&readAt,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	msg.Title = title.String
	msg.ReadAt = formatNullTime(readAt)
	msg.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	msg.Channel = "personal"

	return &msg, nil
}

func fetchSystemMessage(ctx context.Context, db *sql.DB, receiptID int64) (*types.Message, error) {
	var (
		msg       types.Message
		readAt    sql.NullTime
		createdAt time.Time
	)

	query := `
SELECT
	snu.id,
	sn.created_by,
	snu.user_id,
	sn.title,
	sn.content,
	snu.is_read,
	snu.read_at,
	snu.created_at,
	sn.priority
FROM system_notification_receipts snu
JOIN system_notifications sn ON snu.notification_id = sn.id
WHERE snu.id = ?`

	err := db.QueryRowContext(ctx, query, receiptID).Scan(
		&msg.Id,
		&msg.SenderId,
		&msg.ReceiverId,
		&msg.Title,
		&msg.Content,
		&msg.IsRead,
		&readAt,
		&createdAt,
		&msg.Priority,
	)
	if err != nil {
		return nil, err
	}

	msg.ReadAt = formatNullTime(readAt)
	msg.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	msg.Channel = "system"

	return &msg, nil
}

func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
