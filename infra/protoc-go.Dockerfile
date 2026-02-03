# From psds-microservice/infra — единый образ для генерации Go из .proto
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache \
  protobuf \
  protobuf-dev \
  git \
  make \
  curl

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0 && \
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

RUN mkdir -p /include && \
  wget -q -O /tmp/googleapis.zip https://github.com/googleapis/googleapis/archive/master.zip && \
  unzip -q /tmp/googleapis.zip -d /tmp && \
  mv /tmp/googleapis-master/google /include/ && \
  rm -rf /tmp/googleapis.zip /tmp/googleapis-master

WORKDIR /workspace
# No ENTRYPOINT — команда передаётся из Makefile (make proto-generate).
# При использовании образа из psds-microservice/infra или helpy — укажите его в PROTOC_IMAGE.
