PKG             ?= ./...
INTEGRATION_PKG ?= ./src/internal/infrastructure/data

.PHONY: test unit test-integration test-all

test: unit

unit:
	go test $(PKG)

test-integration:
	go test -tags=integration -count=1 $(INTEGRATION_PKG)

test-all:
	go test $(PKG)
	go test -tags=integration -count=1 $(INTEGRATION_PKG)