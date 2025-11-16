# PR Reviewer Assignment Service

## Основной функционал
| Метод | Путь | Описание |
| --- | --- | --- |
| `POST` | `/team/add` | Создание команды и массовое добавление/обновление пользователей (id, username, isActive). |
| `GET` | `/team/get?team_name=...` | Получение состава конкретной команды. |
| `GET` | `/team/stats?team_name=...` | Собственная агрегация по команде: общее/активное число участников, количество PR в статусах, среднее время до merge. |
| `POST` | `/pullRequest/create` | Создание PR и автоматическое назначение до двух активных ревьюверов из команды автора (автор исключён). |
| `POST` | `/pullRequest/merge` | Идемпотентная фиксация статуса `MERGED`, после которой назначение запрещено. |
| `POST` | `/pullRequest/reassign` | Переназначение конкретного ревьювера на случайного активного участника из его команды (исключая автора и дубликаты). |
| `POST` | `/users/setIsActive` | Переключение активности пользователя; неактивные не попадают в новые назначения. |
| `GET` | `/users/getReview?user_id=...` | Список PR, где пользователь назначен ревьювером. |
| `GET` | `/health` | Health-check контейнера. |

Автогенерируемая документация доступна на `http://localhost:8080/swagger/index.html` после старта сервиса.

## Используемые технологии
- Go 1.25.
- HTTP роутер `github.com/go-chi/chi/v5`, валидация `go-playground/validator`.
- PostgreSQL 16 + драйвер `pgx/v5`, собственный слой репозиториев и TxManager.
- Миграции на чистом SQL, отдельный мигратор `src/cmd/migrator`.
- Docker, docker-compose.
- Swagger (`swag`) для OpenAPI-спецификации.
- Юнит-тесты, интеграционные тесты на `testcontainers-go`.
- Makefile с целями запуска/тестов.

## Архитектура
- `internal/domain` — сущности, доменные ошибки, интерфейсы репозиториев/сервисов.
- `internal/application` — бизнес-сервисы: назначение ревьюверов и merge, управление командами и пользователями.
- `internal/infrastructure/data` — инфраструктура PostgreSQL: pgx pool, менеджер транзакций, реализации репозиториев, SQL-модели, миграции и интеграционные тесты.
- `internal/http_api` — контроллеры (REST), DTO, middleware логирования, swagger модели.
- `cmd/http_api` — сборка зависимостей, конфигурация, запуск HTTP-сервера и graceful shutdown.
- `cmd/migrator` — утилита миграций, которая читается entrypoint'ом контейнера перед стартом API.

## Структура репозитория
```
.
├─ src
│  ├─ cmd
│  │  ├─ http_api          # инициализация зависимостей и запуск REST API
│  │  └─ migrator          # простая утилита выполнения *.up.sql миграций
│  └─ internal
│     ├─ domain            # модели, контракты, ошибки
│     ├─ application       # бизнес-сервисы, контракты TxManager
│     ├─ http_api          # контроллеры, middleware, модели ответа, swagger
│     └─ infrastructure
│        └─ data           # pgx pool, репозитории, миграции, интеграционные тесты
├─ Dockerfile              # multi-stage сборка API и мигратора
├─ docker-compose.yml      # api + postgres 16 с хранением данных в volume
├─ docker-entrypoint.sh    # ожидание БД, миграции, запуск API
├─ Makefile                # цели для запуска, миграций, тестов и swagger
├─ go.mod / go.sum
└─ Task.md                 # формулировка тестового задания
```

## Переменные окружения
Все параметры читаются через `cmd/config` и могут храниться в `.env` (используется `godotenv`). Ниже перечислены основные переменные и значения по умолчанию:

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `HTTP_PORT` | `8080` | Порт HTTP-сервера. |
| `LOG_LEVEL` | `INFO` | Уровень логирования (`DEBUG/INFO/WARN/ERROR`). |
| `LOG_FORMAT` | `text` | Формат логов (`text` или `json`). |
| `DB_HOST` | `localhost` | Хост PostgreSQL. В docker-compose переопределяется на `postgres`. |
| `DB_PORT` | `5432` | Порт PostgreSQL. |
| `DB_USER` | `user` | Имя пользователя БД. |
| `DB_PASSWORD` | `password` | Пароль пользователя БД. |
| `DB_NAME` | `db` | Имя базы данных. |
| `DB_SSLMODE` | `disable` | Режим SSL соединения. |
| `DB_MAX_CONNS` | `10` | Максимальное число соединений пула pgx. |
| `DB_MIN_CONNS` | `2` | Минимальное число соединений пула. |
| `MAX_CONN_LIFETIME` | `1800` (сек) | TTL соединения. |
| `MAX_CONN_IDLE_TIME` | `300` (сек) | Интервал простоя соединения. |
| `HEALTH_CHECK_PERIOD` | `60` (сек) | Частота health-check'ов пула. |
| `MIGRATIONS_DIR` | `src/internal/infrastructure/data/migrations` | Путь к SQL миграциям для `migrator`. |

## Запуск
### Быстрый старт (docker-compose)
1. Соберите и поднимите сервис: `docker-compose up --build` или `make up`.
2. Сервис станет доступен на `http://localhost:${HTTP_PORT:-8080}`. EntryPoint дождётся Postgres, выполнит `/app/migrator` и запустит API.
3. Swagger UI: `http://localhost:${HTTP_PORT:-8080}/swagger/index.html`.

### Локальный запуск без контейнеров
1. Поднимите PostgreSQL (например, через `docker compose up postgres -d`) и примените миграции: `make migrate`.
2. Экспортируйте переменные окружения подключения к БД или заполните `.env`.
3. Запустите API командой `make api` или `go run ./src/cmd/http_api`.

### Миграции
- Автоматически выполняются в контейнере перед стартом API (`docker-entrypoint.sh`).
- Для ручного запуска используйте `make migrate` (читает те же переменные окружения).

## Тестирование и разработка
- `make test` — юнит-тесты (сервисы, контроллеры).
- `make test-integration` — интеграционные тесты для репозиториев (требуют Docker, запускают временный PostgreSQL через `testcontainers`).
- `make test-all` — оба набора.
- `make swagger` — пересоздание swagger-спецификации после изменения контроллеров/моделей.
- `make lint` — запуск `golangci-lint` (требуется установленный golangci-lint).

## Допущения и решения
- Выбор пользователей для назначения и переназначения ревьюверов происходит случайным образом.
- Добавлены дополнительные статус коды для обработки ошибок
- Добавлен эндпоинт для получения статистики команды `/team/stats?team_name=...`
- Реализован graceful shutdown