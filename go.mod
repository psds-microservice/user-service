module github.com/psds-microservice/user-service

go 1.24.0

toolchain go1.24.4

// Общие типы и proto (psds-microservice/helpy), генерация proto (psds-microservice/infra)
// После генерации: go get github.com/psds-microservice/helpy@latest при необходимости

require (
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/joho/godotenv v1.5.1
	golang.org/x/crypto v0.47.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lib/pq v1.11.1 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260202165425-ce8ad4cf556b // indirect
)
