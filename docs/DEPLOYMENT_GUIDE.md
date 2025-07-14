# 🚀 Deployment Guide

## Overview

Bu kılavuz, Özgür Mutfak API'sının farklı ortamlarda nasıl deploy edileceğini açıklar.

## 📋 Prerequisites

### System Requirements

- **Server**: Linux (Ubuntu 20.04+ önerilen)
- **Memory**: Minimum 2GB RAM (Production için 4GB+)
- **Storage**: Minimum 20GB disk alanı
- **Network**: HTTP/HTTPS portları (80/443)

### Required Software

1. **Docker & Docker Compose**
   ```bash
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo curl -L "https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

2. **Git**
   ```bash
   sudo apt update
   sudo apt install git
   ```

3. **SSL Certificate** (Production için)
   - Let's Encrypt (ücretsiz)
   - CloudFlare SSL
   - Kendi sertifikanız

## 🐳 Docker Deployment

### 1. Repository Clone

```bash
git clone https://github.com/yourusername/ozgur-mutfak.git
cd ozgur-mutfak
```

### 2. Environment Configuration

#### Development Environment

```bash
# .env.development dosyası oluştur
cat > .env.development << EOF
# Database
DB_HOST=db
DB_PORT=5432
DB_NAME=ozgur_mutfak_dev
DB_USER=postgres
DB_PASSWORD=your_dev_password

# API
API_PORT=8080
API_HOST=localhost
JWT_SECRET=your_jwt_secret_dev

# Other
GIN_MODE=debug
LOG_LEVEL=debug
EOF
```

#### Production Environment

```bash
# .env.production dosyası oluştur
cat > .env.production << EOF
# Database
DB_HOST=db
DB_PORT=5432
DB_NAME=ozgur_mutfak_prod
DB_USER=postgres
DB_PASSWORD=your_strong_production_password

# API
API_PORT=8080
API_HOST=0.0.0.0
JWT_SECRET=your_very_strong_jwt_secret

# Other
GIN_MODE=release
LOG_LEVEL=info

# SSL (if using)
SSL_CERT_PATH=/etc/ssl/certs/domain.crt
SSL_KEY_PATH=/etc/ssl/private/domain.key
EOF
```

### 3. Docker Compose Setup

#### Development (docker-compose.dev.yml)

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3001:8080"
    environment:
      - GIN_MODE=debug
    env_file:
      - .env.development
    volumes:
      - .:/app
      - /app/vendor
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ozgur_mutfak_dev
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_dev_password
    ports:
      - "5433:5432"
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    restart: unless-stopped

  adminer:
    image: adminer
    ports:
      - "8081:8080"
    depends_on:
      - db
    restart: unless-stopped

volumes:
  postgres_dev_data:
```

#### Production (docker-compose.prod.yml)

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
    env_file:
      - .env.production
    depends_on:
      - db
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ozgur_mutfak_prod
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_strong_production_password
    volumes:
      - postgres_prod_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
      - ./backups:/backups
    restart: always
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - api
    restart: always

volumes:
  postgres_prod_data:
```

### 4. Nginx Configuration

```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream api {
        server api:8080;
    }

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name yourdomain.com www.yourdomain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name yourdomain.com www.yourdomain.com;

        ssl_certificate /etc/ssl/certs/domain.crt;
        ssl_certificate_key /etc/ssl/certs/domain.key;
        
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

        # API proxy
        location /api/ {
            proxy_pass http://api/api/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Health check
        location /health {
            proxy_pass http://api/health;
        }

        # Swagger documentation
        location /docs/ {
            proxy_pass http://api/docs/;
        }
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    
    server {
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://api/api/;
        }
    }
}
```

### 5. Build and Deploy

#### Development
```bash
# Development ortamında çalıştır
docker-compose -f docker-compose.dev.yml up --build

# Background'da çalıştır
docker-compose -f docker-compose.dev.yml up -d --build
```

#### Production
```bash
# Production build
docker-compose -f docker-compose.prod.yml up --build -d

# Logs kontrol et
docker-compose -f docker-compose.prod.yml logs -f api
```

## ☁️ Cloud Deployment

### AWS EC2 Deployment

#### 1. EC2 Instance Setup

```bash
# Amazon Linux 2 AMI
sudo yum update -y
sudo yum install -y docker git

# Docker servisini başlat
sudo service docker start
sudo usermod -a -G docker ec2-user

# Docker Compose yükle
sudo curl -L https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 2. Security Groups

```bash
# HTTP/HTTPS trafiği için port açma
# AWS Console'dan:
# - Port 80 (HTTP)
# - Port 443 (HTTPS) 
# - Port 22 (SSH)
```

#### 3. RDS PostgreSQL (Önerilen)

```bash
# AWS RDS PostgreSQL instance oluştur
# .env.production dosyasını güncelle:
DB_HOST=your-rds-endpoint.region.rds.amazonaws.com
DB_PORT=5432
DB_NAME=ozgur_mutfak
DB_USER=postgres
DB_PASSWORD=your_rds_password
```

### Google Cloud Platform

#### 1. Cloud Run Deployment

```dockerfile
# Dockerfile.cloudrun
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./main"]
```

```bash
# Build ve deploy
gcloud builds submit --tag gcr.io/PROJECT_ID/ozgur-mutfak
gcloud run deploy ozgur-mutfak --image gcr.io/PROJECT_ID/ozgur-mutfak --platform managed --region us-central1 --allow-unauthenticated
```

### DigitalOcean Droplet

#### 1. Droplet Setup

