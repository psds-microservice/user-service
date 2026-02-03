# API (OpenAPI/Swagger из proto)

Документация REST API генерируется **из .proto** — единый источник правды для gRPC и HTTP.

## Генерация OpenAPI

В `user_service.proto` заданы маппинги `google.api.http` для каждого RPC. Спека собирается так:

```bash
# Установить плагин (один раз)
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Сгенерировать api/openapi.json
make proto-openapi
```

Нужны: `protoc`, `third_party/google/api/` (http.proto, annotations.proto).

## Просмотр

- После запуска сервиса: **http://localhost:8080/swagger/index.html**
- Спека: **http://localhost:8080/swagger/openapi.json**

Пути и методы в спецификации соответствуют аннотациям в `pkg/user_service/user_service.proto`.
