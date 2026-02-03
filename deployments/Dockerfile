# Образ API (по умолчанию). Для отдельных образов см. build/api/Dockerfile, build/migrate/Dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /user-service ./cmd/user-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /user-service .
EXPOSE 8080 9090
CMD ["./user-service", "api"]
