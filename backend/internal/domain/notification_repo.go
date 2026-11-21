package domain

import (
	"context"

	"github.com/google/uuid"
)

// NotificationRepository handles notification operations
type NotificationRepository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *Notification) error
	
	// GetByUserID gets all notifications for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, isRead *bool) ([]Notification, error)
	
	// GetUnreadCount gets count of unread notifications
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	
	// GetTotalCount gets total count of notifications for a user (including read)
	GetTotalCount(ctx context.Context, userID uuid.UUID, isRead *bool) (int, error)
	
	// MarkAsRead marks notification as read
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	
	// MarkAllAsRead marks all user's notifications as read
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	
	// RegisterPushToken registers a push notification token
	RegisterPushToken(ctx context.Context, token *PushToken) error
	
	// GetActivePushTokens gets all active push tokens for a user
	GetActivePushTokens(ctx context.Context, userID uuid.UUID) ([]PushToken, error)
	
	// DeactivatePushToken deactivates a push token
	DeactivatePushToken(ctx context.Context, tokenID uuid.UUID) error

	// Delete deletes a notification
	Delete(ctx context.Context, id uuid.UUID) error
}
