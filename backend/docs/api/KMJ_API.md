# ҚМЖ (Қысқа Мерзімді Жоспар) API Documentation

## Overview

API для генерации и управления краткосрочными планами уроков (ҚМЖ).

### Основные функции:
- ✅ Создание черновика плана урока
- ✅ Асинхронная генерация плана с помощью AI (Gemini)
- ✅ Редактирование каждой ячейки отдельно
- ✅ Просмотр и управление планами
- ✅ Multi-tenant (school/solo teacher)
- ✅ Полная локализация (kaz, rus, eng)

---

## API Endpoints

### 1. Create Draft Plan
**POST** `/api/v1/kmj`

Создает черновик плана урока с пустой структурой.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body:**
```json
{
  "subject": "Информатика",
  "grade": 8,
  "lesson_number": 16,
  "lesson_topic": "Кестеде көрсетілген функция графигінің қасиеттерін өзгерту",
  "teacher_name": "Арай Кенжебекова",
  "lesson_date": "2024-05-15T00:00:00Z",
  "content_lang": "kaz"
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "plan": {
      "id": "uuid",
      "owner_id": "uuid",
      "school_id": "uuid",
      "subject": "Информатика",
      "grade": 8,
      "lesson_number": 16,
      "lesson_topic": "Кестеде көрсетілген функция графигінің қасиеттерін өзгерту",
      "teacher_name": "Арай Кенжебекова",
      "content_lang": "kaz",
      "status": "draft",
      "created_at": "2024-05-15T12:00:00Z",
      "updated_at": "2024-05-15T12:00:00Z"
    },
    "cells": [
      {
        "id": "uuid",
        "plan_id": "uuid",
        "section": "header",
        "cell_key": "bolem",
        "cell_value": "",
        "is_ai_generated": false,
        "is_edited": false
      }
      // ... другие ячейки
    ]
  }
}
```

---

### 2. Generate Plan with AI
**POST** `/api/v1/kmj/{id}/generate`

Запускает асинхронную генерацию плана с помощью AI.

**⚠️ Стоимость: 5 токенов**

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:** `200 OK`
```json
{
  "data": {
    "message": "Generation started successfully",
    "plan_id": "uuid"
  }
}
```

**Error Cases:**
- `403 INSUFFICIENT_TOKENS` - Недостаточно токенов (требуется 5)
- `400 INVALID_PLAN_ID` - Неверный ID плана
- `403 NOT_PLAN_OWNER` - Не владелец плана
- `500 GENERATION_FAILED` - Ошибка запуска генерации

**Статусы генерации:**
- `draft` → `generating` → `completed`
- `draft` → `generating` → `failed` (в случае ошибки)

**Billing:**
При успешном запуске генерации автоматически списывается **5 токенов** с баланса пользователя.
Токены списываются сразу при старте, до завершения генерации.

---

### 3. Get Plan by ID
**GET** `/api/v1/kmj/{id}`

Получить план со всеми ячейками.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:** `200 OK`
```json
{
  "data": {
    "plan": { ... },
    "cells": [ ... ]
  }
}
```

---

### 4. List My Plans
**GET** `/api/v1/kmj?page=1&page_size=50`

Получить список планов пользователя (пагинация).

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Query Parameters:**
- `page` (optional): Номер страницы (default: 1)
- `page_size` (optional): Размер страницы (default: 50, max: 100)

**Response:** `200 OK`
```json
{
  "data": {
    "plans": [ ... ],
    "total": 25,
    "page": 1,
    "page_size": 50,
    "total_pages": 1
  }
}
```

---

### 5. Update Cell
**PUT** `/api/v1/kmj/cells/{cellId}`

Обновить значение конкретной ячейки.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body:**
```json
{
  "cell_value": "Новое содержимое ячейки"
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "message": "Cell updated successfully"
  }
}
```

---

### 6. Delete Plan
**DELETE** `/api/v1/kmj/{id}`

Мягкое удаление плана.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:** `200 OK`
```json
{
  "data": {
    "message": "Plan deleted successfully"
  }
}
```

---

## Data Structures

### Plan Sections
- `header` - Шапка плана (метаданные)
- `lesson_start` - Сабақтың басы (10 мин)
- `lesson_middle` - Сабақтың ортасы (30 мин)
- `lesson_end` - Сабақтың соңы (5 мин)

