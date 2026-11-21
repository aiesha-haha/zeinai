# ZeinAI

AI-платформа для учителей: учебные материалы, тесты и задания с генерацией через AI, AI-студия (изображения/видео), интерактивные игры, кабинеты ученика и учителя. Интерфейс: **kk / ru / en**.

## Стек

| Слой | Технологии |
|---|---|
| Frontend | Next.js (App Router, `[locale]`), TypeScript |
| Backend | Go 1.23, Fiber v2 |
| БД | PostgreSQL 16 + pgvector |
| Кэш / очереди | Redis, RabbitMQ |
| AI | Google Gemini, OpenAI, Kimi K2 |
| Auth | JWT, Google OAuth, Firebase (FCM push) |
| Миграции | Goose (SQL) |
| Деплой | Docker / Docker Compose |

## Структура

```
├── frontend/            # Next.js — интерфейс (учитель, ученик, админ)
├── backend/             # Go API + воркеры AI (KMJ, генерация, тесты)
│   ├── cmd/             # точки входа: server, kmj_worker, cron
│   ├── internal/        # domain, handler, service, repository, worker
│   └── migrations/      # SQL-миграции (Goose)
├── docker-compose.yml         # прод-набор (с Cloudflare-туннелем)
├── docker-compose.local.yml   # локальный запуск, без туннеля
└── .env.example           # шаблон переменных окружения
```

## Быстрый старт

### 1. Переменные окружения

```bash
cp .env.example .env
# заполнить ключи AI / БД / JWT
```

### 2. Локальный запуск (без влияния на прод)

```bash
docker compose -f docker-compose.local.yml up -d --build
```

| Сервис | Адрес |
|---|---|
| Сайт | http://localhost:3005 |
| API (Swagger) | http://localhost:8000/swagger/index.html |
| RabbitMQ UI | http://localhost:15672 |

### 3. Прод-набор

```bash
docker compose up -d --build
```

## Безопасность

- `.env`, Firebase service-account и собранные бинарники **не хранятся в репозитории** (см. `.gitignore`).
- Пароли — bcrypt; авторизация — JWT; RBAC по ролям.

Подробности по API — в [backend/README.md](backend/README.md).
