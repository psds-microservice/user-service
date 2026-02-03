# user-service

Микросервис для управления пользователями (GORM, PostgreSQL). Адаптирован под [gss-template](https://github.com/psds-microservice/gss-template): один бинарник, HTTP + gRPC.

## API

- **HTTP** (порт 8080): REST для пользователей, `/health`, `/ready`.
- **gRPC** (порт 9090): `UserService` (CreateUser, GetUser, UpdateUser, DeleteUser, Login). Proto: `pkg/proto/user_service/user_service.proto`.

## Setup

1. Скопировать `.env.example` в `.env`, задать DB_*, при необходимости HTTP_PORT/GRPC_PORT.
2. `make run` или `go run ./cmd/user-service`.
3. Список команд: `make help`.

## Proto

- Исходники: `pkg/proto/user_service/`. Сгенерированный Go-код: `pkg/gen/user_service/`.
- `make proto` — через Docker-образ из `infra/protoc-go.Dockerfile` генерирует код в `pkg/gen/`.
- В репозитории уже есть сгенерированные файлы; при изменении `.proto` перегенерируйте через `make proto`.

## Deploy

- `deployments/Dockerfile` и `deployments/docker-compose.yml` — развёртывание в стиле gss-template (HTTP 8080, gRPC 9090, Postgres).

## Migrations

Версионированные SQL-миграции в `database/migrations/` (golang-migrate). При старте приложения и по `make migrate` выполняется `migrate up`. Начальная миграция `000001_init_users` создаёт таблицу `users` по модели GORM.
