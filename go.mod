module github.com/psds-microservice/user-service

go 1.24.0

toolchain go1.24.4

// Общие типы и proto (psds-microservice/helpy), генерация proto (psds-microservice/infra)
// Константы (paths, roles, permissions, http) — github.com/psds-microservice/helpy
replace github.com/psds-microservice/helpy => ../helpy

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7
	github.com/joho/godotenv v1.5.1
	github.com/psds-microservice/helpy v0.0.0
	github.com/spf13/cobra v1.10.2
	github.com/swaggo/http-swagger v1.3.4
	golang.org/x/crypto v0.47.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260203192932-546029d2fa20
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	gorm.io/datatypes v1.2.7
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/spec v0.22.3 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.11.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	github.com/swaggo/swag v1.16.6 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
)
