package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/repository"
	"zeinai-backend/platform/rabbitmq"

	"github.com/google/uuid"
)

type STPWorker struct {
	repo repository.STPRepository
	ai   ai.TextGenerator
	cfg  *config.Config
}

func NewSTPWorker(repo repository.STPRepository, ai ai.TextGenerator, cfg *config.Config) *STPWorker {
	return &STPWorker{
		repo: repo,
		ai:   ai,
		cfg:  cfg,
	}
}

func (w *STPWorker) Start() {
	log.Println("Starting STP Worker...")
	ch := rabbitmq.GetChannel()
	if ch == nil {
		log.Println("RabbitMQ channel not ready, STP worker could not start")
		return
	}

	msgs, err := ch.Consume(
		"ai.stp", // queue
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register STP consumer: %v", err)
		return
	}

	go func() {
		for d := range msgs {
			log.Printf("STP Worker received a message: %s", d.Body)

			var payload map[string]string
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("Error unmarshalling STP message: %v", err)
				d.Ack(false)
				continue
			}

			idStr, ok := payload["id"]
			if !ok {
				d.Ack(false)
				continue
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				d.Ack(false)
				continue
			}

			if err := w.processSTP(id); err != nil {
				log.Printf("STP Generation failed for %s: %v", id, err)
				d.Nack(false, true) // Requeue
			} else {
				d.Ack(false)
			}
		}
	}()
}

func (w *STPWorker) processSTP(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Fetch record
	stp, err := w.repo.GetSTPByID(ctx, id)
	if err != nil {
		return err
	}

	stp.Status = domain.AIStatusProcessing
	w.repo.UpdateSTP(ctx, stp)

	// 2. Prepare Prompt
	prompt := w.constructPrompt(stp)

	// 3. Call AI
	respText, err := w.ai.GenerateText(ctx, prompt)
	if err != nil {
		errMsg := err.Error()
		stp.Status = domain.AIStatusFailed
		stp.Content = errMsg
		w.repo.UpdateSTP(ctx, stp)
		return err
	}

	// 4. Save result
	stp.Content = respText
	stp.Status = domain.AIStatusCompleted
	stp.UpdatedAt = time.Now()

	// Add terms/explanations if we can extract them or if AI part of content
	// For now, AI will generate everything in one text block

	return w.repo.UpdateSTP(ctx, stp)
}

func (w *STPWorker) constructPrompt(stp *domain.ShortTermPlan) string {
	var methodNames []string
	json.Unmarshal(stp.SelectedMethods, &methodNames)
	methodsStr := strings.Join(methodNames, ", ")

	return fmt.Sprintf(`
Сен Қазақстанның озық педагог-сарапшысысың. Төмендегі параметрлер бойынша "Қысқа мерзімді жоспар" (ҚМЖ) жасап бер. 
Жоспар қазақ тілінде, мұғалімдерге арналған ресми стильде болуы керек.

Пәні: %s
Сыныбы: %d (%s)
Сабақ тақырыбы: %s
Оқу мақсаттары: %s
Сабақ мақсаттары: %s
Айдың құндылығы: %s
Таңдалған әдіс-тәсілдер: %s

Құрылымы келесідей болсын (Markdown форматындағы кесте):
1. Сабақтың басы (10 мин): Психологиялық ахуал, үй жұмысын сұрау, қызығушылықты ояту. Таңдалған әдістердің кемінде біреуін қолдан.
2. Сабақтың ортасы (30 мин): Жаңа тақырыпты түсіндіру, тәжірибелік жұмыстар, тапсырмалар (жеке, жұптық, топтық). Саралау (дифференциация) міндетті түрде болсын.
3. Сабақтың соңы (5 мин): Кері байланыс, бағалау, үй тапсырмасы.

Маңызды талаптар:
- Әр кезеңде Педагогтің әрекеті, Оқушының әрекеті, Бағалау және Ресурстар бағандары болсын.
- Сабақ мазмұны оқушылардың жас ерекшелігіне (%s) сәйкес, түсінікті және қызықты болсын.
- Айдың құндылығын (%s) сабақтың тәрбиелік бөлігі ретінде міндетті түрде кіріктір.

Сабақ жоспарынан кейін міндетті түрде мынадай екі блок қосыңыз:
1. "Әдіс-тәсілдерге түсініктеме": Сабақта қолданылған %s әдістерін қалай тиімді пайдалану керек екенін мұғалімге сипаттап жаз.
2. "Жаңа терминдер мен ұғымдар": Сабақ тақырыбына қатысты 5-7 негізгі терминді және олардың анықтамасын бер.

Тек нәтижені (Markdown форматында) қайтар.
`,
		stp.Subject, stp.Grade, stp.AgeRange, stp.Topic,
		stp.LearningObjectives, stp.LessonObjectives,
		stp.MonthlyValue, methodsStr,
		stp.AgeRange, stp.MonthlyValue, methodsStr)
}
