-- PSDS: роли, операторы, доступность, профиль (promt.txt)

ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'client'
  CHECK (role IN ('client', 'operator', 'admin'));
ALTER TABLE users ADD COLUMN IF NOT EXISTS operator_status VARCHAR(20) DEFAULT 'pending'
  CHECK (operator_status IN ('pending', 'verified', 'blocked'));
ALTER TABLE users ADD COLUMN IF NOT EXISTS max_sessions INTEGER DEFAULT 1;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT FALSE;

ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(10) DEFAULT 'en';

ALTER TABLE users ADD COLUMN IF NOT EXISTS company VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS specialization VARCHAR(255);

ALTER TABLE users ADD COLUMN IF NOT EXISTS total_sessions INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS rating DECIMAL(3,2) DEFAULT 0.0;

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_online BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP WITH TIME ZONE;

-- Расширяем размер полей под promt
ALTER TABLE users ALTER COLUMN phone TYPE VARCHAR(50) USING phone::VARCHAR(50);
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(100) USING username::VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_users_role_status_available ON users(role, operator_status, is_available);
CREATE INDEX IF NOT EXISTS idx_users_online ON users(is_online, last_seen_at);
