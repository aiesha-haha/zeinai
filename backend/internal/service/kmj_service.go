package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/repository"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type KMJService interface {
	// Plan operations
	CreateDraftPlan(ctx context.Context, req *CreateKMJRequest, userID uuid.UUID, schoolID *uuid.UUID) (*domain.KMJPlanWithCells, error)
	GenerateKMJ(ctx context.Context, planID uuid.UUID, userID uuid.UUID) error
	GetPlanByID(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (*domain.KMJPlanWithCells, error)
	ListMyPlans(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedKMJPlans, error)
	DeletePlan(ctx context.Context, planID uuid.UUID, userID uuid.UUID) error

	// Cell operations
	UpdateCell(ctx context.Context, cellID uuid.UUID, cellValue string, userID uuid.UUID) error

	// Generation callback (called by worker)
	HandleGenerationComplete(ctx context.Context, taskID uuid.UUID, cells []domain.KMJCell) error
	HandleGenerationFailed(ctx context.Context, taskID uuid.UUID, errorMsg string) error

	// Data fetching for frontend
	ListTeachingMethods(ctx context.Context) ([]domain.TeachingMethod, error)
	GetNationalValueByMonth(ctx context.Context, month int) ([]domain.NationalValue, error)
}

type kmjService struct {
	kmjRepo           repository.KMJRepository
	channel           *amqp.Channel
	subscriptionRepo  domain.SubscriptionRepository
	nationalValueRepo repository.NationalValueRepository
	methodRepo        repository.TeachingMethodRepository
}

func NewKMJService(
	kmjRepo repository.KMJRepository,
	channel *amqp.Channel,
	subscriptionRepo domain.SubscriptionRepository,
	nationalValueRepo repository.NationalValueRepository,
	methodRepo repository.TeachingMethodRepository,
) KMJService {
	return &kmjService{
		kmjRepo:           kmjRepo,
		channel:           channel,
		subscriptionRepo:  subscriptionRepo,
		nationalValueRepo: nationalValueRepo,
		methodRepo:        methodRepo,
	}
}

// ==================== DTOs ====================

type CreateKMJRequest struct {
	Subject      string      `json:"subject" validate:"required,max=255"`
	Grade        int         `json:"grade" validate:"required,min=1,max=12"`
	LessonNumber int         `json:"lesson_number" validate:"required,min=1"`
	LessonTopic  string      `json:"lesson_topic" validate:"required,max=500"`
	TeacherName  string      `json:"teacher_name" validate:"required,max=255"`
	LessonDate   *time.Time  `json:"lesson_date,omitempty"`
	ContentLang  string      `json:"content_lang" validate:"required,oneof=kaz rus eng"`
	MethodIDs    []uuid.UUID `json:"method_ids,omitempty" validate:"omitempty,max=4,dive,required"`
	// Learning objective typed by the teacher (curriculum code + text)
	LearningObjective string `json:"learning_objective,omitempty" validate:"omitempty,max=2000"`
	// Total lesson length in minutes; 0 means the default 45
	DurationMinutes int `json:"duration_minutes,omitempty" validate:"omitempty,min=20,max=120"`
}

