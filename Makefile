.PHONY: help dev prod build clean test test-integration migrate-up migrate-down

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

# Production (see DEPLOY.md). Stamps the API with the git version.
prod: ## Build & run the production stack (SPA + API + DB behind Caddy)
	VERSION=$(VERSION) docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build

prod-down: ## Stop production environment
	docker compose -f docker-compose.prod.yml --env-file .env.prod down

prod-logs: ## View production logs
	docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f api web

# Staging (see DEPLOY.md). Production-shaped, throwaway; ENV=staging permits mocks.
staging: ## Build & run the staging stack (coexists with prod on ports 8090/8453)
	VERSION=$(VERSION) docker compose -f docker-compose.staging.yml --env-file .env.staging up -d --build

staging-down: ## Stop staging (add ARGS=-v to wipe its data)
	docker compose -f docker-compose.staging.yml --env-file .env.staging down $(ARGS)

staging-logs: ## View staging logs
	docker compose -f docker-compose.staging.yml --env-file .env.staging logs -f api web

# Local development (without Docker)
run: ## Run locally without Docker
	go run ./cmd/api

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build: ## Build binary (version stamped from git describe)
	go build -ldflags "-X main.version=$(VERSION)" -o bin/food-delivery ./cmd/api

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Integration tests run the real Postgres + Redis adapters against throwaway
# containers in Docker. They are gated behind the `integration` build tag, so
# the plain `make test` above never needs services.
TEST_DB_URL ?= postgres://postgres:postgres@localhost:5433/food_delivery_test?sslmode=disable
TEST_REDIS_URL ?= redis://localhost:6380/0

test-integration: ## Run repository + redisstore integration tests against Dockerized Postgres + Redis
	docker compose -f docker-compose.test.yml up -d --wait
	@TEST_DATABASE_URL="$(TEST_DB_URL)" TEST_REDIS_URL="$(TEST_REDIS_URL)" \
		go test -tags=integration ./internal/repository/... ./internal/redisstore/... ; \
		status=$$? ; \
		docker compose -f docker-compose.test.yml down -v ; \
		exit $$status

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

# Swagger documentation
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go --output docs
	@echo "✓ Swagger docs generated! Visit http://localhost:8080/swagger/index.html"

swagger-fmt: ## Format Swagger annotations
	@echo "Formatting Swagger annotations..."
	swag fmt
	@echo "✓ Swagger annotations formatted"

swagger-clean: ## Clean Swagger docs
	@echo "Cleaning Swagger docs..."
	rm -rf docs/
	@echo "✓ Swagger docs cleaned"

install-swag: ## Install swag CLI
	@echo "Installing swag..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ Swag installed"

# Docker cleanup
docker-clean: ## Remove all containers and volumes
	docker compose -f docker-compose.dev.yml down -v
	docker compose -f docker-compose.prod.yml down -v
