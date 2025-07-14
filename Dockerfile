# Multi-stage build
FROM golang:1.21-alpine AS builder

# Çalışma dizini ayarla
WORKDIR /app

# Go mod dosyalarını kopyala
COPY go.mod go.sum ./

# Bağımlılıkları indir ve tidy yap
RUN go mod download
RUN go mod tidy

# Kaynak kodları kopyala
COPY . .

# go mod tidy tekrar çalıştır
RUN go mod tidy

# Eksik modülleri indir
RUN go get gopkg.in/yaml.v3
RUN go get github.com/gin-gonic/gin
RUN go get github.com/lib/pq
RUN go get github.com/golang-jwt/jwt/v5

# Swagger bağımlılıklarını ekle
RUN go get github.com/swaggo/swag/cmd/swag
RUN go get github.com/swaggo/gin-swagger
RUN go get github.com/swaggo/files

# Swagger CLI'yi kur
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Go-migrate CLI'yi kur (Go 1.21 ile uyumlu versiyon)
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2

# Swagger dokümantasyonu oluştur
RUN swag init -g cmd/main.go

# Binary oluştur
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Final stage
FROM alpine:latest

# SSL sertifikaları ekle
RUN apk --no-cache add ca-certificates

# Çalışma dizini
WORKDIR /root/

# Binary'yi kopyala
COPY --from=builder /app/main .

# Go binary'lerini kopyala (migrate için)
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Config dosyasını kopyala
COPY --from=builder /app/config ./config

# Migration dosyalarını kopyala
COPY --from=builder /app/migrations ./migrations

# Swagger docs'ları kopyala
COPY --from=builder /app/docs ./docs

# Port'u expose et
EXPOSE 8080

# Uygulamayı başlat
CMD ["./main"]
