# 📚 Методы обучения (Әдістер) - Реализация

## ✅ Что создано:

### 1. База данных ✅
- ✅ Таблица `teaching_methods` - 50 методов обучения
- ✅ Таблица `kmj_plan_methods` - связь many-to-many (макс 4)
- ✅ 10 начальных методов добавлено для примера
- ✅ Constraint: максимум 4 метода на план

### 2. Миграция ✅
- ✅ `012_teaching_methods.up.sql`
- ✅ `012_teaching_methods.down.sql`

---

## 🎯 Что нужно доделать:

### 1. Domain модель
Создать `internal/domain/teaching_method.go`:
```go
package domain

import (
	"database/sql"
	"time"
	"github.com/google/uuid"
)

type TeachingMethod struct {
	ID             uuid.UUID    `json:"id" db:"id"`
	NameKaz        string       `json:"name_kaz" db:"name_kaz"`
	NameRus        string       `json:"name_rus" db:"name_rus"`
	NameEng        string       `json:"name_eng" db:"name_eng"`
	TypeKaz        *string      `json:"type_kaz,omitempty" db:"type_kaz"`
	PurposeKaz     *string      `json:"purpose_kaz,omitempty" db:"purpose_kaz"`
	PurposeRus     *string      `json:"purpose_rus,omitempty" db:"purpose_rus"`
	PurposeEng     *string      `json:"purpose_eng,omitempty" db:"purpose_eng"`
	Modules        *string      `json:"modules,omitempty" db:"modules"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt      sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}

type KMJPlanMethod struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PlanID    uuid.UUID `json:"plan_id" db:"plan_id"`
	MethodID  uuid.UUID `json:"method_id" db:"method_id"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

### 2. Repository
Создать `internal/repository/teaching_method_repository.go`:
```go
package repository

import (
	"context"
	"zeinai-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeachingMethodRepository interface {
	// CRUD
	Create(ctx context.Context, method *domain.TeachingMethod) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TeachingMethod, error)
	List(ctx context.Context) ([]domain.TeachingMethod, error)
	Update(ctx context.Context, id uuid.UUID, method *domain.TeachingMethod) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Plan methods
	AddMethodToPlan(ctx context.Context, planID, methodID uuid.UUID, sortOrder int) error
	GetPlanMethods(ctx context.Context, planID uuid.UUID) ([]domain.TeachingMethod, error)
	RemoveMethodFromPlan(ctx context.Context, planID, methodID uuid.UUID) error
	SetPlanMethods(ctx context.Context, planID uuid.UUID, methodIDs []uuid.UUID) error
}
```

### 3. Обновить KMJ Service
В `internal/service/kmj_service.go`:

**Добавить в строку 34:**
```go
type kmjService struct {
	kmjRepo          repository.KMJRepository
	channel          *amqp.Channel
	subscriptionRepo domain.SubscriptionRepository
	nationalValueRepo repository.NationalValueRepository
	methodRepo       repository.TeachingMethodRepository // ← ДОБАВИТЬ
}
```

**Обновить buildPrompt (примерно строка 343):**
```go
func (s *kmjService) buildPrompt(plan *domain.KMJPlan) string {
	// ... существующий код с national value ...
	
	// Add teaching methods
	var methodsInstruction string
	methods, err := s.methodRepo.GetPlanMethods(context.Background(), plan.ID)
	if err == nil && len(methods) > 0 {
		methodNames := make([]string, len(methods))
		for i, m := range methods {
			methodNames[i] = m.GetLocalizedName(plan.ContentLang)
		}
		
		methodsInstruction = fmt.Sprintf(`
ӘДІС-ТӘСІЛДЕР: Сабақты өткізу үшін келесі әдістерді қолданыңыз:
%s

Әр әдісті сабақтың әртүрлі кезеңдерінде қолданып, әрбір әдістің мақсатына 
сәйкес қолдануды көрсетіңіз. Әдістерді тақырыппен табиғи түрде біріктіріңіз.
`, strings.Join(methodNames, ", "))
	}
	
	// ... қалған код ...
}
```

### 4. Обновить CreateKMJRequest DTO
```go
type CreateKMJRequest struct {
	Subject      string       `json:"subject" validate:"required,max=255"`
	Grade        int          `json:"grade" validate:"required,min=1,max=12"`
	LessonNumber int          `json:"lesson_number" validate:"required,min=1"`
	LessonTopic  string       `json:"lesson_topic" validate:"required,max=500"`
	TeacherName  string       `json:"teacher_name" validate:"required,max=255"`
	LessonDate   *time.Time   `json:"lesson_date,omitempty"`
	ContentLang  string       `json:"content_lang" validate:"required,oneof=kaz rus eng"`
	MethodIDs    []uuid.UUID  `json:"method_ids,omitempty" validate:"omitempty,max=4"` // ← MAX 4
}
```

### 5. Обновить CreateDraftPlan
После создания плана, добавить методы:
```go
func (s *kmjService) CreateDraftPlan(...) {
	// ... существующий код создания плана ...
	
	// Add selected methods (max 4)
	if len(req.MethodIDs) > 0 {
		err = s.methodRepo.SetPlanMethods(ctx, plan.ID, req.MethodIDs)
		if err != nil {
			log.Printf("[KMJ] Failed to add methods: %v", err)
			// Не критично - продолжаем
		}
	}
	
	// ... остальной код ...
}
```

---

## 🌐 API Endpoints (создать handler):

```
GET    /api/v1/teaching-methods        - Список всех методов
GET    /api/v1/teaching-methods/:id    - Получить метод
POST   /api/v1/teaching-methods         - Создать (admin)
PUT    /api/v1/teaching-methods/:id     - Обновить (admin)
DELETE /api/v1/teaching-methods/:id     - Удалить (admin)
```

---

## 📋 Пример запроса:

```json
POST /api/v1/kmj
{
  "subject": "Информатика",
  "grade": 8,
  "lesson_number": 16,
  "lesson_topic": "Excel функциялары",
  "teacher_name": "Арай",
  "lesson_date": "2024-09-15T00:00:00Z",
  "content_lang": "kaz",
  "method_ids": [
    "uuid-brainstorming",
    "uuid-jigsaw",
    "uuid-fishbone",
    "uuid-cluster"
  ]
}
```

**AI Prompt будет включать:**
1. ✅ Национальную ценность месяца
2. ✅ 4 выбранных метода обучения
3. ✅ Тему урока

**Результат:** AI создаст урок где естественно применяются все 4 метода в разных этапах!

---

## 📊 Полная структура 50 методов:

Добавить все 50 методов из файла `әдістер.md` через INSERT statements или админ панель.

Методы включают:
- Миға шабуыл (Brainstorming)
- Джигсо (Jigsaw)
- Фишбоун (Fishbone)
- Кластер (Cluster)
- Галереяны шарлау (Gallery Walk)
- INSERT
- Ыстық орындық (Hot Seat)
- ... и 43 других

---

## 🎓 Приоритет:

1. ✅ Миграция создана
2. ⏳ Domain модель (5 мин)
3. ⏳ Repository (15 мин)
4. ⏳ Обновить Service (10 мин)
5. ⏳ Handler для CRUD (20 мин)
6. ⏳ Протестировать

**Статус:** 30% готово  
**Осталось:** Domain, Repository, Service integration
