package command

import (
	"github.com/psds-microservice/user-service/internal/database"
	"gorm.io/gorm"
)

// Seed выполняет сиды (разовая команда). Перед вызовом миграции должны быть применены.
func Seed(db *gorm.DB) error {
	return database.RunSeeds(db)
}
