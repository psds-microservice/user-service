package cmd

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/psds-microservice/user-service/internal/command"
	"github.com/psds-microservice/user-service/internal/config"
	"github.com/psds-microservice/user-service/internal/database"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeds (migrate up first, then seed)",
	RunE:  runSeed,
}

func runSeed(cmd *cobra.Command, args []string) error {
	if err := godotenv.Load(".env"); err != nil {
		_ = godotenv.Load("../.env")
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := command.MigrateUp(cfg.DatabaseURL()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	db, err := database.Open(cfg.DSN())
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	if err := command.Seed(db); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	log.Println("seed: ok")
	return nil
}
