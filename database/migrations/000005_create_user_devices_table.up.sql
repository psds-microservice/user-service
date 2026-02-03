-- user_devices: устройства и WebSocket (promt.txt)

CREATE TABLE IF NOT EXISTS user_devices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  device_id VARCHAR(255) NOT NULL,
  device_type VARCHAR(50) NOT NULL,
  user_agent TEXT,
  ip_address INET,

  connection_id VARCHAR(255),
  is_connected BOOLEAN DEFAULT FALSE,
  last_heartbeat TIMESTAMP WITH TIME ZONE,

  supports_webrtc BOOLEAN DEFAULT FALSE,
  supports_websocket BOOLEAN DEFAULT TRUE,
  bandwidth_limit INTEGER,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(user_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_connected ON user_devices(user_id, is_connected);
CREATE INDEX IF NOT EXISTS idx_user_devices_connection_id ON user_devices(connection_id);

CREATE OR REPLACE FUNCTION update_user_devices_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_user_devices_updated_at ON user_devices;
CREATE TRIGGER trigger_update_user_devices_updated_at
  BEFORE UPDATE ON user_devices
  FOR EACH ROW
  EXECUTE PROCEDURE update_user_devices_updated_at();
