package repository

import (
	"context"
	"fmt"
	"time"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) domain.NotificationRepository {
	return &NotificationRepo{db: db}
}

// Create creates a new notification
func (r *NotificationRepo) Create(ctx context.Context, notification *domain.Notification) error {
	query := `
		INSERT INTO notifications (user_id, sender_id, sender_name, title, message, type, link, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id uuid.UUID
	var senderID interface{} = notification.SenderID
	if notification.SenderID == "" {
		senderID = nil
	}

	err := r.db.QueryRow(ctx, query,
		notification.UserID,
		senderID,
		notification.SenderName,
		notification.Title,
		notification.Message,
		notification.Type,
		notification.Link,
		notification.IsRead,
		notification.CreatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	notification.ID = id.String()
	return nil
}

// GetByUserID gets all notifications for a user
func (r *NotificationRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, isRead *bool) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, sender_id, sender_name, title, message, type, link, is_read, created_at, read_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	if isRead != nil {
		argCount++
		query += fmt.Sprintf(" AND is_read = $%d", argCount)
		args = append(args, *isRead)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var id uuid.UUID
		var userIDStr uuid.UUID
		var senderID *uuid.UUID
		var senderName *string

		err := rows.Scan(&id, &userIDStr, &senderID, &senderName, &n.Title, &n.Message, &n.Type, &n.Link, &n.IsRead, &n.CreatedAt, &n.ReadAt)
		if err != nil {
			return nil, err
		}

		n.ID = id.String()
		n.UserID = userIDStr.String()
		if senderID != nil {
			n.SenderID = senderID.String()
		}
		if senderName != nil {
			n.SenderName = *senderName
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// GetUnreadCount gets count of unread notifications
func (r *NotificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// GetTotalCount gets total count of notifications for a user (including read)
func (r *NotificationRepo) GetTotalCount(ctx context.Context, userID uuid.UUID, isRead *bool) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	args := []interface{}{userID}

	if isRead != nil {
		query += " AND is_read = $2"
		args = append(args, *isRead)
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

// MarkAsRead marks notification as read
func (r *NotificationRepo) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all user's notifications as read
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW()
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// RegisterPushToken registers a push notification token
func (r *NotificationRepo) RegisterPushToken(ctx context.Context, token *domain.PushToken) error {
	// First, deactivate old tokens for same device
	deactivateQuery := `
		UPDATE push_tokens
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND platform = $2 AND token != $3
	`
	_, err := r.db.Exec(ctx, deactivateQuery, token.UserID, token.Platform, token.Token)
	if err != nil {
		return fmt.Errorf("failed to deactivate old tokens: %w", err)
	}

	// Insert or update token
	query := `
		INSERT INTO push_tokens (user_id, token, platform, device_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		ON CONFLICT (user_id, token) 
		DO UPDATE SET is_active = true, updated_at = NOW(), device_name = EXCLUDED.device_name
		RETURNING id
	`

	var id uuid.UUID
	err = r.db.QueryRow(ctx, query,
		token.UserID,
		token.Token,
		token.Platform,
		token.DeviceName,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to register push token: %w", err)
	}

	token.ID = id.String()
	token.IsActive = true
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()

	return nil
}

// GetActivePushTokens gets all active push tokens for a user
func (r *NotificationRepo) GetActivePushTokens(ctx context.Context, userID uuid.UUID) ([]domain.PushToken, error) {
	query := `
		SELECT id, user_id, token, platform, device_name, is_active, created_at, updated_at
		FROM push_tokens
		WHERE user_id = $1 AND is_active = true
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push tokens: %w", err)
	}
	defer rows.Close()

	var tokens []domain.PushToken
	for rows.Next() {
		var t domain.PushToken
		var id, userIDUUID uuid.UUID

		err := rows.Scan(&id, &userIDUUID, &t.Token, &t.Platform, &t.DeviceName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}

		t.ID = id.String()
		t.UserID = userIDUUID.String()
		tokens = append(tokens, t)
	}

	return tokens, nil
}

// DeactivatePushToken deactivates a push token
func (r *NotificationRepo) DeactivatePushToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE push_tokens
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to deactivate push token: %w", err)
	}

	return nil
}

// Delete deletes a notification
func (r *NotificationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}
