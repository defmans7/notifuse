.PHONY: build test-unit run clean keygen test-service test-repo test-http dev coverage docker-build docker-run docker-stop docker-clean docker-logs docker-compose-up docker-compose-down docker-compose-build

build:
	go build -o bin/server ./cmd/api

test-unit:
	go test -v ./internal/domain  ./internal/http ./internal/service ./internal/service/broadcast ./internal/repository

test-domain:
	go test -v ./internal/domain

test-service:
	go test -v ./internal/service ./internal/service/broadcast

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
	go run ./cmd/api

dev:
	air

clean:
	rm -rf bin/ tmp/ coverage.out coverage.html

keygen:
	go run cmd/keygen/main.go

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t notifuse:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -d --name notifuse \
		-p 8080:8080 \
		-e PASETO_PRIVATE_KEY=$${PASETO_PRIVATE_KEY} \
		-e PASETO_PUBLIC_KEY=$${PASETO_PUBLIC_KEY} \
		-e SECRET_KEY=$${SECRET_KEY} \
		-e ROOT_EMAIL=$${ROOT_EMAIL:-admin@example.com} \
		-e API_ENDPOINT=$${API_ENDPOINT:-http://localhost:8080} \
		-e WEBHOOK_ENDPOINT=$${WEBHOOK_ENDPOINT:-http://localhost:8080} \
		notifuse:latest

docker-stop:
	@echo "Stopping Docker container..."
	docker stop notifuse || true
	docker rm notifuse || true

docker-clean: docker-stop
	@echo "Removing Docker image..."
	docker rmi notifuse:latest || true

docker-logs:
	@echo "Showing Docker container logs..."
	docker logs -f notifuse

# Docker compose commands
docker-compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose down

docker-compose-build:
	@echo "Building services with Docker Compose..."
	docker-compose build

.DEFAULT_GOAL := build 