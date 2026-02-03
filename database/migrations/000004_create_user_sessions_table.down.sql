DROP TRIGGER IF EXISTS trigger_update_user_sessions_updated_at ON user_sessions;
DROP FUNCTION IF EXISTS update_user_sessions_updated_at();
DROP INDEX IF EXISTS idx_user_sessions_active;
DROP INDEX IF EXISTS idx_user_sessions_user_joined;
DROP TABLE IF EXISTS user_sessions;