```bash
# Ubuntu 20.04 Droplet oluştur
# SSH ile bağlan
ssh root@your_droplet_ip

# Sistem güncelle
apt update && apt upgrade -y

# Docker yükle
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Firewall ayarla
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

## 🔧 Environment Management

### Configuration Files

#### config.yaml (Production)

```yaml
server:
  host: "0.0.0.0"
  port: "8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: "${DB_HOST}"
  port: "${DB_PORT}"
  name: "${DB_NAME}"
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s

auth:
  jwt_secret: "${JWT_SECRET}"
  jwt_expiry: 24h
  bcrypt_cost: 12

logging:
  level: "${LOG_LEVEL}"
  format: "json"
  file: "/var/log/ozgur-mutfak/app.log"

cors:
  allowed_origins:
    - "https://yourdomain.com"
    - "https://www.yourdomain.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allowed_headers:
    - "Content-Type"
    - "Authorization"

rate_limiting:
  enabled: true
  requests_per_minute: 60
  burst: 100
```

### Environment Variables

```bash
# Production environment variables
export DB_HOST="your-db-host"
export DB_PORT="5432"
export DB_NAME="ozgur_mutfak"
export DB_USER="postgres"
export DB_PASSWORD="your-secure-password"
export JWT_SECRET="your-jwt-secret-min-32-chars"
export GIN_MODE="release"
export LOG_LEVEL="info"
export API_BASE_URL="https://api.yourdomain.com"
```

## 📊 Monitoring and Logging

### Health Checks

```go
// Health check endpoint implementation
func HealthCheck(c *gin.Context) {
    health := struct {
        Status    string    `json:"status"`
        Timestamp time.Time `json:"timestamp"`
        Version   string    `json:"version"`
        Database  string    `json:"database"`
    }{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   "1.0.0",
    }

    // Database connection check
    if err := db.Ping(); err != nil {
        health.Status = "unhealthy"
        health.Database = "disconnected"
        c.JSON(503, health)
        return
    }
    
    health.Database = "connected"
    c.JSON(200, health)
}
```

### Log Management

```bash
# Docker logs
docker-compose logs -f api

# Log rotation setup
sudo nano /etc/logrotate.d/ozgur-mutfak
```

```bash
# /etc/logrotate.d/ozgur-mutfak
/var/log/ozgur-mutfak/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0644 app app
    postrotate
        docker-compose -f /path/to/docker-compose.prod.yml restart api
    endscript
}
```

### Prometheus Monitoring (İsteğe bağlı)

```yaml
# monitoring/docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  grafana_data:
```

## 🔄 Database Migrations

### Production Migration Strategy

```bash
# Migration script
#!/bin/bash
# migrate-prod.sh

echo "Starting database migration..."

# Backup database
docker exec postgres_container pg_dump -U postgres ozgur_mutfak > backup_$(date +%Y%m%d_%H%M%S).sql

# Run migrations
docker exec api_container /app/migrate up

echo "Migration completed successfully"
```

### Rollback Strategy

```bash
# Rollback script
#!/bin/bash
# rollback-prod.sh

if [ -z "$1" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

echo "Rolling back to backup: $1"

# Stop API
docker-compose -f docker-compose.prod.yml stop api

# Restore database
docker exec postgres_container psql -U postgres -c "DROP DATABASE ozgur_mutfak;"
docker exec postgres_container psql -U postgres -c "CREATE DATABASE ozgur_mutfak;"
docker exec -i postgres_container psql -U postgres ozgur_mutfak < $1

# Start API
docker-compose -f docker-compose.prod.yml start api

echo "Rollback completed"
```

## 🔒 Security Hardening

### SSL/TLS Setup

```bash
# Let's Encrypt ile ücretsiz SSL
sudo apt install certbot

# Certificate al
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# Auto-renewal setup
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Firewall Configuration

```bash
# UFW firewall setup
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### Database Security

```sql
-- Production database security
-- Create application-specific user
CREATE USER ozgur_app WITH PASSWORD 'secure_app_password';

-- Grant minimal permissions
GRANT CONNECT ON DATABASE ozgur_mutfak TO ozgur_app;
GRANT USAGE ON SCHEMA public TO ozgur_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ozgur_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ozgur_app;

-- Remove postgres user from production access
-- Update connection string to use ozgur_app user
```

## 📋 Deployment Checklist

### Pre-deployment

- [ ] Environment variables configured
- [ ] SSL certificates ready
- [ ] Database backups taken
- [ ] Performance testing completed
- [ ] Security scan passed
- [ ] Docker images built and tested

### Deployment

- [ ] Application deployed
- [ ] Database migrations run
- [ ] Health checks passing
- [ ] SSL/HTTPS working
- [ ] Logs configured
- [ ] Monitoring active

### Post-deployment

- [ ] Functionality testing
- [ ] Performance monitoring
- [ ] Error tracking active
- [ ] Backup verification
- [ ] Team notification
- [ ] Documentation updated

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database container
   docker-compose logs db
   
   # Test connection
   docker exec -it postgres_container psql -U postgres -d ozgur_mutfak
   ```

2. **API Container Not Starting**
   ```bash
   # Check logs
   docker-compose logs api
   
   # Check environment variables
   docker exec api_container env
   ```

3. **SSL Certificate Issues**
   ```bash
   # Check certificate validity
   openssl x509 -in /path/to/cert.crt -text -noout
   
   # Test SSL
   curl -I https://yourdomain.com
   ```

4. **Performance Issues**
   ```bash
   # Monitor resource usage
   docker stats
   
   # Database performance
   docker exec postgres_container psql -U postgres -c "SELECT * FROM pg_stat_activity;"
   ```

---

**Note**: Bu deployment guide production ortamında kullanılmadan önce test ortamında doğrulanmalıdır.
