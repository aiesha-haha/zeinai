# 🚀 КМЖ API - Инструкция по развертыванию

## ✅ Что реализовано:

### 1. **База данных**
- ✅ Миграция `010_kmj_generation.up.sql` создана
- ✅ Таблицы: `kmj_plans`, `kmj_cells`, `kmj_generation_tasks`
- ✅ Индексы для производительности

### 2. **Backend компоненты**
- ✅ Repository (pgxpool) - полностью рабочий
- ✅ Service с асинхронной генерацией
- ✅ Handler с REST API
- ✅ Worker для обработки очереди
- ✅ **Интеграция списания токенов (5 токенов за генерацию)**

### 3. **Функциональность**
- ✅ Создание черновика плана
- ✅ Генерация с AI (Gemini)
- ✅ Редактирование каждой ячейки отдельно
- ✅ Список планов с пагинацией
- ✅ Удаление планов (soft delete)
- ✅ **Автоматическое списание 5 токенов при генерации**
- ✅ Проверка баланса токенов перед генерацией

---

## 📋 Шаги для запуска:

### 1. Запуск базы данных

```powershell
# Из папки backend
cd d:\zeinai\backend
docker-compose up -d postgres rabbitmq
```

### 2. Применение миграций

```powershell
# Убедитесь что БД запущена, затем:
# (Если используете migrate CLI)
migrate -path ./migrations -database "postgresql://postgres:password@localhost:5432/zeinai?sslmode=disable" up

# ИЛИ вручную через psql:
docker exec -it zeinai-postgres psql -U postgres -d zeinai -f /migrations/010_kmj_generation.up.sql
```

### 3. Запуск сервера

```powershell
# Проверьте .env файл (должны быть настроены):
# - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
# - RABBITMQ_URL
# - GEMINI_API_KEY
# - JWT_SECRET

# Запуск сервера
go run ./cmd/server/main.go

# ИЛИ скомпилировать и запустить:
go build ./cmd/server -o server.exe
./server.exe
```

---

## 🧪 Тестирование API

### 1. Регистрация/Логин учителя

```bash
# Регистрация
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "iin": "123456789012",
    "full_name": "Арай Кенжебекова",
    "email": "arai@example.com",
    "phone": "+77012345678",
    "password": "SecurePass123!",
    "role": "teacher"
  }'

# Логин
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "iin": "123456789012",
    "password": "SecurePass123!"
  }'

# Сохраните access_token из ответа
```

### 2. Создание черновика ҚМЖ

```bash
curl -X POST http://localhost:8000/api/v1/kmj \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Информатика",
    "grade": 8,
    "lesson_number": 16,
    "lesson_topic": "Кестеде көрсетілген функция графигінің қасиеттерін өзгерту",
    "teacher_name": "Арай Кенжебекова",
    "content_lang": "kaz"
  }'

# Сохраните plan.id из ответа
```

### 3. Запуск генерации (стоимость: 5 токенов)

```bash
curl -X POST http://localhost:8000/api/v1/kmj/{PLAN_ID}/generate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# ⚠️ При этом запросе:
# 1. Проверяется баланс токенов
# 2. Списывается 5 токенов
# 3. Задача отправляется в RabbitMQ
# 4. Worker обрабатывает через Gemini AI
```

### 4. Проверка статуса

```bash
curl -X GET http://localhost:8000/api/v1/kmj/{PLAN_ID} \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Статусы:
# - "draft" - черновик
# - "generating" - генерируется
# - "completed" - готово (cells будут заполнены)
# - "failed" - ошибка
```

### 5. Редактирование ячейки

```bash
curl -X PUT http://localhost:8000/api/v1/kmj/cells/{CELL_ID} \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cell_value": "Обновленное содержимое ячейки"
  }'
```

### 6. Список планов

```bash
curl -X GET "http://localhost:8000/api/v1/kmj?page=1&page_size=50" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## 🔍 Логирование токенов

При запуске генерации в логах сервера будет:

```
[KMJ] 5 tokens consumed for user {UUID} (KMJ generation)
[KMJ] Generation task {TASK_UUID} queued for plan {PLAN_UUID} (cost: 5 tokens)
```

В БД в таблице `token_usage`:
```sql
SELECT * FROM token_usage WHERE action_type = 'kmj_generation' ORDER BY created_at DESC LIMIT 10;
```

---

## 📊 Структура данных

### План (kmj_plans)
- Метаданные урока
- Статус генерации
- Владелец (teacher/school)

### Ячейки (kmj_cells)
- Редактируемые независимо
- Помечены флагом `is_ai_generated`
- Флаг `is_edited` при ручном изменении

### Секции ячеек:
- **header** - Шапка (бөлім, педагог, тақырып, мақсаттар)
- **lesson_start** - Басы (10 мин)
- **lesson_middle** - Ортасы (30 мин)
- **lesson_end** - Соңы (5 мин)

### Колонки для этапов урока:
- `pedagog_action` - Педагогтың әрекеті
- `student_action` - Оқушының әрекеті
- `assessment` - Бағалау
- `resources` - Ресурстар

---

## 🔧 Troubleshooting

### Ошибка "insufficient tokens"
```json
{
  "error": {
    "code": "INSUFFICIENT_TOKENS",
    "message": "insufficient tokens: ..."
  }
}
```
**Решение:** Увеличьте лимит токенов пользователю через admin API или обновите подписку.

### База данных недоступна
```
Failed to connect to database
```
**Решение:**
```powershell
docker-compose up -d postgres
# Подождите 5-10 секунд для инициализации
```

### RabbitMQ недоступен
```
Failed to publish to queue
```
**Решение:**
```powershell
docker-compose up -d rabbitmq
# Проверьте: http://localhost:15672 (guest/guest)
```

### Worker не обрабатывает задачи
**Проверьте:**
1. Gemini API key в `.env`
2. Worker запущен (должен быть лог `[KMJ Worker] Started consuming from queue`)
3. Очередь `kmj.generation` существует в RabbitMQ

---

## 📚 Дополнительная документация

- **API Reference:** `docs/api/KMJ_API.md`
- **Swagger UI:** http://localhost:8000/swagger/index.html (после запуска)
- **Архитектура:** User Rules (Global Antigravity Rules) в проекте

---

## ✨ Следующие шаги (будущие улучшения)

- [ ] Vector search для похожих планов (семантический кэш)
- [ ] Экспорт в PDF/DOCX
- [ ] Библиотека шаблонов
- [ ] Совместное редактирование
- [ ] История версий

---

## 🎯 Соответствие требованиям

✅ **Каждая ячейка в отдельной записи БД** - можно редактировать  
✅ **Асинхронная генерация через RabbitMQ** - не блокирует API  
✅ **Списание 5 токенов** - интегрировано с subscription  
✅ **Multi-tenant** - поддержка школ и соло-учителей  
✅ **Production-ready** - pgxpool, retry logic, timeout, audit  
✅ **Swagger документация** - полное описание API  
✅ **Локализация** - content_lang: kaz/rus/eng  

**Статус: ✅ Готово к использованию!**
