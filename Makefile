PKG             ?= ./...
INTEGRATION_PKG ?= ./src/internal/infrastructure/data

.PHONY: test unit test-integration test-all

api:
	go run ./src/cmd/http_api

swagger:
	swag init -g src/cmd/http_api/main.go -o src/internal/http_api/swagger

test: unit

unit:
	go test $(PKG)

test-integration:
	go test -tags=integration -count=1 $(INTEGRATION_PKG)

test-all:
	go test $(PKG)
	go test -tags=integration -count=1 $(INTEGRATION_PKG)

lint:
	golangci-lint run --fix