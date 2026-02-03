package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateUp runs all pending migrations. It looks for database/migrations in cwd or parent (when run from bin/).
func MigrateUp(databaseURL string) error {
	cwd, _ := os.Getwd()
	dirs := []string{
		filepath.Join(cwd, "database", "migrations"),
		filepath.Join(cwd, "..", "database", "migrations"),
	}
	var absDir string
	for _, d := range dirs {
		if _, err := os.Stat(d); err == nil {
			absDir, _ = filepath.Abs(d)
			break
		}
	}
	if absDir == "" {
		return fmt.Errorf("migrations dir not found (tried cwd and parent)")
	}
	sourceURL := "file://" + filepath.ToSlash(absDir)
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	if err == migrate.ErrNoChange {
		log.Println("migrate: no pending migrations")
	} else {
		log.Println("migrate: up ok")
	}
	return nil
}
