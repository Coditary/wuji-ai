.PHONY: build proto tidy test run install-tools

PROTO_DIR := api/proto/v1
GEN_DIR := api/proto/v1

build:
	go build -o bin/wuji ./cmd/wuji
	go build -o bin/wuji-driver-dummy ./cmd/wuji-driver-dummy

proto:
	buf generate

tidy:
	go mod tidy

test:
	go test ./...

run:
	go run ./cmd/wuji

install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
