-- user_sessions: связь с session-manager (promt.txt)

CREATE TABLE IF NOT EXISTS user_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_type VARCHAR(50) NOT NULL,
  session_external_id VARCHAR(255) NOT NULL,

  participant_role VARCHAR(20) NOT NULL CHECK (participant_role IN ('host', 'operator', 'viewer')),
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  left_at TIMESTAMP WITH TIME ZONE,
  duration_seconds INTEGER DEFAULT 0,

  consultation_rating INTEGER CHECK (consultation_rating IS NULL OR (consultation_rating BETWEEN 1 AND 5)),
  consultation_feedback TEXT,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_joined ON user_sessions(user_id, joined_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(user_id, left_at) WHERE left_at IS NULL;

CREATE OR REPLACE FUNCTION update_user_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_user_sessions_updated_at ON user_sessions;
CREATE TRIGGER trigger_update_user_sessions_updated_at
  BEFORE UPDATE ON user_sessions
  FOR EACH ROW
  EXECUTE PROCEDURE update_user_sessions_updated_at();
