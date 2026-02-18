.PHONY: help init build run run-dev migrate migrate-create worker test test-api test-db \
 version clean proto proto-build proto-generate proto-openapi proto-pkg proto-pkg-simple proto-pkg-script proto-clean proto-help lint vet fmt docker-build \
 docker-run docker-compose-up docker-compose-down install-deps health-check \
 update generate-docs bench load-test security-check dev

# Конфигурация
APP_NAME = user-service
BIN_DIR = bin
BUILD_INFO = $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT_HASH = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
PROTOC_IMAGE = local/protoc-go:latest
PROTO_ROOT = pkg/user_service
# Сгенерированные Go-файлы из proto (go_package → pkg/gen/user_service при module=)
GEN_DIR = pkg/gen/user_service
GO_MODULE = github.com/psds-microservice/user-service
# OpenAPI/Swagger из proto (protoc-gen-openapiv2)
OPENAPI_OUT = api
OPENAPI_SPEC = $(OPENAPI_OUT)/openapi.json

# Главная цель по умолчанию
.DEFAULT_GOAL := help

## 📚 Помощь
help:
	@echo "🚀 user-service - Makefile"
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "📦 Proto файлы:"
	@echo "  make proto              - Build image and generate all proto files"
	@echo "  make proto-generate     - Generate code for internal use"
	@echo "  make proto-pkg          - Generate code for external services"
	@echo "  make proto-pkg-simple   - Simple version for Windows"
	@echo "  make proto-pkg-script   - Generate via script (recommended)"
	@echo "  make proto-clean        - Clean generated files"
	@echo ""
	@echo "🏗️ Сборка и запуск:"
	@echo "  make build    - Сборка бинарника"
	@echo "  make run      - Сборка и запуск сервера"
	@echo "  make run-dev  - Запуск в режиме разработки"
	@echo "  make dev      - Запуск с hot reload (требуется air)"
	@echo "  make clean    - Очистка сборки"
	@echo ""
	@echo "🔧 Управление:"
	@echo "  make migrate        - Выполнить миграции БД"
	@echo "  make migrate-create - Создать новую миграцию"
	@echo "  make worker         - Запустить фоновых воркеров"
	@echo "  make health-check   - Проверить здоровье сервиса"
	@echo ""
	@echo "🧪 Тестирование:"
	@echo "  make test           - Запуск всех тестов"
	@echo "  make test-api       - Тестирование API"
	@echo "  make test-db        - Тестирование БД"
	@echo "  make bench          - Бенчмарки"
	@echo "  make load-test      - Нагрузочное тестирование"
	@echo "  make lint           - Линтинг кода"
	@echo "  make vet            - Проверка кода"
	@echo "  make fmt            - Форматирование кода"
	@echo "  make security-check - Проверка безопасности"
	@echo "  make proto-openapi    - Сгенерировать OpenAPI/Swagger из .proto"
	@echo "  make proto-http-paths - Сгенерировать константы путей из proto (router)"
	@echo ""

## 📄 OpenAPI/Swagger из proto (единый источник правды — .proto)
proto-openapi:
	@command -v protoc >/dev/null 2>&1 || (echo "Установите protoc" && exit 1); \
	command -v protoc-gen-openapiv2 >/dev/null 2>&1 || (echo "Установите: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest" && exit 1)
	@mkdir -p $(OPENAPI_OUT)
	@PATH="$$(go env GOPATH)/bin:$$PATH"; \
	protoc -I $(PROTO_ROOT) -I third_party \
		--openapiv2_out=$(OPENAPI_OUT) \
		--openapiv2_opt=logtostderr=true \
		--openapiv2_opt=allow_merge=true \
		--openapiv2_opt=merge_file_name=openapi \
		$(PROTO_ROOT)/user_service.proto
	@if [ -f $(OPENAPI_OUT)/openapi.swagger.json ]; then cp $(OPENAPI_OUT)/openapi.swagger.json $(OPENAPI_OUT)/openapi.json; echo "✅ OpenAPI: $(OPENAPI_SPEC)"; elif [ -f $(OPENAPI_OUT)/openapi.json ]; then echo "✅ OpenAPI: $(OPENAPI_SPEC)"; else echo "⚠ Проверьте вывод protoc выше"; fi

## 🛣️ Константы HTTP-путей из proto (google.api.http) — единственный источник путей для router
proto-http-paths:
	@go run scripts/gen_http_paths.go

## 📦 Proto файлы (образ из vendor infra)
proto: proto-build proto-generate

# Сборка образа: из vendored infra (go mod vendor)
INFRA_VENDOR := vendor/github.com/psds-microservice/infra
proto-build:
	@echo "📦 Building protoc-go image (from $(INFRA_VENDOR))..."
	@docker build -t $(PROTOC_IMAGE) -f $(INFRA_VENDOR)/protoc-go.Dockerfile $(INFRA_VENDOR)
	@echo "✅ Docker image built"

