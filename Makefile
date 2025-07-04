# E-Commerce Docker Makefile

.PHONY: help build up down logs clean restart

# Varsayılan hedef
help:
	@echo "E-Commerce Docker Commands:"
	@echo "  make build    - Docker image'larını oluştur"
	@echo "  make up       - Container'ları başlat"
	@echo "  make down     - Container'ları durdur"
	@echo "  make logs     - Logları göster"
	@echo "  make restart  - Container'ları yeniden başlat"
	@echo "  make clean    - Tüm container'ları ve volume'ları temizle"

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
