package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User — сущность пользователя (схема БД: users, PSDS).
type User struct {
	ID             string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username       string `gorm:"size:100;uniqueIndex;not null"`
	Email          string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash   string `gorm:"column:password_hash;size:255;not null"`
	Role           string `gorm:"size:20;not null;default:client"`                // client, operator, admin
	OperatorStatus string `gorm:"column:operator_status;size:20;default:pending"` // pending, verified, blocked
	MaxSessions    int    `gorm:"column:max_sessions;default:1"`
	IsAvailable    bool   `gorm:"column:is_available;default:false"`

	FullName       string `gorm:"column:full_name;size:255"`
	AvatarURL      string `gorm:"column:avatar_url;size:500"`
	Timezone       string `gorm:"size:50"`
	Language       string `gorm:"size:10;default:en"`
	Phone          string `gorm:"size:50"`
	Company        string `gorm:"size:255"`
	Specialization string `gorm:"size:255"`

	TotalSessions int     `gorm:"column:total_sessions;default:0"`
	Rating        float64 `gorm:"type:decimal(3,2);default:0"`

	IsActive   bool       `gorm:"column:is_active;default:true"`
	IsOnline   bool       `gorm:"column:is_online;default:false"`
	LastSeenAt *time.Time `gorm:"column:last_seen_at"`

	Status          string         `gorm:"size:20;default:active"`
	Settings        datatypes.JSON `gorm:"type:jsonb"`
	StreamingConfig datatypes.JSON `gorm:"column:streaming_config;type:jsonb"`
	Stats           datatypes.JSON `gorm:"type:jsonb"`
	Metadata        datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLogin       *time.Time
	LastActivity    *time.Time
}

func (User) TableName() string { return "users" }

// UserSession — сессия пользователя (связь с session-manager).
type UserSession struct {
	ID                   string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               string     `gorm:"type:uuid;not null;index"`
	SessionType          string     `gorm:"column:session_type;size:50;not null"`
	SessionExternalID    string     `gorm:"column:session_external_id;size:255;not null"`
	ParticipantRole      string     `gorm:"column:participant_role;size:20;not null"`
	JoinedAt             time.Time  `gorm:"column:joined_at"`
	LeftAt               *time.Time `gorm:"column:left_at"`
	DurationSeconds      int        `gorm:"column:duration_seconds;default:0"`
	ConsultationRating   *int       `gorm:"column:consultation_rating"`
	ConsultationFeedback string     `gorm:"column:consultation_feedback;type:text"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (UserSession) TableName() string { return "user_sessions" }

// UserDevice — устройство пользователя (WebSocket/подключения).
type UserDevice struct {
	ID                string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            string     `gorm:"type:uuid;not null;index"`
	DeviceID          string     `gorm:"column:device_id;size:255;not null"`
	DeviceType        string     `gorm:"column:device_type;size:50;not null"`
	UserAgent         string     `gorm:"column:user_agent;type:text"`
	IPAddress         string     `gorm:"column:ip_address"`
	ConnectionID      string     `gorm:"column:connection_id;size:255"`
	IsConnected       bool       `gorm:"column:is_connected;default:false"`
	LastHeartbeat     *time.Time `gorm:"column:last_heartbeat"`
	SupportsWebRTC    bool       `gorm:"column:supports_webrtc;default:false"`
	SupportsWebSocket bool       `gorm:"column:supports_websocket;default:true"`
	BandwidthLimit    *int       `gorm:"column:bandwidth_limit"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (UserDevice) TableName() string { return "user_devices" }

// UserService — сервис пользователя (схема БД: user_services).
type UserService struct {
	ID                 string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string         `gorm:"type:uuid;not null;index"`
	ServiceName        string         `gorm:"column:service_name;size:100;not null"`
	ServiceType        string         `gorm:"column:service_type;size:50;not null"` // streaming, recording, monitoring, analytics, custom
	BaseURL            string         `gorm:"column:base_url;size:500;not null"`
	APIEndpoint        string         `gorm:"column:api_endpoint;size:200;default:/api/v1"`
	Port               int            `gorm:"not null"`
	UseSSL             bool           `gorm:"column:use_ssl;default:false"`
	SSLCertificatePath string         `gorm:"column:ssl_certificate_path;size:500"`
	RoutingKey         string         `gorm:"column:routing_key;size:100"`
	QueueName          string         `gorm:"column:queue_name;size:100"`
	TopicName          string         `gorm:"column:topic_name;size:100"`
	MaxBitrate         int            `gorm:"column:max_bitrate;default:5000"`
	MaxConnections     int            `gorm:"column:max_connections;default:10"`
	Priority           int            `gorm:"default:1"`
	IsActive           bool           `gorm:"column:is_active;default:true"`
	EnabledButtons     datatypes.JSON `gorm:"column:enabled_buttons;type:jsonb"`
	Schedule           datatypes.JSON `gorm:"type:jsonb"`
	Parameters         datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (UserService) TableName() string { return "user_services" }

// Base — общие поля для сущностей с автоинкрементом (если понадобятся другие таблицы).
// Для users/user_services используем UUID и явные timestamps.
type Base struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
