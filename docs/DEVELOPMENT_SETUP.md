# 🛠️ Development Setup Guide

## Overview

Bu kılavuz, Özgür Mutfak API projesinin geliştirme ortamının nasıl kurulacağını açıklar.

## 📋 Prerequisites

### Required Software

1. **Go Programming Language**
   ```bash
   # Windows (Chocolatey)
   choco install golang
   
   # macOS (Homebrew)
   brew install go
   
   # Linux (Ubuntu/Debian)
   sudo apt update
   sudo apt install golang-go
   
   # Verify installation
   go version  # Should show Go 1.21 or later
   ```

2. **PostgreSQL Database**
   ```bash
   # Windows (Chocolatey)
   choco install postgresql
   
   # macOS (Homebrew)
   brew install postgresql
   brew services start postgresql
   
   # Linux (Ubuntu/Debian)
   sudo apt install postgresql postgresql-contrib
   sudo systemctl start postgresql
   sudo systemctl enable postgresql
   ```

3. **Git Version Control**
   ```bash
   # Windows (Chocolatey)
   choco install git
   
   # macOS (Homebrew)
   brew install git
   
   # Linux (Ubuntu/Debian)
   sudo apt install git
   ```

4. **Docker (İsteğe bağlı)**
   ```bash
   # Windows: Docker Desktop'tan indirin
   # macOS: Docker Desktop'tan indirin
   # Linux
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   ```

5. **Code Editor (VS Code Önerilen)**
   ```bash
   # Windows (Chocolatey)
   choco install vscode
   
   # macOS (Homebrew)
   brew install --cask visual-studio-code
   
   # Linux (Snap)
   sudo snap install code --classic
   ```

## 🚀 Project Setup

### 1. Repository Clone

```bash
# Repository'yi clone edin
git clone https://github.com/yourusername/ozgur-mutfak.git
cd ozgur-mutfak

# Branch yapısını kontrol edin
git branch -a
```

### 2. Go Module Setup

```bash
# Dependencies'leri yükleyin
go mod download

# Go module'ları verify edin
go mod verify

# Unused dependencies'leri temizleyin
go mod tidy
```

### 3. Database Setup

#### Option A: Local PostgreSQL

```bash
# PostgreSQL'e bağlan
sudo -u postgres psql

# Database oluştur
CREATE DATABASE ozgur_mutfak_dev;
CREATE USER ozgur_dev WITH PASSWORD 'dev_password';
GRANT ALL PRIVILEGES ON DATABASE ozgur_mutfak_dev TO ozgur_dev;
\q
```

#### Option B: Docker PostgreSQL

```bash
# Docker ile PostgreSQL çalıştır
docker run --name postgres-dev \
  -e POSTGRES_DB=ozgur_mutfak_dev \
  -e POSTGRES_USER=ozgur_dev \
  -e POSTGRES_PASSWORD=dev_password \
  -p 5432:5432 \
  -d postgres:15-alpine

# Database'in çalıştığını kontrol et
docker ps
docker logs postgres-dev
```

### 4. Environment Configuration

```bash
# Development environment dosyası oluştur
cat > .env.development << EOF
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ozgur_mutfak_dev
DB_USER=ozgur_dev
DB_PASSWORD=dev_password
DB_SSL_MODE=disable

# API Configuration
API_PORT=8080
API_HOST=localhost
BASE_URL=http://localhost:8080

# JWT Configuration
JWT_SECRET=your-development-jwt-secret-min-32-characters
JWT_EXPIRY=24h

# Application Configuration
GIN_MODE=debug
LOG_LEVEL=debug
ENVIRONMENT=development

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Upload Configuration
UPLOAD_MAX_SIZE=10485760  # 10MB
UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,image/gif

# Rate Limiting (Development - disabled)
RATE_LIMIT_ENABLED=false
RATE_LIMIT_REQUESTS_PER_MINUTE=1000

# External Services (Development)
SMTP_HOST=localhost
SMTP_PORT=1025  # MailHog for development
SMTP_USER=
SMTP_PASSWORD=
EOF
```

### 5. VS Code Setup

#### Extensions

```json
// .vscode/extensions.json
{
    "recommendations": [
        "golang.go",
        "ms-vscode.vscode-json",
        "humao.rest-client",
        "ms-vscode.vscode-thunder-client",
        "bradlc.vscode-tailwindcss",
        "esbenp.prettier-vscode",
        "ms-vscode.vscode-yaml",
        "redhat.vscode-xml",
        "ms-vscode.vscode-docker"
    ]
}
```

#### Workspace Settings

```json
// .vscode/settings.json
{
    "go.useLanguageServer": true,
    "go.autocompleteUnimportedPackages": true,
    "go.gocodeAutoBuild": true,
    "go.installDependenciesWhenBuilding": true,
    "go.testFlags": ["-v"],
    "go.testTimeout": "60s",
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,128,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--fast"
    ],
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "files.exclude": {
        "**/.git": true,
        "**/.DS_Store": true,
        "**/node_modules": true,
        "**/vendor": true
    }
}
```

