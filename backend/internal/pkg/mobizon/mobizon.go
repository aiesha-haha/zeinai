package mobizon

import (
	"log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	MobizonAPIURL = "https://api.mobizon.kz/service"
)

type MobizonClient struct {
	apiKey     string
	httpClient *http.Client
}

type MobizonResponse struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type SMSMessageData struct {
	MessageID string      `json:"messageId"`
	Status    interface{} `json:"status"`
}

// NewMobizonClient creates a new Mobizon API client
func NewMobizonClient(apiKey string) *MobizonClient {
	return &MobizonClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendSMS sends an SMS message via Mobizon API
func (m *MobizonClient) SendSMS(phone, message string) (string, error) {
	// Mobizon expects phone in international format WITHOUT '+' (e.g. 77771112233)
	cleanPhone := strings.TrimPrefix(phone, "+")

	// Prepare request
	data := url.Values{}
	data.Set("recipient", cleanPhone)
	data.Set("text", message)
	data.Set("apiKey", m.apiKey)

	// Make request
	resp, err := m.httpClient.PostForm(MobizonAPIURL+"/message/sendsmsmessage", data)
	if err != nil {
		return "", fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var mobizonResp MobizonResponse
	if err := json.Unmarshal(body, &mobizonResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	if mobizonResp.Code != 0 {
		// Keep Mobizon's own explanation (e.g. low balance, invalid number) for the logs
		log.Printf("[Mobizon] send failed: code=%d message=%q data=%s", mobizonResp.Code, mobizonResp.Message, string(mobizonResp.Data))
		return "", fmt.Errorf("mobizon API error: code %d: %s", mobizonResp.Code, mobizonResp.Message)
	}

	// Parse message data
	var msgData SMSMessageData
	if err := json.Unmarshal(mobizonResp.Data, &msgData); err != nil {
		return "", fmt.Errorf("failed to parse message data: %w", err)
	}

	return msgData.MessageID, nil
}

// GetBalance gets account balance
func (m *MobizonClient) GetBalance() (float64, error) {
	data := url.Values{}
	data.Set("apiKey", m.apiKey)

	resp, err := m.httpClient.PostForm(MobizonAPIURL+"/user/getownbalance", data)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var mobizonResp struct {
		Code int `json:"code"`
		Data struct {
			Balance float64 `json:"balance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &mobizonResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if mobizonResp.Code != 0 {
		return 0, fmt.Errorf("mobizon API error: code %d", mobizonResp.Code)
	}

	return mobizonResp.Data.Balance, nil
}

// SendVerificationCode sends a verification code via SMS
func (m *MobizonClient) SendVerificationCode(phone, code string) error {
	message := fmt.Sprintf("Код подтверждения: %s\nZeinAI", code)

	_, err := m.SendSMS(phone, message)
	return err
}

// SendPasswordResetCode sends a password reset code via SMS
func (m *MobizonClient) SendPasswordResetCode(phone, code string) error {
	message := fmt.Sprintf("Код сброса пароля: %s\nZeinAI", code)

	_, err := m.SendSMS(phone, message)
	return err
}
