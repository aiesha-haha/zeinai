package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const StudioQueue = "ai_studio.generate"

// ErrStudioNotConfigured is returned when the AI provider key needed for a kind is missing.
var ErrStudioNotConfigured = errors.New("ai provider is not configured")

// ErrStudioNotFound is returned when the item does not exist or belongs to someone else.
var ErrStudioNotFound = errors.New("item not found")

// StudioMessage is published to the AI Studio queue.
type StudioMessage struct {
	ItemID uuid.UUID `json:"item_id"`
}

type CreateStudioRequest struct {
	Kind       string         `json:"kind" validate:"required,oneof=image video feedback lab news certificate"`
	Title      string         `json:"title" validate:"omitempty,max=500"`
	RequestKey string         `json:"request_key" validate:"omitempty,max=100"`
	Input      map[string]any `json:"input" validate:"required"`
}

type UpdateStudioRequest struct {
	Title   string          `json:"title" validate:"required,max=500"`
	Content json.RawMessage `json:"content" validate:"required"`
}

type StudioService struct {
	repo             domain.StudioRepository
	subscriptionRepo domain.SubscriptionRepository
	channel          *amqp.Channel
	cfg              *config.Config
}

func NewStudioService(repo domain.StudioRepository, subscriptionRepo domain.SubscriptionRepository, channel *amqp.Channel, cfg *config.Config) *StudioService {
	return &StudioService{repo: repo, subscriptionRepo: subscriptionRepo, channel: channel, cfg: cfg}
}

// MissingKeyFor returns the env variable that must be set for a kind, or "".
func (s *StudioService) MissingKeyFor(kind string) string {
	if domain.IsStudioMediaKind(kind) {
		if s.cfg.OpenAI.APIKey == "" {
			return "OPENAI_API_KEY"
		}
		return ""
	}
	if s.cfg.Kimi.APIKey == "" && s.cfg.OpenAI.APIKey == "" && s.cfg.Gemini.APIKey == "" {
		return "KIMI_API_KEY, OPENAI_API_KEY or GEMINI_API_KEY"
	}
	return ""
}

func studioStr(in map[string]any, key string) string {
	if v, ok := in[key]; ok && v != nil {
		switch t := v.(type) {
		case string:
			return strings.TrimSpace(t)
		case float64:
			return fmt.Sprintf("%g", t)
		}
	}
	return ""
}

func studioTruncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n]) + "…"
}

// validateInput checks the required fields of each kind and fills defaults.
// It returns a default title for the item.
func validateStudioInput(kind string, in map[string]any) (string, error) {
	need := func(keys ...string) error {
		for _, k := range keys {
			if studioStr(in, k) == "" {
				return fmt.Errorf("field %q is required", k)
			}
			if utf8.RuneCountInString(studioStr(in, k)) > 3000 {
				return fmt.Errorf("field %q is too long", k)
			}
		}
		return nil
	}
	if studioStr(in, "language") == "" {
		in["language"] = "kaz"
	}
	switch kind {
	case domain.StudioKindImage:
		if err := need("prompt"); err != nil {
			return "", err
		}
		switch studioStr(in, "size") {
		case "", "1024x1024", "1536x1024", "1024x1536":
		default:
			return "", fmt.Errorf("unsupported size")
		}
		return studioTruncate(studioStr(in, "prompt"), 80), nil
	case domain.StudioKindVideo:
		if err := need("description"); err != nil {
			return "", err
		}
		switch studioStr(in, "seconds") {
		case "":
			in["seconds"] = "8"
		case "4", "8", "12":
		default:
			return "", fmt.Errorf("seconds must be 4, 8 or 12")
		}
		if t := studioStr(in, "title"); t != "" {
			return studioTruncate(t, 120), nil
		}
		return studioTruncate(studioStr(in, "description"), 80), nil
	case domain.StudioKindFeedback:
		if err := need("topic", "purpose"); err != nil {
			return "", err
		}
		return studioTruncate(studioStr(in, "topic"), 120), nil
	case domain.StudioKindLab:
		if err := need("subject", "topic"); err != nil {
			return "", err
		}
		return studioTruncate(studioStr(in, "topic"), 120), nil
	case domain.StudioKindNews:
		if err := need("topic"); err != nil {
			return "", err
		}
		return studioTruncate(studioStr(in, "topic"), 120), nil
	case domain.StudioKindCertificate:
		if err := need("student_name", "achievement"); err != nil {
			return "", err
		}
		return studioTruncate(studioStr(in, "student_name")+" — "+studioStr(in, "achievement"), 120), nil
	}
	return "", fmt.Errorf("unknown kind")
}

