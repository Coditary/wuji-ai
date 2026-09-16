.PHONY: build test run tidy

BIN_DIR := bin

build:
	go build -mod=mod -o $(BIN_DIR)/wuji ./cmd/wuji

test:
	go test ./...

run:
	go run ./cmd/wuji

tidy:
	go mod tidy
