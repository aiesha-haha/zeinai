package domain

import "time"

// NotificationType represents type of notification
type NotificationType string

const (
	NotificationSubscriptionExpiring NotificationType = "subscription_expiring"
	NotificationSubscriptionExpired  NotificationType = "subscription_expired"
	NotificationTokenLimitWarning    NotificationType = "token_limit_warning"
	NotificationTokenLimitReached    NotificationType = "token_limit_reached"
	NotificationStudentLimitWarning  NotificationType = "student_limit_warning"
	NotificationMonthlyReset         NotificationType = "monthly_reset"
	NotificationSystem               NotificationType = "system"

	// Material generation
	NotificationMaterialReady NotificationType = "material_ready"

	// Whiteboard
	NotificationWhiteboardStarted      NotificationType = "whiteboard_started"
	NotificationWhiteboardSummaryReady NotificationType = "whiteboard_summary_ready"
)

// Notification represents a user notification
type Notification struct {
	ID         string           `json:"id"`
	UserID     string           `json:"user_id"`
	SenderID   string           `json:"sender_id,omitempty"`
	SenderName string           `json:"sender_name,omitempty"`
	Title      string           `json:"title"`
	Message    string           `json:"message"`
	Type       NotificationType `json:"type"`
	Link       string           `json:"link,omitempty"`
	IsRead     bool             `json:"is_read"`
	CreatedAt  time.Time        `json:"created_at"`
	ReadAt     *time.Time       `json:"read_at,omitempty"`
}

// PushToken represents a device push notification token
type PushToken struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Token      string    `json:"token"`
	Platform   string    `json:"platform"` // 'web', 'ios', 'android'
	DeviceName string    `json:"device_name,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RegisterPushTokenRequest represents push token registration
type RegisterPushTokenRequest struct {
	Token      string `json:"token" validate:"required"`
	Platform   string `json:"platform" validate:"required,oneof=web ios android"`
	DeviceName string `json:"device_name,omitempty"`
}