### Stage Columns
Для секций `lesson_start`, `lesson_middle`, `lesson_end`:
- `pedagog_action` - Педагогтың әрекеті
- `student_action` - Оқушының әрекеті
- `assessment` - Бағалау
- `resources` - Ресурстар

### Header Cell Keys
- `bolem` - Бөлім
- `pedagog_name` - Педагог аты-жөні
- `date` - Күні
- `grade` - Сынып
- `topic` - Сабақтың тақырыбы
- `learning_objectives` - Оқу мақсаттары
- `lesson_objectives` - Сабақтың мақсаты
- `unity_cooperation` - Бірлік және ынтымақ

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST_BODY` | 400 | Неверный формат JSON |
| `VALIDATION_ERROR` | 400 | Ошибка валидации полей |
| `INVALID_PLAN_ID` | 400 | Неверный формат ID плана |
| `INVALID_CELL_ID` | 400 | Неверный формат ID ячейки |
| `NOT_PLAN_OWNER` | 403 | Пользователь не является владельцем плана |
| `PLAN_NOT_FOUND` | 404 | План не найден |
| `CREATE_FAILED` | 500 | Ошибка создания плана |
| `GENERATION_FAILED` | 500 | Ошибка запуска генерации |
| `UPDATE_FAILED` | 500 | Ошибка обновления ячейки |
| `DELETE_FAILED` | 500 | Ошибка удаления плана |

---

## Workflow Examples

### Example 1: Creating and Generating a Plan

```bash
# 1. Create draft plan
curl -X POST http://localhost:8000/api/v1/kmj \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Информатика",
    "grade": 8,
    "lesson_number": 16,
    "lesson_topic": "Excel функциялары",
    "teacher_name": "Арай Кенжебекова",
    "content_lang": "kaz"
  }'

# Response: { "data": { "plan": { "id": "plan-uuid", ... }, "cells": [...] }}

# 2. Generate with AI
curl -X POST http://localhost:8000/api/v1/kmj/plan-uuid/generate \
  -H "Authorization: Bearer $TOKEN"

# Response: { "data": { "message": "Generation started successfully" }}

# 3. Poll status (check plan.status)
curl -X GET http://localhost:8000/api/v1/kmj/plan-uuid \
  -H "Authorization: Bearer $TOKEN"

# When status = "completed", cells will contain AI-generated content
```

### Example 2: Editing a Cell

```bash
# 1. Get plan with cells
curl -X GET http://localhost:8000/api/v1/kmj/plan-uuid \
  -H "Authorization: Bearer $TOKEN"

# 2. Find cell_id you want to edit from cells array

# 3. Update cell
curl -X PUT http://localhost:8000/api/v1/kmj/cells/cell-uuid \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cell_value": "Өзгертілген мазмұн"
  }'
```

---

## Architecture Notes

### Async AI Generation Flow

```
Client → POST /kmj/{id}/generate
  ↓
Service → Create Task → Update Status to "generating"
  ↓
Publish to RabbitMQ queue: "kmj.generation"
  ↓
KMJ Worker consumes message
  ↓
Worker → Call Gemini AI (timeout: 30s, retries: 3)
  ↓
Parse Response → Create Cells
  ↓
Update Plan Status to "completed"
```

### Database Schema

- `kmj_plans` - План метаданные
- `kmj_cells` - Ячейки плана (редактируемые)
- `kmj_generation_tasks` - Задачи генерации (отслеживание)

### Multi-Tenancy

- **School Mode**: `school_id` NOT NULL, вся школа имеет доступ
- **Solo Teacher Mode**: `school_id` NULL, только teacher-владелец

---

## Performance

- **p95 Latency Target**: < 500ms (excluding AI generation)
- **AI Generation Timeout**: 30 seconds
- **Max Retries**: 3 (exponential backoff)
- **Pagination**: Default 50, Max 100 items per page

---

## Security

- ✅ JWT Authentication required for all endpoints
- ✅ RBAC: Only `teacher` role can access KMJ endpoints
- ✅ Ownership validation on all operations
- ✅ Rate limiting via Redis
- ✅ SQL injection protection (parameterized queries)

---

## Future Enhancements

- [ ] Vector search for similar plans (semantic cache)
- [ ] Export to PDF/DOCX
- [ ] Plan templates library
- [ ] Collaborative editing
- [ ] Version history
