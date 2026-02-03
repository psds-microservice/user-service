package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	AppHost  string // APP_HOST, bind address (e.g. 0.0.0.0)
	HTTPPort string // APP_PORT or HTTP_PORT
	GRPCPort string // GRPC_PORT or METRICS_PORT
	AppEnv   string // APP_ENV
	AppDebug bool   // APP_DEBUG
	LogLevel string // LOG_LEVEL
	DB       struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
		SSLMode  string
	}
}

func Load() (*Config, error) {
	c := &Config{
		AppHost:  getEnv("APP_HOST", "0.0.0.0"),
		HTTPPort: firstEnv("APP_PORT", "HTTP_PORT", "8080"),
		GRPCPort: firstEnv("GRPC_PORT", "METRICS_PORT", "9090"),
		AppEnv:   getEnv("APP_ENV", "development"),
		AppDebug: getEnv("APP_DEBUG", "false") == "true",
		LogLevel: getEnv("LOG_LEVEL", "info"),
		DB: struct {
			Host     string
			Port     string
			User     string
			Password string
			Database string
			SSLMode  string
		}{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "user_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
	return c, nil
}

// firstEnv returns the first non-empty env value from keys, else def (last argument).
func firstEnv(keysAndDef ...string) string {
	if len(keysAndDef) == 0 {
		return ""
	}
	def := keysAndDef[len(keysAndDef)-1]
	keys := keysAndDef[:len(keysAndDef)-1]
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return def
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.Database, c.DB.SSLMode)
}

// DatabaseURL returns postgres URL for golang-migrate (postgres://user:pass@host:port/dbname?sslmode=...).
func (c *Config) DatabaseURL() string {
	pass := url.QueryEscape(c.DB.Password)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DB.User, pass, c.DB.Host, c.DB.Port, c.DB.Database, c.DB.SSLMode)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
