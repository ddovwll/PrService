PKG             ?= ./...
INTEGRATION_PKG ?= ./src/internal/infrastructure/data/integration_tests

.PHONY: test unit test-integration test-all

up:
	docker-compose up

api:
	go run ./src/cmd/http_api

migrate:
	go run ./src/cmd/migrator

swagger:
	swag init -g src/cmd/http_api/main.go -o src/internal/http_api/swagger

test: unit

unit:
	go test $(PKG)

test-integration:
	go test -tags=integration -count=1 $(INTEGRATION_PKG)

test-all: unit test-integration

lint:
	golangci-lint run --fix