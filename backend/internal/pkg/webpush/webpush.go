package webpush

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type WebPushClient struct {
	vapidPublicKey  string
	vapidPrivateKey string
	subscriber      string // Your email or website
}

// NewWebPushClient creates a new Web Push client
func NewWebPushClient(vapidPublicKey, vapidPrivateKey, subscriber string) *WebPushClient {
	return &WebPushClient{
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		subscriber:      subscriber,
	}
}

// PushSubscription represents a web push subscription
type PushSubscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// SendNotification sends a web push notification
func (w *WebPushClient) SendNotification(ctx context.Context, subscriptionJSON, title, body string, data map[string]interface{}) error {
	// Parse subscription
	var sub PushSubscription
	if err := json.Unmarshal([]byte(subscriptionJSON), &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Create notification payload
	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"icon":  "/icon-192x192.png",
		"badge": "/badge-72x72.png",
		"data":  data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create subscription object
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.Keys.P256dh,
			Auth:   sub.Keys.Auth,
		},
	}

	// Send notification
	resp, err := webpush.SendNotification(payloadBytes, subscription, &webpush.Options{
		Subscriber:      w.subscriber,
		VAPIDPublicKey:  w.vapidPublicKey,
		VAPIDPrivateKey: w.vapidPrivateKey,
		TTL:             30,
	})

	if err != nil {
		return fmt.Errorf("failed to send web push: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("web push failed with status: %d", resp.StatusCode)
	}

	log.Printf("Successfully sent web push notification: %s - %s", title, body)
	return nil
}

// GenerateVAPIDKeys generates VAPID keys for web push
// This should be run once and keys stored securely
func GenerateVAPIDKeys() (privateKey, publicKey string, err error) {
	privateKey, publicKey, err = webpush.GenerateVAPIDKeys()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate VAPID keys: %w", err)
	}
	return privateKey, publicKey, nil
}