# Генерация: сначала пробуем локальный protoc (PATH + go install плагины), иначе Docker с обходом entrypoint
proto-generate:
	@PATH="$$(go env GOPATH 2>/dev/null)/bin:$$PATH"; \
	if command -v protoc >/dev/null 2>&1 && command -v protoc-gen-go >/dev/null 2>&1 && command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		$(MAKE) proto-generate-local; \
	else \
		$(MAKE) proto-generate-docker; \
	fi

# Локальная генерация: protoc + go + grpc + grpc-gateway (как в contract-service: все RPC сразу в HTTP по proto).
proto-generate-local:
	@echo "🔧 Generating Go code (local protoc)..."
	@mkdir -p $(GEN_DIR)
	@command -v protoc-gen-grpc-gateway >/dev/null 2>&1 || (echo "Установите: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest" && exit 1)
	@PATH="$$(go env GOPATH)/bin:$$PATH"; \
	for f in $(PROTO_ROOT)/*.proto; do \
		[ -f "$$f" ] || continue; \
		echo "📁 Processing: $$f"; \
		protoc -I $(PROTO_ROOT) -I third_party \
			--go_out=. --go_opt=module=$(GO_MODULE) \
			--go-grpc_out=. --go-grpc_opt=module=$(GO_MODULE) \
			--grpc-gateway_out=. --grpc-gateway_opt=module=$(GO_MODULE) \
			"$$f" || exit 1; \
	done
	@echo "✅ Generated in $(GEN_DIR)"

# Docker: обходим entrypoint образа infra; добавляем grpc-gateway (образ должен содержать protoc-gen-grpc-gateway).
proto-generate-docker:
	@echo "🔧 Generating Go code (Docker)..."
	@mkdir -p $(GEN_DIR)
	@docker run --rm \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		--entrypoint sh \
		$(PROTOC_IMAGE) \
		-c ' \
		PROTO_ROOT="$(PROTO_ROOT)" MODULE="$(GO_MODULE)" && \
		find $$PROTO_ROOT -name "*.proto" 2>/dev/null | while read f; do \
		echo "📁 Processing: $$f" && \
		protoc -I $$PROTO_ROOT -I third_party -I /include \
		--go_out=. --go_opt=module=$$MODULE \
		--go-grpc_out=. --go-grpc_opt=module=$$MODULE \
		--grpc-gateway_out=. --grpc-gateway_opt=module=$$MODULE \
		"$$f" || exit 1; \
		done && echo "✅ Generated in $(GEN_DIR)" \
		'
	@echo "✅ Proto files generated"


proto-pkg:
	@echo "🚀 Generating for external services (in pkg/gen/)..."
	@mkdir -p pkg/gen
	@echo "Using Docker image: $(PROTOC_IMAGE)"
	@docker run --rm \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		--entrypoint sh \
		$(PROTOC_IMAGE) \
		-c ' \
		echo "Finding proto files..." && \
		find pkg/user_service -name "*.proto" | while read f; do \
		echo "Processing $$f" && \
		protoc -I pkg/user_service -I /include \
		--go_out=. --go_opt=module=github.com/psds-microservice/user-service \
		--go-grpc_out=. --go-grpc_opt=module=github.com/psds-microservice/user-service \
		"$$f" || exit 1; \
		done && \
		echo "✅ Shared library generated in $(GEN_DIR)" \
		'
	@echo "✅ Shared library generated"

proto-pkg-simple:
	@echo "🚀 Generating for external services (simple version)..."
	@mkdir -p pkg/gen
	@docker run --rm \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		--entrypoint sh \
		$(PROTOC_IMAGE) \
		-c 'find pkg/user_service -name "*.proto" -exec echo "Processing {}" \; -exec protoc -I pkg/user_service -I /include --go_out=. --go_opt=module=github.com/psds-microservice/user-service --go-grpc_out=. --go-grpc_opt=module=github.com/psds-microservice/user-service {} \;'
	@echo "✅ Shared library generated in pkg/gen/"

proto-pkg-script:
	@echo "🚀 Generating via script..."
	@docker run --rm \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		--entrypoint sh \
		$(PROTOC_IMAGE) \
		-c ' \
		PROTO_ROOT="pkg/user_service" && \
		mkdir -p $(GEN_DIR) && \
		find $$PROTO_ROOT -name "*.proto" | while read proto_file; do \
		echo "📁 Processing: $$proto_file" && \
		protoc -I pkg/user_service -I /include \
		--go_out=. --go_opt=module=github.com/psds-microservice/user-service \
		--go-grpc_out=. --go-grpc_opt=module=github.com/psds-microservice/user-service \
		"$$proto_file" || exit 1; \
		done && \
		echo "✅ Done! Check $(GEN_DIR)" \
		'
	@echo "✅ Generated via script"

proto-clean:
	@echo "🧹 Cleaning generated files..."
	@if exist "internal\gen" rmdir /s /q "internal\gen" 2>nul || rm -rf pkg/gen
	@if exist "pkg\gen" rmdir /s /q "pkg\gen" 2>nul || rm -rf pkg/gen
	@echo "✅ Clean complete"

## 🏗️ Сборка и запуск
build:
	@echo "🔨 Building $(APP_NAME)..."
	mkdir -p $(BIN_DIR)
	go build -ldflags="-X 'main.Version=$(BUILD_INFO)' \
		-X 'main.Commit=$(COMMIT_HASH)' \
		-X 'main.BuildDate=$(BUILD_DATE)'" \
		-o $(BIN_DIR)/$(APP_NAME) ./cmd/user-service
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

run: build
	@echo "🚀 Starting server..."
	@echo "Server will be available at: http://localhost:8080"
	@echo "Health check: http://localhost:8080/health"
	@echo ""
	@cd $(BIN_DIR) && ./$(APP_NAME)

run-dev:
	@echo "🚀 Starting in development mode..."
	@echo "For hot reload use: make dev"
	go run ./cmd/user-service

dev:
	@echo "🔥 Starting with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "⚠ air is not installed. Install: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		make run-dev; \
	fi

## 🔧 Управление
migrate: build
	@echo "🔄 Running migrations..."
	@cd $(BIN_DIR) && ./$(APP_NAME) migrate up

migrate-create: build
	@echo "📝 Creating migration..."
	@read -p "Enter migration name: " name; \
	cd $(BIN_DIR) && ./$(APP_NAME) migrate create --name $$name

seed: build
	@echo "🌱 Running seeds..."
	@cd $(BIN_DIR) && ./$(APP_NAME) seed

db-init: build
	@echo "🗄️ DB init (migrate + seed)..."
	@cd $(BIN_DIR) && ./$(APP_NAME) migrate up && ./$(APP_NAME) seed

worker: build
	@echo "👷 Starting workers..."
	@cd $(BIN_DIR) && ./$(APP_NAME) worker --workers 5

health-check:
	@echo "❤️ Health checking service..."
	@if curl -s http://localhost:8080/health > /dev/null; then \
		echo "✅ Service is running"; \
	else \
		echo "❌ Service is not available"; \
	fi

## 🧪 Тестирование
test: proto
	@echo "🧪 Running all tests..."
	go test -v -race ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo "✅ Tests completed"

bench:
	@echo "📊 Running benchmarks..."
	go test -bench=. -benchmem ./...

load-test:
	@echo "⚡ Running load tests..."
	@if command -v k6 > /dev/null; then \
		k6 run scripts/loadtest.js; \
	else \
		echo "⚠ k6 is not installed. Install: https://k6.io/docs/getting-started/installation/"; \
	fi

## 🛠️ Code quality
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠ golangci-lint is not installed"; \
	fi

vet:
	@echo "🔎 Checking code with vet..."
	go vet ./...
	@echo "✅ Vet completed"

fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Formatting completed"

security-check:
	@echo "🔒 Security checking..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "⚠ gosec is not installed. Install: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

## 📋 Утилиты
version: build
	@echo "📋 Version information:"
	@cd $(BIN_DIR) && ./$(APP_NAME) version

generate-docs: build
	@echo "📖 Generating documentation..."
	@cd $(BIN_DIR) && ./$(APP_NAME) generate docs
	@echo "✅ Documentation generated"

install-deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "✅ Dependencies installed"

update:
	@echo "🔄 Updating dependencies..."
	go get -u ./... \
		github.com/psds-microservice/helpy \
		github.com/psds-microservice/infra
	go mod tidy
	go mod vendor
	$(MAKE) proto
	@$(MAKE) proto-openapi 2>/dev/null || true
	@echo "✅ Dependencies updated"

init: install-deps proto
	@echo "✅ Project initialized"

clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BIN_DIR) coverage.out
	go clean
	@echo "✅ Clean completed"

tidy:
	go mod tidy

## 🌐 Dual API (HTTP + gRPC) — опционально для сервисов с gRPC
run-dual:
	@echo "🚀 Starting in DUAL mode (HTTP:8080 + gRPC:9090)..."
	@echo "HTTP REST: http://localhost:8080"
	@echo "gRPC: localhost:9090"
	@echo ""
	go run ./cmd/user-service

test-dual:
	@echo "🧪 Testing DUAL API..."
	@echo "1. Starting server..."
	@make run-dual &
	@SERVER_PID=$$!; sleep 3; echo ""; echo "2. Testing HTTP API..."; curl -s http://localhost:8080/health; echo ""; echo "✅ Dual API tests completed"; kill $$SERVER_PID 2>/dev/null || true

grpc-client:
	@echo "🚀 Running gRPC client..."
	@cd scripts/clients 2>/dev/null && go run test_grpc_client.go || echo "⚠ scripts/clients not found"

http-client:
	@echo "🌐 Running HTTP client..."
	@cd scripts/clients 2>/dev/null && python test_http_client.py || echo "⚠ scripts/clients not found"
