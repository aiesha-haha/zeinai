# Zeinai EdTech Backend

## 🔐 Безопасность

**ВАЖНО:** Перед первым коммитом:

1. Скопируйте `.env.example` в `.env`:
   ```bash
   cp .env.example .env
   ```

2. Заполните реальные значения в `.env`:
   - `DB_PASSWORD` - пароль для PostgreSQL
   - `RABBITMQ_PASSWORD` - пароль для RabbitMQ
   - `JWT_SECRET` - секретный ключ для JWT (минимум 32 символа)
   - `GEMINI_API_KEY` - ваш API ключ от Google Gemini

3. **НИКОГДА** не коммитьте файл `.env` в Git!

## 🚀 Быстрый старт

### Требования
- Docker и Docker Compose
- (Опционально) Go 1.23+ для локальной разработки

### Запуск проекта

1. Клонируйте репозиторий:
   ```bash
   git clone <repository-url>
   cd backend
   ```

2. Настройте переменные окружения:
   ```bash
   cp .env.example .env
   # Отредактируйте .env и добавьте реальные значения
   ```

3. Запустите все сервисы:
   ```bash
   docker-compose up -d --build
   ```

4. Проверьте статус:
   ```bash
   docker-compose ps
   docker logs zeinai_backend
   ```

## 📚 API Документация

После запуска, Swagger UI доступен по адресу:
```
http://localhost:8000/swagger/index.html
```

### Основные endpoints:

**Авторизация:**
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход в систему

**Школы:**
- `GET /api/v1/schools` - Список школ (для формы регистрации)

**AI (требует авторизации):**
- `POST /api/v1/ai/generate` - Генерация текста через Gemini
- `GET /api/v1/ai/status/:id` - Статус AI задачи

**Система:**
- `GET /api/v1/health` - Health check

## 🏗 Архитектура

```
backend/
├── cmd/
│   └── server/          # Точка входа приложения
├── internal/
│   ├── config/          # Конфигурация
│   ├── domain/          # Модели данных и интерфейсы
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # Middleware (auth, RBAC)
│   ├── repository/      # Работа с БД
│   ├── service/         # Бизнес-логика
│   ├── worker/          # Фоновые воркеры (AI)
│   └── pkg/
│       ├── gemini/      # Клиент Gemini API
│       ├── i18n/        # Локализация
│       ├── response/    # Стандартные ответы API
│       └── utils/       # Утилиты (JWT и др.)
├── platform/
│   ├── database/        # PostgreSQL подключение
│   └── rabbitmq/        # RabbitMQ клиент
├── migrations/          # SQL миграции (Goose)
├── locales/            # Файлы переводов (en, ru, kk)
├── docs/               # Swagger документация (генерируется)
├── docker-compose.yml
├── Dockerfile
└── .env.example        # Шаблон переменных окружения
```

## 🛠 Разработка

### Локальная разработка (без Docker)

1. Установите зависимости:
   ```bash
   go mod download
   ```

2. Запустите инфраструктуру (БД, Redis, RabbitMQ):
   ```bash
   docker-compose up -d db redis rabbitmq
   ```

3. Примените миграции:
   ```bash
   goose -dir migrations postgres "user=zeinai_user password=zeinai_pass dbname=zeinai_db sslmode=disable" up
   ```

4. Запустите сервер:
   ```bash
   go run cmd/server/main.go
   ```

### Генерация Swagger документации

```bash
swag init -g cmd/server/main.go --output docs
```

### Создание новой миграции

```bash
goose -dir migrations create <migration_name> sql
```

## 🧪 Тестирование

### Примеры запросов

**Регистрация:**
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+77771112233",
    "iin": "900101300456",
    "password": "SecurePass123!",
    "first_name": "Айдар",
    "last_name": "Беков",
    "birth_year": 1990,
    "gender": "male",
    "role": "teacher"
  }'
```

**Логин:**
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+77771112233",
    "password": "SecurePass123!"
  }'
```

## 🌍 Локализация

Поддерживаемые языки:
- `kk` - Казахский
- `ru` - Русский
- `en` - Английский

Язык определяется по заголовку `Accept-Language`.

## 📦 Технологический стек

- **Backend:** Go 1.23, Fiber v2
- **Database:** PostgreSQL 16 + pgvector
- **Cache:** Redis
- **Queue:** RabbitMQ
- **AI:** Google Gemini API
- **Auth:** JWT
- **Docs:** Swagger/OpenAPI
- **Migrations:** Goose
- **Containerization:** Docker, Docker Compose

## 🔒 Безопасность

- Пароли хешируются с помощью bcrypt
- JWT токены для авторизации
- RBAC (Role-Based Access Control)
- Валидация всех входных данных
- Parameterized SQL queries (защита от SQL injection)
- Rate limiting (планируется)

## 📝 Лицензия

[Укажите лицензию]

## 👥 Команда

[Информация о команде]
