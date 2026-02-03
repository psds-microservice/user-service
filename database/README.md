# Database

Структура перенесена из [haqury/user-service/db](https://github.com/haqury/user-service/tree/main/db).

## Структура

```
database/
├── migrations/   # Миграции схемы (CREATE TABLE, ALTER TABLE, индексы, триггеры)
└── seeds/        # Начальные данные (INSERT)
```

## Миграции

- **Нумерация:** `000001_name.up.sql` / `000001_name.down.sql` (golang-migrate).
- **Запуск:** `make migrate` или при старте приложения.
- **Отслеживание:** таблица `schema_migrations`.

Текущие миграции (перенесены из haqury/user-service db/migrations/):
- `000001_create_users_table` — расширение pgcrypto, таблица `users` (UUID, username, email, phone, password_hash, status, settings, streaming_config, stats, metadata, created_at, updated_at, last_login, last_activity), индексы, триггер updated_at.
- `000002_create_user_services_table` — таблица `user_services` (id UUID, user_id, service_name, service_type, base_url, api_endpoint, port, use_ssl, routing_key, queue_name, topic_name, max_bitrate, max_connections, priority, is_active, enabled_buttons, schedule, parameters), индексы, триггер updated_at, UNIQUE(user_id, service_name).

## Seeds

- **Файлы:** `001_seed_data.sql`, … (по порядку имени).
- **Запуск:** `make seed` (сначала применяются миграции, затем все `*.sql` из `database/seeds/`).
- **Содержимое:** тестовый пользователь `testuser` / `test@example.com` и три записи в `user_services`.

## Команды

```bash
make migrate    # Только миграции
make seed       # Миграции + seeds
make db-init    # То же: migrate up + seed
```

**Важно:** текущая Go-модель `internal/model/entity.go` (User с id, email, name, password, notes) не совпадает со схемой из репозитория (users с UUID, username, email, password_hash, settings, streaming_config и т.д.). Для полного соответствия нужно обновить модель и код под новую схему БД.
