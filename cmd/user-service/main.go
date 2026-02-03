// Package main — точка входа user-service (HTTP + gRPC).
// OpenAPI/Swagger генерируется из proto: make proto-openapi
package main

import (
	"log"

	"github.com/psds-microservice/user-service/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
