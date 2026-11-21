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
	"zeinai-backend/internal/service"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	whiteboardQueue   = "whiteboard.summary"
	whiteboardTimeout = 120 * time.Second
)

type whiteboardAIResponse struct {
	SummaryRU     string `json:"summary_ru"`
	ExplanationRU string `json:"explanation_ru"`
	SummaryKK     string `json:"summary_kk"`
	ExplanationKK string `json:"explanation_kk"`
	SummaryEN     string `json:"summary_en"`
	ExplanationEN string `json:"explanation_en"`
}

type whiteboardAction struct {
	Type     string  `json:"type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	IsEraser bool    `json:"isEraser"`
}

type whiteboardSnapshot struct {
	Pages   [][]whiteboardAction `json:"pages"`
	Version int                  `json:"version"`
}

type WhiteboardWorker struct {
	repo         domain.WhiteboardRepository
	notification *service.NotificationService
	classService *service.ClassService
	ai           ai.TextGenerator
	channel      *amqp.Channel
}

func NewWhiteboardWorker(
	repo domain.WhiteboardRepository,
	notification *service.NotificationService,
	classService *service.ClassService,
	aiGen ai.TextGenerator,
	channel *amqp.Channel,
) *WhiteboardWorker {
	return &WhiteboardWorker{
		repo:         repo,
		notification: notification,
		classService: classService,
		ai:           aiGen,
		channel:      channel,
	}
}

func (w *WhiteboardWorker) Start(ctx context.Context) error {
	q, err := w.channel.QueueDeclare(
		whiteboardQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("[WhiteboardWorker] failed to declare queue: %w", err)
	}

	if err := w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("[WhiteboardWorker] failed to set QoS: %w", err)
	}

	msgs, err := w.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("[WhiteboardWorker] failed to consume: %w", err)
	}

	log.Printf("[WhiteboardWorker] Started, listening on queue: %s", q.Name)

	for {
		select {
		case <-ctx.Done():
			log.Println("[WhiteboardWorker] Shutting down...")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[WhiteboardWorker] Channel closed")
				return nil
			}
			w.processMessage(ctx, msg)
		}
	}
}

func (w *WhiteboardWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	var summaryMsg struct {
		SessionID uuid.UUID   `json:"session_id"`
		TeacherID uuid.UUID   `json:"teacher_id"`
		ClassID   int64       `json:"class_id"`
		Title     string      `json:"title"`
		Snapshot  interface{} `json:"snapshot"`
	}

	if err := json.Unmarshal(msg.Body, &summaryMsg); err != nil {
		log.Printf("[WhiteboardWorker] Bad message body: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("[WhiteboardWorker] Generating summary for session %s", summaryMsg.SessionID)

	// Generate with AI
	aiResp, raw, err := w.generateSummary(ctx, summaryMsg.Title, summaryMsg.Snapshot)
	if err != nil {
		log.Printf("[WhiteboardWorker] AI generation failed for %s: %v", summaryMsg.SessionID, err)
		log.Printf("[WhiteboardWorker] Raw response (if any): %.500s", raw)
		
		// Check for authentication error (401/403) - do NOT retry if it's persistent
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "403") {
			log.Printf("[WhiteboardWorker] Persistent Auth error detected for %s, using local fallback.", summaryMsg.SessionID)
			aiResp = &whiteboardAIResponse{
				SummaryRU: "### Конспект урока (AI недоступен)\n\n" +
					"Урок успешно завершен. Автоматический конспект временно недоступен из-за ограничений AI-инструмента.\n\n" +
					"**Что было на уроке:**\n- Записаны действия на доске.\n- Урок сохранен в истории.",
				ExplanationRU: "Объяснение недоступно, так как AI-сервис временно отключен или требует настройки ключа доступа.",
				SummaryKK:     "### Сабақ конспектісі (AI қолжетімсіз)\n\nСабақ сәтті аяқталды. Автоматты конспект уақытша қолжетімсіз.",
				ExplanationKK: "Түсіндірме қолжетімсіз.",
				SummaryEN:     "### Lesson Summary (AI Unavailable)\n\nLesson completed successfully. Auto-summary is temporarily unavailable.",
				ExplanationEN: "Explanation unavailable.",
			}
		} else {
			// Other transient errors - can retry
			msg.Nack(false, true)
			return
		}
	}

	// Save to DB
	summary := &domain.WhiteboardSummary{
		ID:            uuid.New(),
		SessionID:     summaryMsg.SessionID,
		SummaryText:   aiResp.SummaryRU,
		SummaryKK:     aiResp.SummaryKK,
		SummaryEN:     aiResp.SummaryEN,
		Explanation:   aiResp.ExplanationRU,
		ExplanationKK: aiResp.ExplanationKK,
		ExplanationEN: aiResp.ExplanationEN,
		CreatedAt:     time.Now(),
	}

	if err := w.repo.CreateSummary(ctx, summary); err != nil {
		log.Printf("[WhiteboardWorker] Failed to save summary for %s: %v", summaryMsg.SessionID, err)
		msg.Nack(false, true)
		return
	}

	// Notify students
	go func() {
		bgCtx := context.Background()
		students, err := w.classService.ListStudents(bgCtx, summaryMsg.ClassID)
		if err != nil {
			log.Printf("[WhiteboardWorker] Failed to list students for notification: %v", err)
			return
		}

		log.Printf("[WhiteboardWorker] Sending summary notifications to %d students", len(students))
		for _, student := range students {
			err := w.notification.CreateNotification(
				bgCtx,
				student.ID,
				nil,
				"System",
				"notif_whiteboard_summary_ready_title",
				"notif_whiteboard_summary_ready_msg",
				domain.NotificationWhiteboardSummaryReady,
				fmt.Sprintf("/student/whiteboard/summary/%s", summaryMsg.SessionID),
			)
			if err != nil {
				log.Printf("[WhiteboardWorker] Failed to send summary notification to student %s: %v", student.ID, err)
			}
		}
	}()

	msg.Ack(false)
}

func (w *WhiteboardWorker) generateSummary(ctx context.Context, title string, snapshot interface{}) (*whiteboardAIResponse, string, error) {
	aiCtx, cancel := context.WithTimeout(ctx, whiteboardTimeout)
	defer cancel()

	// Convert snapshot to simplified SVG for AI understanding
	svgData := w.snapshotToSVG(snapshot)

	prompt := fmt.Sprintf(`You are a Lead Teacher and Pedagogical Expert at Zeiyn.kz EdTech platform.
Your goal is to transform whiteboard notes into a high-quality, professional lesson summary (konspekt) that feels like it was written by an experienced teacher.

CORE CONTEXT:
The lesson topic is: "%s"

TASK:
1. Examine the provided "Whiteboard Data" which contains simplified SVG paths of handwritten notes.
2. "Read" the handwriting to extract key terms, formulas, diagrams, and concepts.
3. Don't just transcribe what is written. Use the notes as anchor points to provide a DEEP, EDUCATIONAL explanation of the topic.
4. If the teacher wrote a short term (e.g., "Photosynthesis"), you should provide a clear definition and explain the process as a teacher would.
5. If there are formulas, explain their components and meaning.

PEDAGOGICAL REQUIREMENTS:
- TONE: Professional, encouraging, and clear.
- STRUCTURE: Use clear headings, bullet points, and highlight important terms.
- DEPTH: Provide the "why" and "how", not just the "what". The explanation should help a student who missed the class to fully understand the topic.
- LANGUAGE: Ensure the pedagogical terminology is accurate in each language.

CRITICAL INSTRUCTIONS:
1. Generate the response in THREE separate languages: Russian (RU), Kazakh (KK), and English (EN).
2. The field "summary_ru" MUST contain a structured lesson summary in Russian.
3. The field "explanation_ru" MUST contain a warm, teacher-like explanation/message to the student in Russian.
4. Follow the same for Kazakh ("summary_kk", "explanation_kk") and English ("summary_en", "explanation_en").

JSON EXPECTED FORMAT:
{
  "summary_ru": "### Тема: %s\\n\\n#### Основные понятия:\\n- **[Понятие]**: [Глубокое объяснение]...\\n\\n#### Детализация:\\n[Подробный разбор материала на основе записей]...",
  "explanation_ru": "Привет! На этом уроке мы разобрали очень важную тему: %s. Главное, что тебе нужно запомнить — это...",
  "summary_kk": "### Тақырыбы: %s\\n\\n#### Негізгі ұғымдар:\\n- **[Ұғым]**: [Терең түсіндірме]...\\n\\n#### Толық мәлімет:\\n[Жазбалар негізіндегі толық талдау]...",
  "explanation_kk": "Сәлем! Бұл сабақта біз %s тақырыбын талқыладық. Ең бастысы...",
  "summary_en": "### Topic: %s\\n\\n#### Key Concepts:\\n- **[Concept]**: [Deep explanation]...\\n\\n#### Detailed Breakdown:\\n[Comprehensive analysis based on notes]...",
  "explanation_en": "Hello! In this lesson, we covered the topic of %s. The most important thing to remember is..."
}

Whiteboard Data (Simplified SVG Paths):
%s`, title, title, title, title, title, title, title, svgData)

	var aiResp whiteboardAIResponse
	if err := w.ai.GenerateJSON(aiCtx, prompt, &aiResp); err != nil {
		return nil, "", err
	}

	return &aiResp, "", nil
}

func (w *WhiteboardWorker) snapshotToSVG(snapshot interface{}) string {
	bytes, err := json.Marshal(snapshot)
	if err != nil {
		return "Error serializing snapshot"
	}

	var s whiteboardSnapshot
	if err := json.Unmarshal(bytes, &s); err != nil {
		return "Empty or invalid snapshot"
	}

	var sb strings.Builder
	for pIdx, page := range s.Pages {
		if len(page) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("Page %d:\n", pIdx+1))

		var path strings.Builder
		count := 0
		pointsInStroke := 0

		for _, action := range page {
			if action.IsEraser {
				continue
			}

			switch action.Type {
			case "start":
				if pointsInStroke > 0 {
					sb.WriteString(path.String() + "\n")
					path.Reset()
				}
				path.WriteString(fmt.Sprintf("M %.1f %.1f", action.X, action.Y))
				pointsInStroke = 1
				count++
			case "move":
				// Downsample: only add every 3rd point to save tokens
				if pointsInStroke%3 == 0 {
					path.WriteString(fmt.Sprintf(" L %.1f %.1f", action.X, action.Y))
				}
				pointsInStroke++
			case "end":
				if pointsInStroke > 0 {
					sb.WriteString(path.String() + "\n")
					path.Reset()
					pointsInStroke = 0
				}
			}

			// Safety limit: 100 paths per page to avoid token overflow
			if count > 100 {
				sb.WriteString("... (too many strokes)\n")
				break
			}
		}
		// Final stroke if any
		if pointsInStroke > 0 {
			sb.WriteString(path.String() + "\n")
		}
	}

	res := sb.String()
	if res == "" {
		return "No visible strokes found on the board."
	}
	return res
}