#### Tasks Configuration

```json
// .vscode/tasks.json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "go: build",
            "type": "shell",
            "command": "go",
            "args": ["build", "-o", "bin/main", "cmd/main.go"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": "$go"
        },
        {
            "label": "go: run",
            "type": "shell",
            "command": "go",
            "args": ["run", "cmd/main.go"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": "$go"
        },
        {
            "label": "go: test",
            "type": "shell",
            "command": "go",
            "args": ["test", "-v", "./..."],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": "$go"
        },
        {
            "label": "docker: build",
            "type": "shell",
            "command": "docker",
            "args": ["build", "-t", "ozgur-mutfak", "."],
            "group": "build"
        },
        {
            "label": "docker: run dev",
            "type": "shell",
            "command": "docker-compose",
            "args": ["-f", "docker-compose.dev.yml", "up", "--build"],
            "group": "build"
        }
    ]
}
```

#### Launch Configuration

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch main.go",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/main.go",
            "env": {
                "GIN_MODE": "debug",
                "DB_HOST": "localhost",
                "DB_PORT": "5432",
                "DB_NAME": "ozgur_mutfak_dev",
                "DB_USER": "ozgur_dev",
                "DB_PASSWORD": "dev_password"
            },
            "args": []
        },
        {
            "name": "Test Current File",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${fileDirname}"
        },
        {
            "name": "Test Current Package",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/${relativeFileDirname}"
        }
    ]
}
```

## 🔧 Development Tools

### 1. Go Tools Installation

```bash
# Essential Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/pressly/goose/v3/cmd/goose@latest

# Additional helpful tools
go install github.com/air-verse/air@latest  # Hot reload
go install github.com/golang/mock/mockgen@latest  # Mock generation
go install github.com/securecodewarrior/sast-scan@latest  # Security scanning
```

### 2. Air (Hot Reload) Setup

```bash
# Air configuration
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
args_bin = []
bin = "./tmp/main"
cmd = "go build -o ./tmp/main ./cmd/main.go"
delay = 1000
exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules"]
exclude_file = []
exclude_regex = ["_test.go"]
exclude_unchanged = false
follow_symlink = false
full_bin = ""
include_dir = []
include_ext = ["go", "tpl", "tmpl", "html"]
include_file = []
kill_delay = "0s"
log = "build-errors.log"
send_interrupt = false
stop_on_root = false

[color]
app = ""
build = "yellow"
main = "magenta"
runner = "green"
watcher = "cyan"

[log]
main_only = false
time = false

[misc]
clean_on_exit = false

[screen]
clear_on_rebuild = false
keep_scroll = true
EOF

# Hot reload ile çalıştır
air
```

### 3. Database Migration Tools

```bash
# Goose migration tool
go install github.com/pressly/goose/v3/cmd/goose@latest

# Migration oluştur
goose -dir migrations create initial_schema sql

# Migration çalıştır
goose -dir migrations postgres "user=ozgur_dev password=dev_password dbname=ozgur_mutfak_dev sslmode=disable" up

# Migration geri al
goose -dir migrations postgres "user=ozgur_dev password=dev_password dbname=ozgur_mutfak_dev sslmode=disable" down
```

### 4. API Documentation (Swagger)

```bash
# Swagger docs generate et
swag init -g cmd/main.go -o docs/

# Swagger UI'ı kontrol et
# http://localhost:8080/docs/index.html
```

## 🧪 Testing Setup

### 1. Test Database Setup

```bash
# Test database oluştur
sudo -u postgres psql -c "CREATE DATABASE ozgur_mutfak_test;"
sudo -u postgres psql -c "CREATE USER ozgur_test WITH PASSWORD 'test_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE ozgur_mutfak_test TO ozgur_test;"
```

### 2. Test Environment

```bash
# Test environment dosyası
cat > .env.test << EOF
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ozgur_mutfak_test
DB_USER=ozgur_test
DB_PASSWORD=test_password
DB_SSL_MODE=disable

JWT_SECRET=test-jwt-secret-for-testing-only
GIN_MODE=test
LOG_LEVEL=error
ENVIRONMENT=test
EOF
```

### 3. Test Runner Script

```bash
# scripts/run-tests.sh
#!/bin/bash

echo "Running Özgür Mutfak API Tests..."

# Set test environment
export $(cat .env.test | xargs)

# Clean test database
goose -dir migrations postgres "user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=disable" reset
goose -dir migrations postgres "user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=disable" up

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
go tool cover -func=coverage.out

echo "Tests completed. Coverage report: coverage.html"
```

```bash
# Script'i executable yap
chmod +x scripts/run-tests.sh

