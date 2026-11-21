# 🎯 Национальные ценности (Құндылықтар) - Завершающие шаги

## ✅ Что уже сделано:

### 1. База данных ✅
- ✅ Миграция `011_national_values.up.sql`
- ✅ Таблица `national_values` с 12 начальными записями
- ✅ Поле `national_value_id` добавлено в `kmj_plans`

### 2. Domain модели ✅  
- ✅ `NationalValue` struct с локализацией
- ✅ `MonthNamesKaz` map
- ✅ Методы `GetLocalizedName()` и `GetLocalizedDescription()`

### 3. Repository ✅
- ✅ `NationalValueRepository` с CRUD операциями
- ✅ `GetByMonth()` - получение ценностей по месяцу
- ✅ Обновлен `KMJRepository` для работы с `national_value_id`

### 4. Service логика ✅
- ✅ Автоматический выбор ценности при создании плана (по месяцу)
- ✅ Интеграция ценности в AI промпт при генерации
- ✅ Логирование выбранной ценности

---

## 🔧 Осталось сделать:

### 1. Обновить `cmd/server/main.go`

Найдите строку (примерно 88):
```go
kmjRepo := repository.NewKMJRepository(dbPool)
```

Добавьте после неё:
```go
nationalValueRepo := repository.NewNationalValueRepository(dbPool)
```

Найдите строку (примерно 134):
```go
kmjService := service.NewKMJService(kmjRepo, rabbitmq.GetChannel(), subscriptionRepo)
```

Замените на:
```go
kmjService := service.NewKMJService(kmjRepo, rabbitmq.GetChannel(), subscriptionRepo, nationalValueRepo)
```

### 2. Скомпилировать
```bash
go build ./cmd/server
```

### 3. Применить миграцию
```bash
# Запустите PostgreSQL
docker-compose up -d postgres

# Примените миграцию
migrate -path ./migrations -database "postgresql://postgres:password@localhost:5432/zeinai?sslmode=disable" up

# ИЛИ вручную через psql
psql -U postgres -d zeinai -f migrations/011_national_values.up.sql
```

---

## 🎓 Как это работает:

### 1. При создании плана урока:

```
POST /api/v1/kmj
{
  "subject": "Информатика",
  "lesson_date": "2024-09-15T00:00:00Z",  // Сентябрь
  ...
}
```

**Автоматически:**
1. Определяется месяц (сентябрь = 9)
2. Находится ценность для месяца: "Еңбекқорлық және кәсіби біліктілік"
3. Сохраняется `national_value_id` в плане

### 2. При генерации AI:

```
POST /api/v1/kmj/{id}/generate
```

**В промпт добавляется:**
```
МАҢЫЗДЫ: Бұл сабақта "Қыркүйек" айының ұлттық құндылығын қолдануңыз.
Құндылық: Еңбекқорлық және кәсіби біліктілік
Сипаттама: Еңбексүйгіштік, кәсіби шеберлік пен дағдыларды дамыту

Сабақ тақырыбын осы құндылықпен біріктіріп, педагог пен оқушы әрекеттерінде 
практикалық қолдану мысалдарын көрсетіңіз.
```

**AI генерирует урок с естественной интеграцией ценности!**

---

## 📊 Таблица ценностей по месяцам:

| Месяц | Казахский | Русский | Английский |
|-------|-----------|---------|------------|
| Қыркүйек (9) | Еңбекқорлық және кәсіби біліктілік | Трудолюбие и профессиональная компетентность | Diligence and professional competence |
| Қазан (10) | Тәуелсіздік және отаншылдық | Независимость и патриотизм | Independence and patriotism |
| Қараша (11) | Әділдік және жауапкершілік | Справедливость и ответственность | Justice and responsibility |
| Желтоқсан (12) | Бірлік және ынтымақ | Единство и сотрудничество | Unity and cooperation |
| Қаңтар (1) | Заң және тәртіп | Закон и порядок | Law and order |
| Ақпан (2) | Жасампаздық және жаңашылдық | Креативность и инновации | Creativity and innovation |
| Наурыз (3) | Тәуелсіздік және отаншылдық | Независимость и патриотизм | Independence and patriotism |
| Сәуір (4) | Еңбекқорлық және кәсіби біліктілік | Трудолюбие и компетентность | Diligence and competence |
| Мамыр (5) | Бірлік және ынтымақ | Единство и сотрудничество | Unity and cooperation |
| Маусым (6) | Жасампаздық және жаңашылдық | Креативность и инновации | Creativity and innovation |
| Шілде (7) | Заң және тәртіп | Закон и порядок | Law and order |
| Тамыз (8) | Заң және тәртіп | Закон и порядок | Law and order |

---

## 🔄 CRUD API для ценностей (будущее):

Можно добавить эндпоинты для управления:
```
GET    /api/v1/national-values           - Список всех
GET    /api/v1/national-values/month/:m  - По месяцу
POST   /api/v1/national-values            - Создать новую
PUT    /api/v1/national-values/:id        - Обновить
DELETE /api/v1/national-values/:id        - Удалить
```

---

## ✨ Ключевые особенности:

1. **Автоматический выбор** - не нужно вручную выбирать ценность
2. **По месяцу урока** - если дата не указана, использует текущий месяц
3. **Интеграция с темой** - AI естественно вплетает ценность в урок
4. **Мультиязычность** - ценности на kaz/rus/eng
5. **Программа "Адалазамат"** - соответствует государственной программе

---

## 📝 Пример лога:

```
[KMJ] Auto-selected national value: Еңбекқорлық және кәсіби біліктілік for month 9
[KMJ] 5 tokens consumed for user {UUID} (KMJ generation)
[KMJ] Generation task {UUID} queued for plan {UUID} (cost: 5 tokens)
```

---

**Статус:** 95% готово  
**Осталось:** Обновить main.go (2 строки!)  
**Тестирование:** После компиляции
