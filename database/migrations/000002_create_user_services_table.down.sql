DROP TRIGGER IF EXISTS trigger_update_user_services_updated_at ON user_services;
DROP FUNCTION IF EXISTS update_user_services_updated_at();
DROP INDEX IF EXISTS idx_user_services_priority;
DROP INDEX IF EXISTS idx_user_services_is_active;
DROP INDEX IF EXISTS idx_user_services_service_type;
DROP INDEX IF EXISTS idx_user_services_user_id;
DROP TABLE IF EXISTS user_services;
