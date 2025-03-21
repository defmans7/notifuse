.PHONY: build test run clean keygen test-service test-repo test-http dev coverage

build:
	go build -o bin/server cmd/api/main.go

test:
	go test -v ./...

test-service:
	go test -v ./internal/service

test-repo:
	go test -v ./internal/repository

test-http:
	go test -v ./internal/http

# Comprehensive test coverage command
coverage:
	@echo "Running API tests and generating coverage report..."
	@go test -coverprofile=coverage.out ./internal/...
	@echo "\n=== API Test Coverage Summary ==="
	@go tool cover -func=coverage.out | grep total
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Detailed HTML coverage report generated: coverage.html"

run:
	go run cmd/api/main.go

dev:
	air

clean:
	rm -rf bin/ tmp/ coverage.out coverage.html

migrate:
	psql -U postgres -d notifuse -f migrations/001_create_tables.sql

migrate-test:
	psql -U postgres -d notifuse_test -f migrations/001_create_tables.sql

keygen:
	go run cmd/keygen/main.go

.DEFAULT_GOAL := build 