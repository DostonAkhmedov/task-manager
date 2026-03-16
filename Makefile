.PHONY: help build run test clean docker-build docker-up docker-down lint fmt lint-install

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make lint         - Run golangci-lint (v2.4.0+)"
	@echo "  make lint-install - Install golangci-lint v2.4.0+"
	@echo "  make fmt          - Format code with go fmt"

build:
	@echo "Building application..."
	go build -o bin/api cmd/main.go

run: build
	@echo "Running application..."
	./bin/api

test:
	@echo "Running tests..."
	go test -v ./...

test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build:
	@echo "Building Docker image..."
	docker compose build

docker-up:
	@echo "Starting Docker containers..."
	docker compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker compose down

lint:
	@echo "Running golangci-lint (v2.4.0+)..."
	golangci-lint run ./... --no-config

lint-install:
	@echo "Installing golangci-lint v2.4.0+..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

fmt:
	@echo "Formatting code..."
	go fmt ./...
