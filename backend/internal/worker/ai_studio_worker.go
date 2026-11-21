package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StudioUploadDir is where generated images / videos are written.
// It is served by the API at /api/v1/uploads/studio/...
const StudioUploadDir = "./uploads/studio"

// StudioWorker processes AI Studio requests: images and videos via the OpenAI
// media APIs, and structured text (feedback, lab work, news, certificates)
// via the text generator.
type StudioWorker struct {
	repo         domain.StudioRepository
	notification *service.NotificationService
	text         ai.TextGenerator
	media        *ai.MediaClient
	channel      *amqp.Channel
	sem          chan struct{}
}

func NewStudioWorker(
	repo domain.StudioRepository,
	notification *service.NotificationService,
	text ai.TextGenerator,
	media *ai.MediaClient,
	channel *amqp.Channel,
) *StudioWorker {
	return &StudioWorker{
		repo:         repo,
		notification: notification,
		text:         text,
		media:        media,
		channel:      channel,
		sem:          make(chan struct{}, 4), // at most 4 generations at once
	}
}

func (w *StudioWorker) Start(ctx context.Context) error {
	q, err := w.channel.QueueDeclare(service.StudioQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	_ = w.channel.Qos(4, 0, false)
	msgs, err := w.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("[Studio Worker] Started, listening on queue: %s", q.Name)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			w.sem <- struct{}{}
			go func(m amqp.Delivery) {
				defer func() { <-w.sem }()
				w.handle(ctx, m)
			}(msg)
		}
	}
}

func (w *StudioWorker) handle(ctx context.Context, msg amqp.Delivery) {
	var m service.StudioMessage
	if err := json.Unmarshal(msg.Body, &m); err != nil {
		_ = msg.Nack(false, false)
		return
	}
	item, err := w.repo.GetByID(ctx, m.ItemID)
	if err != nil {
		log.Printf("[Studio Worker] item %s not found: %v", m.ItemID, err)
		_ = msg.Ack(false)
		return
	}
	if item.Status == domain.StudioStatusCompleted {
		_ = msg.Ack(false)
		return
	}

	_ = w.repo.UpdateStatus(ctx, item.ID, domain.StudioStatusProcessing, nil)

	if err := w.process(ctx, item); err != nil {
		log.Printf("[Studio Worker] %s %s failed: %v", item.Kind, item.ID, err)
		e := friendlyError(err)
		_ = w.repo.UpdateStatus(ctx, item.ID, domain.StudioStatusFailed, &e)
	} else {
		log.Printf("[Studio Worker] %s %s completed", item.Kind, item.ID)
		w.notify(ctx, item)
	}
	_ = msg.Ack(false)
}

// friendlyError keeps the provider message but trims it to a readable length.
func friendlyError(err error) string {
	s := err.Error()
	if len(s) > 500 {
		s = s[:500]
	}
	return s
}

func (w *StudioWorker) notify(ctx context.Context, item *domain.StudioItem) {
	if w.notification == nil {
		return
	}
	title := "Материал дайын!"
	message := fmt.Sprintf("«%s» дайын болды.", item.Title)
	link := studioLink(item)
	if err := w.notification.CreateNotification(ctx, item.OwnerID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
		log.Printf("[Studio Worker] notification failed: %v", err)
	}
}

func studioLink(item *domain.StudioItem) string {
	switch item.Kind {
	case domain.StudioKindImage:
		return "/teacher/materials/image"
	case domain.StudioKindVideo:
		return "/teacher/materials/video"
	case domain.StudioKindFeedback:
		return "/teacher/materials/feedback"
	case domain.StudioKindLab:
		return "/teacher/materials/lab"
	case domain.StudioKindNews:
		return "/teacher/materials/news"
	case domain.StudioKindCertificate:
		return "/teacher/materials/certificates"
	}
	return "/teacher"
}

func (w *StudioWorker) process(ctx context.Context, item *domain.StudioItem) error {
	var in map[string]any
	_ = json.Unmarshal(item.Input, &in)
	if in == nil {
		in = map[string]any{}
	}
	switch item.Kind {
	case domain.StudioKindImage:
		return w.processImage(ctx, item, in)
	case domain.StudioKindVideo:
		return w.processVideo(ctx, item, in)
	case domain.StudioKindFeedback, domain.StudioKindLab, domain.StudioKindNews, domain.StudioKindCertificate:
		return w.processText(ctx, item, in)
	}
	return fmt.Errorf("unknown kind %q", item.Kind)
}

