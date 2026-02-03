-- Миграция 003: Создание таблицы user_services (из haqury/user-service db/migrations/003_create_user_services_table.sql)

CREATE TABLE IF NOT EXISTS user_services (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

 service_name VARCHAR(100) NOT NULL,
 service_type VARCHAR(50) NOT NULL CHECK (service_type IN ('streaming', 'recording', 'monitoring', 'analytics', 'custom')),

 -- Конечные точки для AGW
 base_url VARCHAR(500) NOT NULL,
 api_endpoint VARCHAR(200) DEFAULT '/api/v1',
 port INTEGER NOT NULL,
 use_ssl BOOLEAN DEFAULT FALSE,
 ssl_certificate_path VARCHAR(500),

 -- Параметры перенаправления
 routing_key VARCHAR(100),
 queue_name VARCHAR(100),
 topic_name VARCHAR(100),

 -- Конфигурация потоков
 max_bitrate INTEGER DEFAULT 5000,
 max_connections INTEGER DEFAULT 10,
 priority INTEGER DEFAULT 1 CHECK (priority BETWEEN 1 AND 10),

 -- Активность
 is_active BOOLEAN DEFAULT TRUE,
 enabled_buttons JSONB DEFAULT '["button1", "button2", "button3"]',

 -- Время работы
 schedule JSONB DEFAULT '{
 "always": true,
 "weekdays": ["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"],
 "start_time": "00:00",
 "end_time": "23:59"
 }',

 -- Дополнительные параметры
 parameters JSONB DEFAULT '{}',

 created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

 UNIQUE(user_id, service_name)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_user_services_user_id ON user_services(user_id);
CREATE INDEX IF NOT EXISTS idx_user_services_service_type ON user_services(service_type);
CREATE INDEX IF NOT EXISTS idx_user_services_is_active ON user_services(is_active);
CREATE INDEX IF NOT EXISTS idx_user_services_priority ON user_services(priority);

-- Триггер
CREATE OR REPLACE FUNCTION update_user_services_updated_at()
RETURNS TRIGGER AS $$
BEGIN
 NEW.updated_at = CURRENT_TIMESTAMP;
 RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS trigger_update_user_services_updated_at ON user_services;
CREATE TRIGGER trigger_update_user_services_updated_at
 BEFORE UPDATE ON user_services
 FOR EACH ROW
 EXECUTE PROCEDURE update_user_services_updated_at();