// Create stores a new request and queues it. Sending the same request_key twice
// returns the first item instead of starting (and paying for) a second generation.
func (s *StudioService) Create(ctx context.Context, userID uuid.UUID, req *CreateStudioRequest) (*domain.StudioItem, error) {
	if req.RequestKey != "" {
		if existing, err := s.repo.GetByRequestKey(ctx, userID, req.RequestKey); err == nil {
			return existing, nil
		}
	}
	if missing := s.MissingKeyFor(req.Kind); missing != "" {
		return nil, fmt.Errorf("%w: set %s on the server", ErrStudioNotConfigured, missing)
	}

	defaultTitle, err := validateStudioInput(req.Kind, req.Input)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = defaultTitle
	}
	inputJSON, err := json.Marshal(req.Input)
	if err != nil {
		return nil, err
	}

	item := &domain.StudioItem{
		ID:      uuid.New(),
		OwnerID: userID,
		Kind:    req.Kind,
		Title:   title,
		Input:   inputJSON,
		Status:  domain.StudioStatusPending,
	}
	if req.RequestKey != "" {
		k := req.RequestKey
		item.RequestKey = &k
	}

	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	if !created {
		// Concurrent duplicate click: return the item that won the race
		return s.repo.GetByRequestKey(ctx, userID, req.RequestKey)
	}

	cost := domain.StudioTokenCost[req.Kind]
	if cost > 0 {
		if err := s.subscriptionRepo.ConsumeTokens(ctx, userID, cost, "ai_studio_"+req.Kind); err != nil {
			msg := "insufficient tokens"
			_ = s.repo.UpdateStatus(ctx, item.ID, domain.StudioStatusFailed, &msg)
			return nil, fmt.Errorf("insufficient tokens: %w", err)
		}
	}

	if err := s.publish(ctx, item.ID); err != nil {
		msg := "queue unavailable, please retry"
		_ = s.repo.UpdateStatus(ctx, item.ID, domain.StudioStatusFailed, &msg)
		item.Status = domain.StudioStatusFailed
		item.ErrorMessage = &msg
		return item, nil
	}
	return item, nil
}

func (s *StudioService) publish(ctx context.Context, id uuid.UUID) error {
	if s.channel == nil {
		return fmt.Errorf("rabbitmq channel is nil")
	}
	body, _ := json.Marshal(StudioMessage{ItemID: id})
	return s.channel.PublishWithContext(ctx, "", StudioQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (s *StudioService) getOwned(ctx context.Context, id, userID uuid.UUID) (*domain.StudioItem, error) {
	it, err := s.repo.GetByID(ctx, id)
	if err != nil || it.OwnerID != userID {
		return nil, ErrStudioNotFound
	}
	return it, nil
}

func (s *StudioService) Get(ctx context.Context, id, userID uuid.UUID) (*domain.StudioItem, error) {
	return s.getOwned(ctx, id, userID)
}

func (s *StudioService) List(ctx context.Context, userID uuid.UUID, kind string) ([]domain.StudioItem, error) {
	return s.repo.ListByOwner(ctx, userID, kind, 50)
}

// Update saves the teacher's edits of a text result (feedback, lab, news, certificate).
func (s *StudioService) Update(ctx context.Context, id, userID uuid.UUID, req *UpdateStudioRequest) (*domain.StudioItem, error) {
	it, err := s.getOwned(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if domain.IsStudioMediaKind(it.Kind) {
		return nil, fmt.Errorf("media items cannot be edited")
	}
	if it.Status != domain.StudioStatusCompleted {
		return nil, fmt.Errorf("item is not ready yet")
	}
	if !json.Valid(req.Content) {
		return nil, fmt.Errorf("content must be valid JSON")
	}
	if err := s.repo.UpdateContent(ctx, id, strings.TrimSpace(req.Title), req.Content); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Retry re-queues a failed item. Tokens are not charged again.
func (s *StudioService) Retry(ctx context.Context, id, userID uuid.UUID) (*domain.StudioItem, error) {
	it, err := s.getOwned(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if it.Status != domain.StudioStatusFailed {
		return nil, fmt.Errorf("only failed items can be retried")
	}
	if missing := s.MissingKeyFor(it.Kind); missing != "" {
		return nil, fmt.Errorf("%w: set %s on the server", ErrStudioNotConfigured, missing)
	}
	if err := s.repo.UpdateStatus(ctx, id, domain.StudioStatusPending, nil); err != nil {
		return nil, err
	}
	if err := s.publish(ctx, id); err != nil {
		msg := "queue unavailable, please retry"
		_ = s.repo.UpdateStatus(ctx, id, domain.StudioStatusFailed, &msg)
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *StudioService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.getOwned(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}
