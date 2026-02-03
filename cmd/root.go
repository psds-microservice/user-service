package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "user-service",
	Short: "User service: auth, profiles, operators, sessions",
	RunE:  runAPI, // по умолчанию — запуск API (для обратной совместимости)
}

// Execute запускает корневую команду (Cobra CLI).
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
}
