.PHONY: help dev prod build clean test migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: ## Run development environment
	docker compose -f docker-compose.dev.yml up --build

dev-down: ## Stop development environment
	docker compose -f docker-compose.dev.yml down

dev-logs: ## View development logs
	docker compose -f docker-compose.dev.yml logs -f api

# Production
prod: ## Run production environment
	docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

prod-down: ## Stop production environment
	docker compose -f docker-compose.prod.yml down

prod-logs: ## View production logs
	docker compose -f docker-compose.prod.yml logs -f api

# Local development (without Docker)
run: ## Run locally without Docker
	go run ./cmd/api

build: ## Build binary
	go build -o bin/food-delivery ./cmd/api

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Database migrations
DB_URL ?= postgres://postgres:postgres123@localhost:5432/food_delivery?sslmode=disable

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Run database migrations down
	@echo "Rolling back last migration..."
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-version: ## Show current migration version
	migrate -path migrations -database "$(DB_URL)" version

migrate-create: ## Create new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		read -p "Migration name: " migration_name; \
		migrate create -ext sql -dir migrations -seq $$migration_name; \
	else \
		migrate create -ext sql -dir migrations -seq $(name); \
	fi
	@echo "✅ Migration files created in migrations/"

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/ tmp/ coverage.out

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Code quality
lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

# Docker cleanup
docker-clean: ## Remove all containers and volumes
	docker compose -f docker-compose.dev.yml down -v
	docker compose -f docker-compose.prod.yml down -v
