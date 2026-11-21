package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	presentationQueue      = "presentation.generation"
	presentationMaxRetries = 3
	presentationTimeout    = 180 * time.Second // Kimi K2 may take longer for large slide decks
)

// aiSlide is the JSON shape we ask Kimi K2 to generate per slide.
type aiSlide struct {
	SlideNumber int    `json:"slide_number"`
	Title       string `json:"title"`
	Content     string `json:"content"` // "\n" separated bullets
	Notes       string `json:"notes"`
	Layout      string `json:"layout"` // title_only | bullets | two_column | image_text
	ImagePrompt string `json:"image_prompt"`
}

// aiPresentationResponse is the full JSON shape we expect from Kimi K2.
type aiPresentationResponse struct {
	Slides []aiSlide `json:"slides"`
}

// PresentationWorker consumes from the presentation.generation queue and
// calls Kimi K2 to generate slide content.
type PresentationWorker struct {
	repo         repository.PresentationRepository
	notification *service.NotificationService
	ai           ai.TextGenerator
	imageAi      ai.ImageGenerator
	channel      *amqp.Channel
}

func NewPresentationWorker(
	repo repository.PresentationRepository,
	notification *service.NotificationService,
	aiGen ai.TextGenerator,
	imageAi ai.ImageGenerator,
	channel *amqp.Channel,
) *PresentationWorker {
	return &PresentationWorker{
		repo:         repo,
		notification: notification,
		ai:           aiGen,
		imageAi:      imageAi,
		channel:      channel,
	}
}

// Start begins consuming messages from the presentation.generation queue.
func (w *PresentationWorker) Start(ctx context.Context) error {
	q, err := w.channel.QueueDeclare(
		presentationQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("[PresentationWorker] failed to declare queue: %w", err)
	}

	if err := w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("[PresentationWorker] failed to set QoS: %w", err)
	}

	msgs, err := w.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("[PresentationWorker] failed to consume: %w", err)
	}

	log.Printf("[PresentationWorker] Started, listening on queue: %s", q.Name)

	for {
		select {
		case <-ctx.Done():
			log.Println("[PresentationWorker] Shutting down...")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[PresentationWorker] Channel closed")
				return nil
			}
			w.processMessage(ctx, msg)
		}
	}
}

func (w *PresentationWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	var genMsg service.PresentationGenerationMessage
	if err := json.Unmarshal(msg.Body, &genMsg); err != nil {
		log.Printf("[PresentationWorker] Bad message body: %v", err)
		msg.Nack(false, false) // Discard – cannot parse
		return
	}

	log.Printf("[PresentationWorker] Processing presentation %s", genMsg.PresentationID)

	// Mark as processing
	_ = w.repo.UpdateStatus(ctx, genMsg.PresentationID, domain.PresentationStatusProcessing, nil)

	// Generate with Kimi K2
	slides, err := w.generateSlides(ctx, &genMsg)
	if err != nil {
		log.Printf("[PresentationWorker] Generation failed for %s: %v", genMsg.PresentationID, err)
		errMsg := err.Error()
		_ = w.repo.UpdateStatus(ctx, genMsg.PresentationID, domain.PresentationStatusFailed, &errMsg)
		msg.Nack(false, true) // Requeue for retry
		return
	}

	// Persist slides
	if err := w.repo.CreateSlidesBatch(ctx, slides); err != nil {
		log.Printf("[PresentationWorker] Failed to save slides for %s: %v", genMsg.PresentationID, err)
		errMsg := "failed to save slides: " + err.Error()
		_ = w.repo.UpdateStatus(ctx, genMsg.PresentationID, domain.PresentationStatusFailed, &errMsg)
		msg.Nack(false, true)
		return
	}

	// Mark completed
	_ = w.repo.UpdateStatus(ctx, genMsg.PresentationID, domain.PresentationStatusCompleted, nil)
	log.Printf("[PresentationWorker] Presentation %s generated successfully (%d slides)", genMsg.PresentationID, len(slides))

	// Send notification
	if w.notification != nil {
		p, err := w.repo.GetByID(ctx, genMsg.PresentationID)
		if err == nil {
			title := "Презентация дайын!"
			message := fmt.Sprintf("«%s» тақырыбындағы презентацияңыз сәтті жасалды.", p.Topic)
			link := fmt.Sprintf("/teacher/materials/slides/%s", genMsg.PresentationID)
			if err := w.notification.CreateNotification(ctx, p.OwnerID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
				log.Printf("[PresentationWorker] Failed to send notification: %v", err)
			}
		}
	}

	msg.Ack(false)
}

