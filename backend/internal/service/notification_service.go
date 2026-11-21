package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/fcm"
	"zeinai-backend/internal/pkg/webpush"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo          domain.NotificationRepository
	fcmClient     *fcm.FCMClient
	webPushClient *webpush.WebPushClient
}

func NewNotificationService(repo domain.NotificationRepository, fcmClient *fcm.FCMClient, webPushClient *webpush.WebPushClient) *NotificationService {
	return &NotificationService{
		repo:          repo,
		fcmClient:     fcmClient,
		webPushClient: webPushClient,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, senderID *uuid.UUID, senderName, title, message string, notifType domain.NotificationType, link string) error {
	notification := &domain.Notification{
		UserID:     userID.String(),
		SenderName: senderName,
		Title:      title,
		Message:    message,
		Type:       notifType,
		Link:       link,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if senderID != nil {
		notification.SenderID = senderID.String()
	}

	log.Printf("[NotificationService] Creating notification for user %s, type: %s, title: %s", userID, notifType, title)
	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send push notification to user's devices
	go s.sendPushNotification(ctx, userID, title, message, link)

	return nil
}

// CheckSubscriptionExpiry checks and notifies users about expiring subscriptions
func (s *NotificationService) CheckSubscriptionExpiry(ctx context.Context, userID uuid.UUID, expiresAt *time.Time) error {
	if expiresAt == nil {
		return nil // No expiry (free plan or lifetime)
	}

	daysUntilExpiry := int(time.Until(*expiresAt).Hours() / 24)

	// Notify 7 days before expiry
	if daysUntilExpiry == 7 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Subscription Expiring Soon",
			"Your subscription will expire in 7 days. Renew now to continue using premium features.",
			domain.NotificationSubscriptionExpiring,
			"",
		)
	}

	// Notify 3 days before expiry
	if daysUntilExpiry == 3 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Subscription Expiring Soon",
			"Your subscription will expire in 3 days. Don't lose access to your premium features!",
			domain.NotificationSubscriptionExpiring,
			"",
		)
	}

	// Notify 1 day before expiry
	if daysUntilExpiry == 1 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Subscription Expiring Tomorrow",
			"Your subscription expires tomorrow. Renew now to avoid interruption.",
			domain.NotificationSubscriptionExpiring,
			"",
		)
	}

	// Notify on expiry
	if daysUntilExpiry <= 0 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Subscription Expired",
			"Your subscription has expired. Please renew to continue using premium features.",
			domain.NotificationSubscriptionExpired,
			"",
		)
	}

	return nil
}

// CheckTokenLimit checks and notifies about token usage
func (s *NotificationService) CheckTokenLimit(ctx context.Context, userID uuid.UUID, tokensUsed, tokensLimit int) error {
	if tokensLimit == -1 {
		return nil // Unlimited
	}

	percentage := float64(tokensUsed) / float64(tokensLimit) * 100

	// Notify at 80% usage
	if percentage >= 80 && percentage < 90 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Token Limit Warning",
			fmt.Sprintf("You've used %d%% of your monthly tokens. Consider upgrading your plan.", int(percentage)),
			domain.NotificationTokenLimitWarning,
			"",
		)
	}

	// Notify at 90% usage
	if percentage >= 90 && percentage < 100 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Token Limit Almost Reached",
			fmt.Sprintf("You've used %d%% of your monthly tokens. Upgrade to continue using AI features.", int(percentage)),
			domain.NotificationTokenLimitWarning,
			"",
		)
	}

	// Notify at 100% usage
	if percentage >= 100 {
		return s.CreateNotification(ctx, userID, nil, "",
			"Token Limit Reached",
			"You've reached your monthly token limit. Upgrade your plan to continue.",
			domain.NotificationTokenLimitReached,
			"",
		)
	}

	return nil
}

// sendPushNotification sends push notification to user's devices
func (s *NotificationService) sendPushNotification(ctx context.Context, userID uuid.UUID, title, message string, link string) {
	tokens, err := s.repo.GetActivePushTokens(ctx, userID)
	if err != nil {
		log.Printf("Failed to get push tokens for user %s: %v", userID, err)
		return
	}

	for _, token := range tokens {
		go func(t domain.PushToken) {
			var err error

			switch t.Platform {
			case "android", "ios":
				// FCM for mobile
				if s.fcmClient != nil {
					data := map[string]string{
						"type":    "notification",
						"user_id": userID.String(),
						"link":    link,
					}
					err = s.fcmClient.SendNotification(ctx, t.Token, title, message, data)
				}

			case "web":
				// Web Push for browser
				if s.webPushClient != nil {
					data := map[string]interface{}{
						"type":    "notification",
						"user_id": userID.String(),
						"link":    link,
					}
					err = s.webPushClient.SendNotification(ctx, t.Token, title, message, data)
				}
			}

			if err != nil {
				log.Printf("Failed to send push to %s device: %v", t.Platform, err)
				// TODO: Deactivate token if it's invalid
			} else {
				log.Printf("Successfully sent push to %s device: %s", t.Platform, t.DeviceName)
			}
		}(token)
	}
}
