// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pineapple/msg-demo/backend/inbox/internal/svc"
	"github.com/pineapple/msg-demo/backend/inbox/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageRequest) (*types.Message, error) {
	channel := req.Channel
	if channel == "" {
		channel = "personal"
	}

	switch channel {
	case "personal":
		return l.sendPersonalMessage(req)
	case "system":
		return l.sendSystemNotification(req)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}
}

func (l *SendMessageLogic) sendPersonalMessage(req *types.SendMessageRequest) (*types.Message, error) {
	if req.SenderId <= 0 {
		return nil, errors.New("senderId is required for personal channel")
	}

	result, err := l.svcCtx.DB.ExecContext(
		l.ctx,
		`INSERT INTO direct_messages (sender_id, receiver_id, title, content) VALUES (?, ?, ?, ?)`,
		req.SenderId,
		req.ReceiverId,
		req.Title,
		req.Content,
	)
	if err != nil {
		return nil, fmt.Errorf("insert personal message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("fetch personal message id: %w", err)
	}

	return fetchPersonalMessage(l.ctx, l.svcCtx.DB, messageID)
}

func (l *SendMessageLogic) sendSystemNotification(req *types.SendMessageRequest) (*types.Message, error) {
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	createdBy := req.SenderId
	if createdBy < 0 {
		createdBy = 0
	}

	res, err := tx.ExecContext(
		l.ctx,
		`INSERT INTO system_notifications (title, content, priority, created_by) VALUES (?, ?, ?, ?)`,
		req.Title,
		req.Content,
		req.Priority,
		createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("insert system notification: %w", err)
	}

	notificationID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("fetch notification id: %w", err)
	}

	receiptRes, err := tx.ExecContext(
		l.ctx,
		`INSERT INTO system_notification_receipts (notification_id, user_id) VALUES (?, ?)`,
		notificationID,
		req.ReceiverId,
	)
	if err != nil {
		return nil, fmt.Errorf("insert notification receipt: %w", err)
	}

	receiptID, err := receiptRes.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("fetch receipt id: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit system notification: %w", err)
	}
	committed = true

	msg, err := fetchSystemMessage(l.ctx, l.svcCtx.DB, receiptID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return &types.Message{
			Id:         receiptID,
			SenderId:   createdBy,
			ReceiverId: req.ReceiverId,
			Title:      req.Title,
			Content:    req.Content,
			IsRead:     false,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
			Channel:    "system",
			Priority:   req.Priority,
		}, nil
	}

	return msg, err
}

type ListMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMessagesLogic) ListMessages(req *types.ListMessagesRequest) (*types.ListMessagesResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	if req.UserId <= 0 {
		return nil, fmt.Errorf("缺少用户信息")
	}

	var (
		items []types.Message
		total int64
		err   error
	)

	switch req.Channel {
	case "system":
		items, total, err = l.listSystemMessages(req)
	default:
		items, total, err = l.listPersonalMessages(req)
	}

	if err != nil {
		return nil, err
	}

	return &types.ListMessagesResponse{
		Items: items,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

func (l *ListMessagesLogic) listPersonalMessages(req *types.ListMessagesRequest) ([]types.Message, int64, error) {
	where := "receiver_id = ?"
	args := []interface{}{req.UserId}

	if req.Status == "sent" {
		where = "sender_id = ?"
		args[0] = req.UserId
	}

	if req.Status == "unread" {
		where += " AND is_read = 0"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM direct_messages WHERE %s", where)
	var total int64
	if err := l.svcCtx.DB.QueryRowContext(l.ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count personal messages: %w", err)
	}

	query := fmt.Sprintf(`
SELECT id, sender_id, receiver_id, title, content, is_read, read_at, created_at
FROM direct_messages
WHERE %s
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, where)

	argsWithPaging := append([]interface{}{}, args...)
	offset := (req.Page - 1) * req.Size
	argsWithPaging = append(argsWithPaging, req.Size, offset)

	rows, err := l.svcCtx.DB.QueryContext(l.ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, fmt.Errorf("list personal messages: %w", err)
	}
	defer rows.Close()

	var items []types.Message
	for rows.Next() {
		msg, err := scanPersonalRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, msg)
	}

	return items, total, nil
}

func (l *ListMessagesLogic) listSystemMessages(req *types.ListMessagesRequest) ([]types.Message, int64, error) {
	where := "snu.user_id = ?"
	args := []interface{}{req.UserId}

	if req.Status == "unread" {
		where += " AND snu.is_read = 0"
	}

	countQuery := fmt.Sprintf(`
SELECT COUNT(*) FROM system_notification_receipts snu
WHERE %s`, where)

	var total int64
	if err := l.svcCtx.DB.QueryRowContext(l.ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count system notifications: %w", err)
	}

	query := fmt.Sprintf(`
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
WHERE %s
ORDER BY snu.created_at DESC
LIMIT ? OFFSET ?`, where)

	argsWithPaging := append([]interface{}{}, args...)
	offset := (req.Page - 1) * req.Size
	argsWithPaging = append(argsWithPaging, req.Size, offset)

	rows, err := l.svcCtx.DB.QueryContext(l.ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, fmt.Errorf("list system notifications: %w", err)
	}
	defer rows.Close()

	var items []types.Message
	for rows.Next() {
		msg, err := scanSystemRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, msg)
	}

	return items, total, nil
}

func scanPersonalRow(scanner interface {
	Scan(dest ...interface{}) error
}) (types.Message, error) {
	var (
		msg       types.Message
		title     sql.NullString
		readAt    sql.NullTime
		createdAt time.Time
	)

	err := scanner.Scan(
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
		return types.Message{}, fmt.Errorf("scan personal message: %w", err)
	}

	msg.Title = title.String
	msg.ReadAt = formatNullTime(readAt)
	msg.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	msg.Channel = "personal"

	return msg, nil
}

func scanSystemRow(scanner interface {
	Scan(dest ...interface{}) error
}) (types.Message, error) {
	var (
		msg       types.Message
		readAt    sql.NullTime
		createdAt time.Time
	)

	err := scanner.Scan(
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
		return types.Message{}, fmt.Errorf("scan system message: %w", err)
	}

	msg.ReadAt = formatNullTime(readAt)
	msg.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	msg.Channel = "system"

	return msg, nil
}

type MarkMessageReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkMessageReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkMessageReadLogic {
	return &MarkMessageReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkMessageReadLogic) MarkMessageRead(req *types.MarkReadRequest) (*types.Message, error) {
	if req.UserId <= 0 {
		return nil, fmt.Errorf("缺少用户信息")
	}

	var (
		msg *types.Message
		err error
	)

	switch req.Channel {
	case "system":
		err = l.markSystemReceipt(req.Id, req.UserId)
		if err == nil {
			msg, err = fetchSystemMessage(l.ctx, l.svcCtx.DB, req.Id)
		}
	default:
		err = l.markPersonalReceipt(req.Id, req.UserId)
		if err == nil {
			msg, err = fetchPersonalMessage(l.ctx, l.svcCtx.DB, req.Id)
		}
	}

	if err != nil {
		return nil, translateNotFound(err)
	}

	return msg, nil
}

func (l *MarkMessageReadLogic) markPersonalReceipt(messageID, userID int64) error {
	result, err := l.svcCtx.DB.ExecContext(
		l.ctx,
		`UPDATE direct_messages SET is_read = 1, read_at = ? WHERE id = ? AND receiver_id = ?`,
		time.Now(),
		messageID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update personal message: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("personal rows affected: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (l *MarkMessageReadLogic) markSystemReceipt(receiptID, userID int64) error {
	result, err := l.svcCtx.DB.ExecContext(
		l.ctx,
		`UPDATE system_notification_receipts SET is_read = 1, read_at = ? WHERE id = ? AND user_id = ?`,
		time.Now(),
		receiptID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update system notification: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("system rows affected: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

type UnreadCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnreadCountLogic {
	return &UnreadCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnreadCountLogic) UnreadCount(req *types.UnreadCountRequest) (*types.UnreadCountResponse, error) {
	if req.UserId <= 0 {
		return nil, fmt.Errorf("缺少用户信息")
	}

	var (
		personal int64
		system   int64
	)

	if err := l.svcCtx.DB.QueryRowContext(
		l.ctx,
		`SELECT COUNT(*) FROM direct_messages WHERE receiver_id = ? AND is_read = 0`,
		req.UserId,
	).Scan(&personal); err != nil {
		return nil, fmt.Errorf("personal unread count: %w", err)
	}

	if err := l.svcCtx.DB.QueryRowContext(
		l.ctx,
		`SELECT COUNT(*) FROM system_notification_receipts WHERE user_id = ? AND is_read = 0`,
		req.UserId,
	).Scan(&system); err != nil {
		return nil, fmt.Errorf("system unread count: %w", err)
	}

	return &types.UnreadCountResponse{
		Personal: personal,
		System:   system,
		Total:    personal + system,
	}, nil
}

func translateNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("message not found or unauthorized")
	}
	return err
}
