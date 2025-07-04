# E-Commerce Docker Makefile

.PHONY: help build up down logs clean restart test test-unit test-integration test-coverage test-race test-bench test-docker test-ci

# Varsayılan hedef
help:
	@echo "E-Commerce Docker Commands:"
	@echo "  make build    - Docker image'larını oluştur"
	@echo "  make up       - Container'ları başlat"
	@echo "  make down     - Container'ları durdur"
	@echo "  make logs     - Logları göster"
	@echo "  make restart  - Container'ları yeniden başlat"
	@echo "  make clean    - Tüm container'ları ve volume'ları temizle"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test           - Tüm testleri çalıştır"
	@echo "  make test-unit      - Unit testleri çalıştır"
	@echo "  make test-integration - Integration testleri çalıştır"
	@echo "  make test-coverage  - Coverage raporu ile testleri çalıştır"
	@echo "  make test-race      - Race condition testleri çalıştır"
	@echo "  make test-bench     - Benchmark testleri çalıştır"
	@echo "  make test-docker    - Docker ile testleri çalıştır"
	@echo "  make test-ci        - CI/CD için testleri çalıştır"
	@echo "  make test-clean     - Test sonuçlarını temizle"

# Docker image'larını oluştur
build:
	docker-compose build

# Container'ları başlat
up:
	docker-compose up -d

# Container'ları durdur
down:
	docker-compose down

# Logları göster
logs:
	docker-compose logs -f

# Container'ları yeniden başlat
restart:
	docker-compose down
	docker-compose up -d --build

# Temizlik
clean:
	docker-compose down -v
	docker system prune -f

# Development mode (log output ile)
dev:
	docker-compose up --build

# Sadece API container'ını yeniden başlat
api-restart:
	docker-compose restart api

# Veritabanına bağlan
db-connect:
	docker exec -it ecommerce_db psql -U postgres -d ecommerce_db

# Test Commands
# =============

# Tüm testleri çalıştır
test:
	@echo "🧪 Running all tests..."
	@mkdir -p test-results
	@chmod +x scripts/run-tests.sh
	@bash scripts/run-tests.sh

# Unit testleri çalıştır
test-unit:
	@echo "🧪 Running unit tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./internal/... 2>&1 | tee test-results/unit-test.log

# Integration testleri çalıştır
test-integration:
	@echo "🧪 Running integration tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./tests/... 2>&1 | tee test-results/integration-test.log

# Coverage raporu ile testleri çalıştır
test-coverage:
	@echo "📊 Running tests with coverage..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./... -coverprofile=test-results/coverage.out
	@go tool cover -html=test-results/coverage.out -o test-results/coverage.html
	@go tool cover -func=test-results/coverage.out | grep total
	@echo "Coverage report: test-results/coverage.html"

# Race condition testleri çalıştır
test-race:
	@echo "🏁 Running race condition tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -race -v ./internal/service/... 2>&1 | tee test-results/race-test.log

# Benchmark testleri çalıştır
test-bench:
	@echo "⚡ Running benchmark tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -bench=. -benchmem ./internal/api/handler/... 2>&1 | tee test-results/benchmark.log

# Docker ile testleri çalıştır
test-docker:
	@echo "🐳 Running tests with Docker..."
	@chmod +x scripts/docker/run-tests.sh
	@bash scripts/docker/run-tests.sh

# CI/CD için testleri çalıştır
test-ci:
	@echo "🔄 Running CI tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./... -coverprofile=test-results/coverage.out -race
	@go tool cover -html=test-results/coverage.out -o test-results/coverage.html
	@go tool cover -func=test-results/coverage.out

# Test sonuçlarını temizle
test-clean:
	@echo "🧹 Cleaning test results..."
	@rm -rf test-results
	@echo "Test results cleaned!"

# Model testleri
test-model:
	@echo "📦 Running model tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./internal/model/... -coverprofile=test-results/model-coverage.out

# Service testleri
test-service:
	@echo "⚙️ Running service tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./internal/service/... -coverprofile=test-results/service-coverage.out

# Handler testleri
test-handler:
	@echo "🌐 Running handler tests..."
	@mkdir -p test-results
	@GO_ENV=test GIN_MODE=test go test -v ./internal/api/handler/... -coverprofile=test-results/handler-coverage.out

# Test watch mode (otomatik yeniden çalıştırma)
test-watch:
	@echo "👀 Running tests in watch mode..."
	@which fswatch > /dev/null || (echo "fswatch not found. Install with: brew install fswatch" && exit 1)
	@fswatch -o . | xargs -n1 -I{} make test-unit

# Test başarısını kontrol et
test-check:
	@echo "✅ Checking test results..."
	@if [ -f "test-results/coverage.out" ]; then \
		COVERAGE=$$(go tool cover -func=test-results/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Coverage: $$COVERAGE%"; \
		if [ $$(echo "$$COVERAGE < 70" | bc -l) -eq 1 ]; then \
			echo "❌ Coverage $$COVERAGE% is below threshold 70%"; \
			exit 1; \
		else \
			echo "✅ Coverage $$COVERAGE% meets threshold"; \
		fi; \
	else \
		echo "❌ Coverage file not found"; \
		exit 1; \
	fi

# Dependency management
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy

# Code formatting
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...

# Code linting
lint:
	@echo "🔍 Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	@golangci-lint run ./...

# Security scan
security:
	@echo "🔐 Running security scan..."
	@which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec ./...

# Full quality check
quality: deps fmt lint security test-coverage
	@echo "🎯 Quality check completed!"

# Development setup
setup:
	@echo "🚀 Setting up development environment..."
	@make deps
	@make fmt
	@chmod +x scripts/run-tests.sh
	@chmod +x scripts/docker/run-tests.sh
	@chmod +x scripts/windows/run-tests.ps1
	@mkdir -p test-results
	@echo "✅ Development environment ready!"