func inStr(in map[string]any, key string) string {
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

func writeStudioFile(name string, data []byte) (string, error) {
	if err := os.MkdirAll(StudioUploadDir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(StudioUploadDir, name), data, 0644); err != nil {
		return "", err
	}
	return "/uploads/studio/" + name, nil
}

// ---------------------------------------------------------------- image

func (w *StudioWorker) processImage(ctx context.Context, item *domain.StudioItem, in map[string]any) error {
	prompt := inStr(in, "prompt")
	if style := inStr(in, "style"); style != "" {
		prompt += "\nStyle: " + style
	}
	prompt += "\nThe image is for a school lesson: clear, age-appropriate, no text overlays unless asked."

	cctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()
	data, err := w.media.GenerateImagePNG(cctx, prompt, inStr(in, "size"))
	if err != nil {
		return err
	}
	url, err := writeStudioFile(item.ID.String()+".png", data)
	if err != nil {
		return fmt.Errorf("saving image: %w", err)
	}
	return w.repo.SetResult(ctx, item.ID, nil, &url)
}

// ---------------------------------------------------------------- video

func (w *StudioWorker) processVideo(ctx context.Context, item *domain.StudioItem, in map[string]any) error {
	prompt := inStr(in, "description")
	if t := inStr(in, "title"); t != "" {
		prompt = t + ". " + prompt
	}
	prompt += "\nEducational video for school students, clear and age-appropriate."

	size := "1280x720"
	if inStr(in, "orientation") == "portrait" {
		size = "720x1280"
	}

	// Resume an existing provider job after a worker restart instead of paying twice
	jobID := ""
	if item.ProviderJobID != nil {
		jobID = *item.ProviderJobID
	}
	if jobID == "" {
		job, err := w.media.StartVideo(ctx, prompt, inStr(in, "seconds"), size)
		if err != nil {
			return err
		}
		jobID = job.ID
		_ = w.repo.SetProviderJob(ctx, item.ID, jobID)
	}

	deadline := time.Now().Add(25 * time.Minute)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("video generation timed out")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
		job, err := w.media.GetVideo(ctx, jobID)
		if err != nil {
			log.Printf("[Studio Worker] video %s status check failed: %v", item.ID, err)
			continue
		}
		switch job.Status {
		case "completed":
			if err := os.MkdirAll(StudioUploadDir, 0755); err != nil {
				return err
			}
			name := item.ID.String() + ".mp4"
			f, err := os.Create(filepath.Join(StudioUploadDir, name))
			if err != nil {
				return err
			}
			dlErr := w.media.DownloadVideo(ctx, jobID, f)
			closeErr := f.Close()
			if dlErr != nil {
				return fmt.Errorf("downloading video: %w", dlErr)
			}
			if closeErr != nil {
				return closeErr
			}
			url := "/uploads/studio/" + name
			return w.repo.SetResult(ctx, item.ID, nil, &url)
		case "failed":
			if job.Error != nil && job.Error.Message != "" {
				return fmt.Errorf("%s", job.Error.Message)
			}
			return fmt.Errorf("video generation failed")
		}
	}
}

// ---------------------------------------------------------------- text kinds

func langName(code string) string {
	switch code {
	case "rus":
		return "Russian"
	case "eng":
		return "English"
	}
	return "Kazakh"
}

var feedbackPurposes = map[string]string{
	"exit_ticket":     "Exit ticket: EXACTLY 3 short questions about what was learned in this lesson",
	"self_assessment": "Self-assessment: statements/questions the student uses to rate their own understanding",
	"peer_assessment": "Peer assessment: questions a student answers about a classmate's work",
	"two_stars_wish":  "Two stars and a wish: two prompts for strengths and one prompt for improvement",
	"traffic_light":   "Traffic light: questions where the student picks green/yellow/red for each lesson goal",
	"reflection":      "Lesson reflection: open questions about what was interesting, difficult and useful",
}

