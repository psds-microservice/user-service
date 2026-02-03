DROP TRIGGER IF EXISTS trigger_update_user_devices_updated_at ON user_devices;
DROP FUNCTION IF EXISTS update_user_devices_updated_at();
DROP INDEX IF EXISTS idx_user_devices_connection_id;
DROP INDEX IF EXISTS idx_user_devices_user_connected;
DROP TABLE IF EXISTS user_devices;
