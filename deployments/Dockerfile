# No replace, no go.work: inject local modules into GOMODCACHE so go mod download finds them.
FROM golang:1.26-alpine AS builder
ENV GOMODCACHE=/go/pkg/mod
WORKDIR /app
COPY helpy /go/pkg/mod/github.com/psds-microservice/helpy@v0.1.0
COPY infra /go/pkg/mod/github.com/psds-microservice/infra@v0.0.3
COPY user-service ./user-service
WORKDIR /app/user-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /user-service ./cmd/user-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /user-service .
COPY --from=builder /app/user-service/database ./database
EXPOSE 8080 9090
CMD ["./user-service", "api"]