func (w *StudioWorker) processText(ctx context.Context, item *domain.StudioItem, in map[string]any) error {
	lang := langName(inStr(in, "language"))
	var prompt string

	switch item.Kind {
	case domain.StudioKindFeedback:
		purpose := inStr(in, "purpose")
		desc, ok := feedbackPurposes[purpose]
		if !ok {
			desc = purpose
		}
		count := inStr(in, "question_count")
		if purpose == "exit_ticket" {
			count = "3"
		}
		if count == "" {
			count = "4"
		}
		prompt = fmt.Sprintf(`You are an experienced teacher. Create a feedback template for students.
Type: %s
Lesson topic: %s
Class: %s
Grade: %s
Number of questions: %s
Write ONLY in %s. Questions must be short, concrete and about THIS lesson topic, understandable for this grade.

Return JSON only:
{
  "title": "template title",
  "instructions": "one sentence for students on how to fill it in",
  "questions": [ { "text": "question text" } ]
}`, desc, inStr(in, "topic"), inStr(in, "class_name"), inStr(in, "grade"), count, lang)

	case domain.StudioKindLab:
		prompt = fmt.Sprintf(`You are a school teacher. Write a laboratory work for students.
Subject: %s
Grade: %s
Topic: %s
Extra wishes: %s
Write ONLY in %s.

Return JSON only:
{
  "title": "lab work title",
  "sections": [
    { "heading": "Goal", "body": "..." },
    { "heading": "Equipment and materials", "body": "- item\n- item" },
    { "heading": "Safety rules", "body": "..." },
    { "heading": "Procedure", "body": "1. ...\n2. ..." },
    { "heading": "Results table / observations", "body": "..." },
    { "heading": "Conclusion questions", "body": "1. ...\n2. ..." },
    { "heading": "Assessment criteria", "body": "..." }
  ]
}
Translate the headings into %s too.`, inStr(in, "subject"), inStr(in, "grade"), inStr(in, "topic"), inStr(in, "details"), lang, lang)

	case domain.StudioKindNews:
		prompt = fmt.Sprintf(`You write news posts for a school website and parent chat.
Topic / event: %s
Details: %s
Audience: %s
Write ONLY in %s. Friendly, factual, 120-220 words, do not invent names, dates or numbers that were not given.

Return JSON only:
{
  "title": "headline",
  "sections": [
    { "heading": "", "body": "news text" },
    { "heading": "short summary for a chat message", "body": "..." }
  ]
}
Translate the headings into %s too.`, inStr(in, "topic"), inStr(in, "details"), inStr(in, "audience"), lang, lang)

	case domain.StudioKindCertificate:
		prompt = fmt.Sprintf(`Write the text of a school certificate (diploma / letter of gratitude).
Student: %s
Achievement: %s
Teacher: %s
School: %s
Write ONLY in %s. Warm and formal, 1-3 sentences, do not invent facts.

Return JSON only:
{
  "title": "certificate heading, e.g. CERTIFICATE / ГРАМОТА / МАҚТАУ ҚАҒАЗЫ",
  "sections": [
    { "heading": "recipient", "body": "student name exactly as given" },
    { "heading": "text", "body": "certificate text" },
    { "heading": "signature", "body": "teacher name and school, as given" }
  ]
}`, inStr(in, "student_name"), inStr(in, "achievement"), inStr(in, "teacher_name"), inStr(in, "school"), lang)
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	var out map[string]any
	if err := w.text.GenerateJSON(cctx, prompt, &out); err != nil {
		return err
	}
	if err := validateStudioContent(item.Kind, out); err != nil {
		return err
	}
	content, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return w.repo.SetResult(ctx, item.ID, content, nil)
}

// validateStudioContent makes sure the AI returned the structure the UI renders.
func validateStudioContent(kind string, out map[string]any) error {
	if kind == domain.StudioKindFeedback {
		qs, ok := out["questions"].([]any)
		if !ok || len(qs) == 0 {
			return fmt.Errorf("AI returned no questions, please retry")
		}
		return nil
	}
	secs, ok := out["sections"].([]any)
	if !ok || len(secs) == 0 {
		return fmt.Errorf("AI returned an empty document, please retry")
	}
	return nil
}
