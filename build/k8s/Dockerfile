# Production image for Kubernetes
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /user-service ./cmd/user-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /user-service .
USER nobody:nobody
EXPOSE 8080 9090
CMD ["./user-service"]
