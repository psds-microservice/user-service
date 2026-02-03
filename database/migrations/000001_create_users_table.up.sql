-- Миграция 002: Создание таблицы users (из haqury/user-service db/migrations/002_create_users_table.sql)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 username VARCHAR(50) UNIQUE NOT NULL,
 email VARCHAR(255) UNIQUE NOT NULL,
 phone VARCHAR(20),
 password_hash VARCHAR(255) NOT NULL,
 status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),

 -- Настройки пользователя
 settings JSONB DEFAULT '{
 "default_quality": "hd",
 "max_parallel_streams": 3,
 "auto_start_recording": true,
 "notifications_enabled": true,
 "timezone": "UTC",
 "language": "en"
 }',

 -- Конфигурация стриминга
 streaming_config JSONB DEFAULT '{
 "server_url": "",
 "server_port": 8080,
 "use_ssl": false,
 "stream_endpoint": "/stream",
 "max_bitrate": 5000,
 "max_resolution": 1080
 }',

 -- Статистика
 stats JSONB DEFAULT '{
 "total_streams": 0,
 "total_duration": 0,
 "total_storage_used": 0,
 "successful_logins": 0,
 "failed_logins": 0
 }',

 -- Методанные
 metadata JSONB DEFAULT '{}',

 created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
 last_login TIMESTAMP WITH TIME ZONE,
 last_activity TIMESTAMP WITH TIME ZONE
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
 NEW.updated_at = CURRENT_TIMESTAMP;
 RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;
CREATE TRIGGER trigger_update_users_updated_at
 BEFORE UPDATE ON users
 FOR EACH ROW
 EXECUTE PROCEDURE update_users_updated_at();