# Testleri çalıştır
./scripts/run-tests.sh
```

## 🚀 Running the Application

### 1. Direct Go Run

```bash
# Environment variables yükle
export $(cat .env.development | xargs)

# Uygulamayı çalıştır
go run cmd/main.go
```

### 2. Build and Run

```bash
# Build
go build -o bin/main cmd/main.go

# Run
./bin/main
```

### 3. Docker Development

```bash
# Docker ile development ortamını başlat
docker-compose -f docker-compose.dev.yml up --build

# Background'da çalıştır
docker-compose -f docker-compose.dev.yml up -d --build

# Logs'ları takip et
docker-compose -f docker-compose.dev.yml logs -f api
```

### 4. Air (Hot Reload)

```bash
# Hot reload ile geliştirme
air

# Custom config ile
air -c .air.toml
```

## 🔍 Development Workflow

### 1. Git Workflow

```bash
# Feature branch oluştur
git checkout -b feature/meal-management

# Değişiklikleri commit et
git add .
git commit -m "feat: add meal creation endpoint"

# Push to remote
git push origin feature/meal-management

# Main branch'e merge (PR sonrası)
git checkout main
git pull origin main
git branch -d feature/meal-management
```

### 2. Code Quality Checks

```bash
# Format code
go fmt ./...

# Imports'ları organize et
goimports -w .

# Lint check
golangci-lint run

# Tests
go test ./...

# Security scan
gosec ./...
```

### 3. API Testing

#### REST Client (VS Code Extension)

```http
### Get all meals
GET http://localhost:8080/api/v1/meals
Content-Type: application/json

### Create a new meal
POST http://localhost:8080/api/v1/meals
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
    "name": "Ev Yapımı Mantı",
    "description": "El açması hamur ile hazırlanan mantı",
    "price": 25.00,
    "category": "Ana Yemek"
}

### Variables
@auth_token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### cURL Examples

```bash
# Health check
curl -X GET http://localhost:8080/health

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# Get meals with auth
curl -X GET http://localhost:8080/api/v1/meals \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📊 Monitoring and Debugging

### 1. Logging

```go
// Development logging setup
log.SetLevel(log.DebugLevel)
log.SetFormatter(&log.JSONFormatter{})

// Request logging middleware
func LoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf(`{"time":"%s","status":%d,"latency":"%s","ip":"%s","method":"%s","path":"%s"}`,
            param.TimeStamp.Format(time.RFC3339),
            param.StatusCode,
            param.Latency,
            param.ClientIP,
            param.Method,
            param.Path,
        ) + "\n"
    })
}
```

### 2. Database Debugging

```bash
# Database connection test
go run scripts/test-db.go

# SQL query logging (development)
# Add to config: log_statement = 'all' in postgresql.conf
```

### 3. Performance Profiling

```go
import _ "net/http/pprof"

// Profiling endpoint'ini ekle
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling  
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 🛠️ Troubleshooting

### Common Issues

1. **Database Connection Error**
   ```bash
   # PostgreSQL servisini kontrol et
   sudo systemctl status postgresql
   
   # Connection test
   psql -h localhost -U ozgur_dev -d ozgur_mutfak_dev
   ```

2. **Port Already in Use**
   ```bash
   # Port'u kullanan process'i bul
   lsof -i :8080
   
   # Process'i kill et
   kill -9 <PID>
   ```

3. **Go Module Issues**
   ```bash
   # Module cache'i temizle
   go clean -modcache
   
   # Dependencies'leri yeniden yükle
   go mod download
   ```

4. **Permission Issues (Linux/macOS)**
   ```bash
   # Dosya permissions'ını düzelt
   chmod +x scripts/*.sh
   
   # Directory ownership
   sudo chown -R $USER:$USER .
   ```

### Debug Mode

```bash
# Debug mode ile çalıştır
GIN_MODE=debug LOG_LEVEL=debug go run cmd/main.go

# Verbose logging
VERBOSE=true go run cmd/main.go
```

## 📋 Development Checklist

### Initial Setup
- [ ] Go installed (version 1.21+)
- [ ] PostgreSQL installed and running
- [ ] Repository cloned
- [ ] Dependencies downloaded
- [ ] Database created
- [ ] Environment variables configured
- [ ] VS Code extensions installed

### Before Coding
- [ ] Feature branch created
- [ ] Database migrations up to date
- [ ] Tests passing
- [ ] Code formatted and linted

### Before Commit
- [ ] All tests passing
- [ ] Code coverage adequate (>80%)
- [ ] No linting errors
- [ ] Documentation updated
- [ ] API docs regenerated

### Before Push
- [ ] Rebase with main branch
- [ ] Integration tests passing
- [ ] No merge conflicts
- [ ] Commit messages clear

---

**Note**: Bu development guide, yeni geliştirici onboarding'i için güncel tutulmalıdır.
