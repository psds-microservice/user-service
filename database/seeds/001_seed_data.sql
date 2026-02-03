-- Миграция 005: Начальные данные (из haqury/user-service db/seeds/001_seed_data.sql)

-- Тестовый пользователь (если не существует)
INSERT INTO users (username, email, phone, password_hash)
SELECT 'testuser', 'test@example.com', '+1234567890', crypt('password123', gen_salt('bf', 10))
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'testuser');

-- Получаем ID и добавляем сервисы
DO $$
DECLARE
 user_uuid UUID;
BEGIN
 SELECT id INTO user_uuid FROM users WHERE username = 'testuser';

 IF user_uuid IS NOT NULL THEN
 -- Добавляем сервисы (3 кнопки)
 INSERT INTO user_services (user_id, service_name, service_type, base_url, port, priority, enabled_buttons) VALUES
 (user_uuid, 'primary_stream', 'streaming', 'http://streaming-service-1.internal', 8081, 1, '["button1"]'),
 (user_uuid, 'backup_stream', 'streaming', 'http://streaming-service-2.internal', 8082, 2, '["button2"]'),
 (user_uuid, 'recording_analytics', 'recording', 'http://recording-service.internal', 8083, 3, '["button3"]')
 ON CONFLICT (user_id, service_name) DO NOTHING;

 RAISE NOTICE '✅ Сервисы добавлены для testuser';
 END IF;
END $$;
