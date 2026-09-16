.PHONY: build test run tidy fmt lint cover-check ci

BIN_DIR := bin
COVER_MIN := 80
COVER_PKGS := ./internal/clix

build:
	go build -mod=mod -o $(BIN_DIR)/wuji ./cmd/wuji

test:
	go test ./...

run:
	go run ./cmd/wuji

tidy:
	go mod tidy

fmt:
	@test -z "$$(gofmt -l .)"

lint:
	golangci-lint run ./...

cover-check:
	@chmod +x scripts/check-cover-min.sh
	@scripts/check-cover-min.sh $(COVER_MIN) $(COVER_PKGS)

ci: fmt lint build test cover-check