func (w *PresentationWorker) generateSlides(ctx context.Context, msg *service.PresentationGenerationMessage) ([]domain.PresentationSlide, error) {
	aiCtx, cancel := context.WithTimeout(ctx, presentationTimeout)
	defer cancel()

	prompt := w.buildPrompt(msg)

	var aiResp aiPresentationResponse
	if err := w.ai.GenerateJSON(aiCtx, prompt, &aiResp); err != nil {
		return nil, fmt.Errorf("kimi AI call failed: %w", err)
	}

	if len(aiResp.Slides) == 0 {
		return nil, fmt.Errorf("kimi returned 0 slides")
	}

	slides := make([]domain.PresentationSlide, 0, len(aiResp.Slides))
	for i, s := range aiResp.Slides {
		slideNum := s.SlideNumber
		if slideNum == 0 {
			slideNum = i + 1
		}
		layout := normalizeLayout(s.Layout)

		// Generate image using ChatGPT (DALL-E 3) if prompt suggests it
		imageURL := ""
		if s.ImagePrompt != "" && (layout == "image_text" || len(s.ImagePrompt) > 10) {
			img, err := w.imageAi.GenerateImage(aiCtx, s.ImagePrompt)
			if err != nil {
				log.Printf("[PresentationWorker] Image generation failed for slide %d: %v", slideNum, err)
				// Fallback or leave empty? User specifically asked for ChatGPT, so we try our best.
			} else {
				imageURL = img
			}
		}

		slides = append(slides, domain.PresentationSlide{
			ID:             uuid.New(),
			PresentationID: msg.PresentationID,
			SlideNumber:    slideNum,
			Title:          s.Title,
			Content:        s.Content,
			Notes:          s.Notes,
			Layout:         layout,
			ImagePrompt:    s.ImagePrompt,
			ImageURL:       imageURL,
		})
	}

	return slides, nil
}


// buildPrompt constructs a multilingual prompt for Kimi K2.
func (w *PresentationWorker) buildPrompt(msg *service.PresentationGenerationMessage) string {
	langMap := map[string]string{
		"kaz": "қазақ тілінде",
		"rus": "на русском языке",
		"eng": "in English",
	}
	langLabel, ok := langMap[msg.Language]
	if !ok {
		langLabel = "in the requested language"
	}

	descBlock := ""
	if strings.TrimSpace(msg.Description) != "" {
		descBlock = fmt.Sprintf("\nAdditional context / requirements: %s", msg.Description)
	}

	style := msg.Style
	if style == "" {
		style = "creative and engaging"
	}

	return fmt.Sprintf(`You are an expert educational content creator and presentation designer.

Create a professional, engaging presentation for a school lesson %s.
Style: %s

Details:
- Subject: %s
- Grade: %d
- Topic: %s
- Number of slides: %d%s

Requirements:
1. Slide 1 MUST be a title slide (layout: "title_only") with only the topic title and subject/grade info.
2. The last slide MUST be a summary/conclusion slide.
3. Each middle slide should cover a specific aspect of the topic with clear bullet points.
4. The "content" field should contain bullet points separated by newline characters (\\n).
5. The "notes" field should include helpful speaker notes for the teacher.
6. Use "layout" values from: title_only | bullets | two_column | image_text
7. Keep content age-appropriate for grade %d students.
8. All text MUST be written %s.
9. For "image_prompt", provide a short, vivid description of an image suitable for the slide. ALWAYS in English (e.g., "A cell structure diagram under a microscope", "Happy students in a classroom"). If layout is "title_only", image prompt should be abstract or thematic.

You MUST respond with ONLY a valid JSON object in this exact format (no markdown, no extra text):
{
  "slides": [
    {
      "slide_number": 1,
      "title": "Slide title here",
      "content": "Bullet point 1\nBullet point 2\nBullet point 3",
      "notes": "Speaker notes for the teacher",
      "layout": "title_only",
      "image_prompt": "A realistic photo of a DNA helix, high quality, 4k"
    }
  ]
}

Generate exactly %d slides.`,
		langLabel,
		style,
		msg.Subject,
		msg.Grade,
		msg.Topic,
		msg.SlideCount,
		descBlock,
		msg.Grade,
		langLabel,
		msg.SlideCount,
	)
}

// normalizeLayout ensures the layout value is one of the valid options.
func normalizeLayout(layout string) string {
	valid := map[string]bool{
		"title_only":  true,
		"bullets":     true,
		"two_column":  true,
		"image_text":  true,
	}
	if valid[layout] {
		return layout
	}
	return "bullets"
}
