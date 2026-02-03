# user-service

Микросервис управления пользователями (GORM, PostgreSQL). Один бинарник: HTTP + gRPC, Cobra CLI. Адаптирован под [gss-template](https://github.com/psds-microservice/gss-template).

## API

- **HTTP** (порт по умолчанию **8080**): REST под префиксом `/api/v1/` — пользователи, аутентификация (JWT), операторы, сессии. Дополнительно: `/health`, `/ready`, `/swagger/` (OpenAPI UI и спека).
- **gRPC** (порт по умолчанию **9091**): сервис `UserService` — CreateUser, GetUser, UpdateUser, DeleteUser, Login, ValidateUserSession, UpdateUserPresence, GetAvailableOperators, UpdateOperatorStatus. Reflection включён.

Все операции доступны и по HTTP, и по gRPC. Спека OpenAPI генерируется из proto.

## Порты и конфиг

- `APP_PORT` / `HTTP_PORT` — HTTP (по умолчанию `8080`).
- `GRPC_PORT` / `METRICS_PORT` — gRPC (по умолчанию `9091`).
- Остальное: см. `.env.example`. В **production** обязательно задать `JWT_SECRET` (не дефолт) и `DB_PASSWORD`; при старте `api` конфиг валидируется.

## Setup

1. Скопировать `.env.example` в `.env`, задать `DB_*`, при необходимости порты и `JWT_SECRET`.
2. Запуск: `make run` или `go run ./cmd/user-service api`.
3. Все команды: `make help`.

## Cobra-команды

- `user-service api` — запуск HTTP + gRPC сервера (по умолчанию).
- `user-service migrate up` — выполнить миграции БД и выйти.
- `user-service seed` — миграции + сиды и выйти.

## Proto и OpenAPI

- **Proto**: `pkg/user_service/user_service.proto`. Сгенерированный Go: `pkg/gen/user_service/`.
- **Генерация Go из proto**: `make proto` (локальный `protoc` или Docker из `infra`). В целях proto добавлен `-I third_party` для `google/api/annotations.proto`.
- **OpenAPI из proto**: `make proto-openapi` (нужны `protoc` и `protoc-gen-openapiv2`). Результат: `api/openapi.json` / `api/openapi.swagger.json`. Swagger UI: `http://localhost:8080/swagger/index.html`.

## Deploy

- `deployments/Dockerfile`, `deployments/docker-compose.yml` — развёртывание (HTTP 8080, gRPC 9091, Postgres).

## Migrations

Версионированные SQL-миграции в `database/migrations/` (golang-migrate). При старте `api` и по `make migrate` выполняется `migrate up`. Сиды: `database/seeds/`, команда `user-service seed` или `make seed`.