type PaginatedKMJPlans struct {
	Plans      []domain.KMJPlan `json:"plans"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type KMJGenerationMessage struct {
	TaskID   uuid.UUID  `json:"task_id"`
	PlanID   uuid.UUID  `json:"plan_id"`
	UserID   uuid.UUID  `json:"user_id"`
	SchoolID *uuid.UUID `json:"school_id,omitempty"`
	Prompt   string     `json:"prompt"`
	Language string     `json:"language"`
}

// ==================== Implementation ====================

func (s *kmjService) CreateDraftPlan(ctx context.Context, req *CreateKMJRequest, userID uuid.UUID, schoolID *uuid.UUID) (*domain.KMJPlanWithCells, error) {
	// Determine month for national value selection
	var month int
	if req.LessonDate != nil {
		month = int(req.LessonDate.Month())
	} else {
		// Use current month if no date provided
		month = int(time.Now().Month())
	}

	// Auto-select national value for the month
	values, err := s.nationalValueRepo.GetByMonth(ctx, month)
	var nationalValueID *uuid.UUID
	if err == nil && len(values) > 0 {
		// Use first value for the month
		nationalValueID = &values[0].ID
		log.Printf("[KMJ] Auto-selected national value: %s for month %d", values[0].NameKaz, month)
	} else {
		log.Printf("[KMJ] No national value found for month %d", month)
	}

	// Create plan
	plan := &domain.KMJPlan{
		ID:              uuid.New(),
		OwnerID:         userID,
		SchoolID:        schoolID,
		Subject:         req.Subject,
		Grade:           req.Grade,
		LessonNumber:    req.LessonNumber,
		LessonTopic:     req.LessonTopic,
		TeacherName:     req.TeacherName,
		LessonDate:      req.LessonDate,
		ContentLang:     req.ContentLang,
		Status:          domain.KMJStatusDraft,
		NationalValueID: nationalValueID, // Set selected value

		LearningObjective: strings.TrimSpace(req.LearningObjective),
		DurationMinutes:   req.DurationMinutes,
	}
	if plan.DurationMinutes == 0 {
		plan.DurationMinutes = domain.KMJDefaultDuration
	}

	err = s.kmjRepo.CreatePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	// Create empty structure cells with data
	cells := s.createEmptyStructure(plan)
	err = s.kmjRepo.CreateCellsBatch(ctx, cells)
	if err != nil {
		return nil, fmt.Errorf("failed to create cells: %w", err)
	}

	// Add selected teaching methods (max 4)
	if len(req.MethodIDs) > 0 {
		err = s.methodRepo.SetPlanMethods(ctx, plan.ID, req.MethodIDs)
		if err != nil {
			log.Printf("[KMJ] Failed to add methods: %v", err)
			// Not critical - continue
		} else {
			log.Printf("[KMJ] Added %d teaching methods to plan %s", len(req.MethodIDs), plan.ID)
		}
	}

	// Fetch national value details if set
	var nationalValue *domain.NationalValue
	if plan.NationalValueID != nil {
		nv, err := s.nationalValueRepo.GetByID(ctx, *plan.NationalValueID)
		if err == nil {
			nationalValue = nv
		}
	}

	return &domain.KMJPlanWithCells{
		Plan:          *plan,
		Cells:         cells,
		NationalValue: nationalValue,
	}, nil
}

func (s *kmjService) GenerateKMJ(ctx context.Context, planID uuid.UUID, userID uuid.UUID) error {
	// Check ownership
	isOwner, err := s.kmjRepo.CheckPlanOwnership(ctx, planID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return fmt.Errorf("unauthorized: not plan owner")
	}

	// Get plan
	plan, err := s.kmjRepo.GetPlanByID(ctx, planID)
	if err != nil {
		return err
	}

	// A failed generation can be retried; a plan that is generating or done cannot.
	if plan.Status != domain.KMJStatusDraft && plan.Status != domain.KMJStatusFailed {
		return fmt.Errorf("plan is not in draft status")
	}

	// Check and consume tokens (5 tokens required for KMJ generation)
	// Consume tokens - this will fail if user doesn't have enough tokens
	err = s.subscriptionRepo.ConsumeTokens(ctx, userID, domain.TokenCostKMJGeneration, domain.ActionTypeKMJGeneration)
	if err != nil {
		log.Printf("[KMJ] Token consumption failed for user %s: %v", userID, err)
		return fmt.Errorf("insufficient tokens: %w", err)
	}

	log.Printf("[KMJ] %d tokens consumed for user %s (KMJ generation)", domain.TokenCostKMJGeneration, userID)

	// Create generation task
	task := &domain.KMJGenerationTask{
		ID:       uuid.New(),
		PlanID:   planID,
		UserID:   userID,
		SchoolID: plan.SchoolID,
		Prompt:   s.buildPrompt(plan),
		Language: plan.ContentLang,
		Status:   domain.TaskStatusPending,
	}

	err = s.kmjRepo.CreateGenerationTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Update plan status
	err = s.kmjRepo.UpdatePlanStatus(ctx, planID, domain.KMJStatusGenerating)
	if err != nil {
		return err
	}

	// Send to queue
	msg := KMJGenerationMessage{
		TaskID:   task.ID,
		PlanID:   planID,
		UserID:   userID,
		SchoolID: plan.SchoolID,
		Prompt:   task.Prompt,
		Language: plan.ContentLang,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish directly to RabbitMQ
	err = s.channel.PublishWithContext(
		ctx,
		"",               // exchange
		"kmj.generation", // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
	if err != nil {
		// Rollback status
		_ = s.kmjRepo.UpdatePlanStatus(ctx, planID, domain.KMJStatusDraft)
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	log.Printf("[KMJ] Generation task %s queued for plan %s (cost: %d tokens)", task.ID, planID, domain.TokenCostKMJGeneration)

	return nil
}

func (s *kmjService) GetPlanByID(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (*domain.KMJPlanWithCells, error) {
	// Check ownership
	isOwner, err := s.kmjRepo.CheckPlanOwnership(ctx, planID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, fmt.Errorf("unauthorized: not plan owner")
	}

	result, err := s.kmjRepo.GetPlanWithCells(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Enrich with national value details
	if result.Plan.NationalValueID != nil {
		nv, err := s.nationalValueRepo.GetByID(ctx, *result.Plan.NationalValueID)
		if err == nil {
			result.NationalValue = nv
		}
	}

	return result, nil
}

func (s *kmjService) ListMyPlans(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedKMJPlans, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	plans, total, err := s.kmjRepo.ListPlansByOwner(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &PaginatedKMJPlans{
		Plans:      plans,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *kmjService) DeletePlan(ctx context.Context, planID uuid.UUID, userID uuid.UUID) error {
	// Check ownership
	isOwner, err := s.kmjRepo.CheckPlanOwnership(ctx, planID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return fmt.Errorf("unauthorized: not plan owner")
	}

	return s.kmjRepo.SoftDeletePlan(ctx, planID)
}

func (s *kmjService) UpdateCell(ctx context.Context, cellID uuid.UUID, cellValue string, userID uuid.UUID) error {
	// Ownership check: cell -> plan -> owner
	planID, err := s.kmjRepo.GetCellPlanID(ctx, cellID)
	if err != nil {
		return err
	}
	isOwner, err := s.kmjRepo.CheckPlanOwnership(ctx, planID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return fmt.Errorf("unauthorized: not plan owner")
	}
	return s.kmjRepo.UpdateCell(ctx, cellID, cellValue)
}

// ==================== Generation Callbacks ====================

func (s *kmjService) HandleGenerationComplete(ctx context.Context, taskID uuid.UUID, cells []domain.KMJCell) error {
	task, err := s.kmjRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Update cells
	err = s.kmjRepo.CreateCellsBatch(ctx, cells)
	if err != nil {
		return fmt.Errorf("failed to save cells: %w", err)
	}

	// Complete task
	err = s.kmjRepo.CompleteTask(ctx, taskID, nil)
	if err != nil {
		return err
	}

	// Update plan status
	err = s.kmjRepo.UpdatePlanStatus(ctx, task.PlanID, domain.KMJStatusCompleted)
	if err != nil {
		return err
	}

	log.Printf("[KMJ] Generation completed for plan %s", task.PlanID)

	return nil
}

func (s *kmjService) HandleGenerationFailed(ctx context.Context, taskID uuid.UUID, errorMsg string) error {
	task, err := s.kmjRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Update task
	err = s.kmjRepo.UpdateTaskStatus(ctx, taskID, domain.TaskStatusFailed, &errorMsg)
	if err != nil {
		return err
	}

	// Update plan status
	err = s.kmjRepo.UpdatePlanStatus(ctx, task.PlanID, domain.KMJStatusFailed)
	if err != nil {
		return err
	}

	log.Printf("[KMJ] Generation failed for plan %s: %s", task.PlanID, errorMsg)

	return nil
}

// ==================== Helpers ====================

func (s *kmjService) buildPrompt(plan *domain.KMJPlan) string {
	lessonDateStr := ""
	if plan.LessonDate != nil {
		lessonDateStr = plan.LessonDate.Format("02.01.2006")
	}

	lang := plan.ContentLang
	if lang == "" {
		lang = "kaz"
	}

	// Localized labels for prompt
	labels := map[string]map[string]string{
		"kaz": {
			"subject":     "Пән",
			"grade":       "Сынып",
			"lesson_num":  "Сабақ нөмірі",
			"topic":       "Сабақ тақырыбы",
			"teacher":     "Мұғалім аты-жөні",
			"date":        "Сабақ күні",
			"write_only":  "Тек қазақ тілінде жаз.",
			"nv_title":    "ҰЛТТЫҚ ҚҰНДЫЛЫҚ (%s айы):",
			"nv_name":     "Құндылық атауы",
			"nv_desc":     "Сипаттама",
			"nv_instr":    "Сабақ барысында (әсіресе тапсырмаларда) осы құндылықты қалай қолданатыныңды нақты көрсет. JSON-дағы \"unity_cooperation\" өрісіне құндылықтың атын жазба, ТЕК оның сабақта іс жүзінде қалай көрініс табатынын жаз.",
			"methods_title": "ӘДІС-ТӘСІЛДЕР (Міндетті түрде қолданыңыз):",
			"methods_instr": "Осы әдістерді сабақтың ортасында топтық немесе жұптық жұмыстар арқылы жүзеге асыруды сипаттаңыз.",
			"role":        "Сен кәсіби ұстаз және білім беру сарапшысысың. Саған келесі мәліметтер бойынша толық Қысқа мерзімді жоспар (ҚМЖ) құру керек.",
			"task_1":      "1. Сабақ жоспарын өте толық, кәсіби және сапалы түрде құр.",
			"task_2":      "2. Сабақтың ортасында (middle) міндетті түрде тақырыпқа сәйкес АНЫҚТАМАЛАР мен теориялық түсініктерді бер, топтық/жұптық тапсырмалар құрастыр.",
			"task_3":      "3. Ұлттық құндылықты (National Value) сабақтың басында мотивация ретінде немесе ортасында тапсырма арқылы табиғи түрде қолданыңыз.",
			"objective":   "Оқу мақсаты (мұғалім енгізген)",
			"duration":    "Сабақ ұзақтығы",
			"timing":      "4. Сабақ кезеңдерінің уақыты МІНДЕТТІ: басы - %d мин, ортасы - %d мин, соңы - %d мин (барлығы %d мин). Әр кезеңдегі әрекеттер осы уақытқа сыюы керек; ортасындағы әр тапсырманың жанына минутын жаз.",
			"tasks":       "5. Әр кезеңде оқушыға берілетін нақты тапсырманы, оны бағалау тәсілін (дескриптор) және қажетті ресурстарды жаз. Сабақ мақсаты оқу мақсатынан туындауы керек.",
		},
		"rus": {
			"subject":     "Предмет",
			"grade":       "Класс",
			"lesson_num":  "Номер урока",
			"topic":       "Тема урока",
			"teacher":     "ФИО учителя",
			"date":        "Дата урока",
			"write_only":  "Пиши только на русском языке.",
			"nv_title":    "НАЦИОНАЛЬНАЯ ЦЕННОСТЬ (месяц - %s):",
			"nv_name":     "Название ценности",
			"nv_desc":     "Описание",
			"nv_instr":    "Покажи в ходе урока (особенно в заданиях), как эта ценность применяется. В поле \"unity_cooperation\" пиши не название ценности, а то, как она реализуется на практике.",
			"methods_title": "МЕТОДЫ ОБУЧЕНИЯ (Обязательно к использованию):",
			"methods_instr": "Опишите реализацию этих методов в середине урока через групповую или парную работу.",
			"role":        "Ты профессиональный учитель и эксперт в образовании. Тебе нужно создать подробный Краткосрочный план (КСП) по следующим данным.",
			"task_1":      "1. Создай очень подробный, профессиональный и качественный план урока.",
			"task_2":      "2. В середине урока (middle) обязательно дай ОПРЕДЕЛЕНИЯ и теорию, составь задания для парной/групповой работы.",
			"task_3":      "3. Используй национальную ценность (National Value) естественным образом через мотивацию или задания.",
			"objective":   "Цель обучения (указана учителем)",
			"duration":    "Длительность урока",
			"timing":      "4. Время этапов ОБЯЗАТЕЛЬНО: начало - %d мин, середина - %d мин, конец - %d мин (всего %d мин). Действия каждого этапа должны укладываться в это время; рядом с каждым заданием середины укажи минуты.",
			"tasks":       "5. Для каждого этапа укажи конкретное задание для учеников, способ оценивания (дескрипторы) и ресурсы. Цели урока должны вытекать из цели обучения.",
		},
		"eng": {
			"subject":     "Subject",
			"grade":       "Grade",
			"lesson_num":  "Lesson Number",
			"topic":       "Lesson Topic",
			"teacher":     "Teacher Name",
			"date":        "Lesson Date",
			"write_only":  "Write only in English.",
			"nv_title":    "NATIONAL VALUE (month - %s):",
			"nv_name":     "Value Name",
			"nv_desc":     "Description",
			"nv_instr":    "Show specifically in the lesson (especially tasks) how this value is applied. In the \"unity_cooperation\" field, don't write the value name, write ONLY how it actually manifests in the lesson.",
			"methods_title": "TEACHING METHODS (Mandatory to use):",
			"methods_instr": "Describe realization of these methods in the middle of the lesson through group or pair work.",
			"role":        "You are a professional teacher and education expert. You need to create a complete Short-Term Plan (STP) based on the following details.",
			"task_1":      "1. Create a very detailed, professional, and high-quality lesson plan.",
			"task_2":      "2. In the middle of the lesson, ensure you provide DEFINITIONS and theory, create tasks for group/pair work.",
			"task_3":      "3. Use national value naturally through motivation at start or tasks in middle.",
			"objective":   "Learning objective (given by the teacher)",
			"duration":    "Lesson duration",
			"timing":      "4. Stage timing is MANDATORY: start - %d min, middle - %d min, end - %d min (total %d min). Activities of each stage must fit this time; write the minutes next to each task in the middle stage.",
			"tasks":       "5. For every stage give the concrete student task, the assessment method (descriptors) and the resources. Lesson objectives must follow from the learning objective.",
		},
	}

	t := labels[lang]
	if t == nil {
		t = labels["kaz"]
	}

	// Month names localization fallback
	monthNamesMap := map[string]map[int]string{
		"kaz": domain.MonthNamesKaz,
		"rus": {1: "Январь", 2: "Февраль", 3: "Март", 4: "Апрель", 5: "Май", 6: "Июнь", 7: "Июль", 8: "Август", 9: "Сентябрь", 10: "Октябрь", 11: "Ноябрь", 12: "Декабрь"},
		"eng": {1: "January", 2: "February", 3: "March", 4: "April", 5: "May", 6: "June", 7: "July", 8: "August", 9: "September", 10: "October", 11: "November", 12: "December"},
	}
	monthNames := monthNamesMap[lang]
	if monthNames == nil {
		monthNames = monthNamesMap["kaz"]
	}

	// National Value block
	var valueInstruction string
	if plan.NationalValueID != nil {
		value, err := s.nationalValueRepo.GetByID(context.Background(), *plan.NationalValueID)
		if err == nil {
			valueName := value.GetLocalizedName(lang)
			valueDesc := value.GetLocalizedDescription(lang)
			monthName := monthNames[value.MonthNumber]

			valueInstruction = fmt.Sprintf(`
%s
%s: %s
%s: %s

%s
`, fmt.Sprintf(t["nv_title"], monthName), t["nv_name"], valueName, t["nv_desc"], valueDesc, t["nv_instr"])
		}
	}

	// Methods block
	var methodsInstruction string
	methods, err := s.methodRepo.GetPlanMethods(context.Background(), plan.ID)
	if err == nil && len(methods) > 0 {
		methodNames := make([]string, len(methods))
		for i, m := range methods {
			methodNames[i] = fmt.Sprintf("%d. %s", i+1, m.GetLocalizedName(lang))
			if m.GetLocalizedPurpose(lang) != "" {
				methodNames[i] += fmt.Sprintf(" - %s", m.GetLocalizedPurpose(lang))
			}
		}

		methodsInstruction = fmt.Sprintf(`
%s
%s

%s
`, t["methods_title"], strings.Join(methodNames, "\n"), t["methods_instr"])
	}

	duration := plan.DurationMinutes
	if duration <= 0 {
		duration = domain.KMJDefaultDuration
	}
	startMin, middleMin, endMin := domain.KMJStageMinutes(duration)
	timing := fmt.Sprintf(t["timing"], startMin, middleMin, endMin, duration)

	objectiveLine := ""
	if strings.TrimSpace(plan.LearningObjective) != "" {
		objectiveLine = fmt.Sprintf("- %s: %s\n", t["objective"], plan.LearningObjective)
	}

	return fmt.Sprintf(`
%s

%s:
- %s: %s
- %s: %d
- %s: %d
- %s: %s
- %s: %s
- %s: %s
- %s: %d min
%s%s%s

%s:
%s
%s
%s
%s
%s

IMPORTANT: 
- Your response MUST be in JSON format only.
- Write actions in depth (at least 5-6 bullet points per stage).
- %s

JSON STRUCTURE:
{
  "header": {
    "bolem": "Subject name",
    "pedagog_name": "Teacher name",
    "date": "Date",
    "grade": "Grade level",
    "topic": "Topic",
    "learning_objectives": "Learning objectives with codes (keep the teacher's objective if given)",
    "lesson_objectives": "Detailed lesson objectives",
    "unity_cooperation": "Description of national value integration",
    "participants_count": "Write 'Full attendance' or similar placeholder",
    "absent_count": "Write '0' or similar placeholder"
  },
  "stages": {
    "start": {
      "pedagog_action": "Teacher activity (greeting, motivation, homework check)",
      "student_action": "Student activity",
      "assessment": "Formative assessment markers/descriptors",
      "resources": "Resources used"
    },
    "middle": {
      "pedagog_action": "Explanation of new topic, providing DEFINITIONS, theory, and giving GROUP/PAIR tasks using chosen METHODS",
      "student_action": "Learning definitions, completing group/pair tasks, discussion, practice",
      "assessment": "Specific descriptors for each task and assessment method",
      "resources": "Textbook, slides, task sheets, etc."
    },
    "end": {
      "pedagog_action": "Reflection, feedback, homework assignment",
      "student_action": "Doing reflection, writing down homework",
      "assessment": "Final summary assessment",
      "resources": "Resources"
    }
  },
  "footer": {
    "assessment": "General assessment criteria and descriptors for the lesson",
    "homework": "Specific homework content"
  }
}
`, t["role"], "DETAILS", t["subject"], plan.Subject, t["grade"], plan.Grade, t["lesson_num"], plan.LessonNumber, t["topic"], plan.LessonTopic, t["teacher"], plan.TeacherName, t["date"], lessonDateStr, t["duration"], duration, objectiveLine, valueInstruction, methodsInstruction,
		"TASK", t["task_1"], t["task_2"], t["task_3"], timing, t["tasks"], t["write_only"])
}

func (s *kmjService) createEmptyStructure(plan *domain.KMJPlan) []domain.KMJCell {
	cells := []domain.KMJCell{}

	lessonDateStr := ""
	if plan.LessonDate != nil {
		lessonDateStr = plan.LessonDate.Format("02.01.2006")
	}

	// Header cells - populated with initial data
	headerCells := []struct {
		key   string
		value string
	}{
		{"bolem", plan.Subject},
		{"pedagog_name", plan.TeacherName},
		{"date", lessonDateStr},
		{"grade", fmt.Sprintf("%d", plan.Grade)},
		{"topic", plan.LessonTopic},
		{"learning_objectives", plan.LearningObjective},
		{"lesson_objectives", ""},
		{"unity_cooperation", ""},
		{"participants_count", ""},
		{"absent_count", ""},
	}

	for _, hc := range headerCells {
		cells = append(cells, domain.KMJCell{
			ID:            uuid.New(),
			PlanID:        plan.ID,
			Section:       domain.KMJSectionHeader,
			CellKey:       hc.key,
			CellValue:     hc.value,
			StageColumn:   nil,
			IsAIGenerated: false,
			IsEdited:      false,
		})
	}

	// Stage cells (lesson_start, lesson_middle, lesson_end)
	startMin, middleMin, endMin := domain.KMJStageMinutes(plan.DurationMinutes)
	stages := []struct {
		section string
		time    string
	}{
		{domain.KMJSectionLessonStart, fmt.Sprintf("%d мин", startMin)},
		{domain.KMJSectionLessonMiddle, fmt.Sprintf("%d мин", middleMin)},
		{domain.KMJSectionLessonEnd, fmt.Sprintf("%d мин", endMin)},
	}

	columns := []string{
		domain.KMJColumnPedagogAction,
		domain.KMJColumnStudentAction,
		domain.KMJColumnAssessment,
		domain.KMJColumnResources,
	}

	for _, stage := range stages {
		// Stage duration cell (editable like any other cell)
		timeCol := domain.KMJColumnTime
		cells = append(cells, domain.KMJCell{
			ID:          uuid.New(),
			PlanID:      plan.ID,
			Section:     stage.section,
			CellKey:     "",
			CellValue:   stage.time,
			StageColumn: &timeCol,
		})
		for _, col := range columns {
			colStr := col
			cells = append(cells, domain.KMJCell{
				ID:            uuid.New(),
				PlanID:        plan.ID,
				Section:       stage.section,
				CellKey:       "",
				CellValue:     "",
				StageColumn:   &colStr,
				IsAIGenerated: false,
				IsEdited:      false,
			})
		}
	}

	// Footer cells
	footerCols := []string{
		domain.KMJColumnAssessmentTotal,
		domain.KMJColumnHomework,
	}

	for _, col := range footerCols {
		colStr := col
		cells = append(cells, domain.KMJCell{
			ID:            uuid.New(),
			PlanID:        plan.ID,
			Section:       domain.KMJSectionFooter,
			CellKey:       "",
			CellValue:     "",
			StageColumn:   &colStr,
			IsAIGenerated: false,
			IsEdited:      false,
		})
	}

	return cells
}

func (s *kmjService) ListTeachingMethods(ctx context.Context) ([]domain.TeachingMethod, error) {
	return s.methodRepo.List(ctx)
}

func (s *kmjService) GetNationalValueByMonth(ctx context.Context, month int) ([]domain.NationalValue, error) {
	return s.nationalValueRepo.GetByMonth(ctx, month)
}
