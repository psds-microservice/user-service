package command

import (
	"github.com/psds-microservice/user-service/internal/database"
)

// MigrateUp выполняет миграции вверх (разовая команда).
func MigrateUp(databaseURL string) error {
	return database.MigrateUp(databaseURL)
}
