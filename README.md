# � Özgür Mutfak - Home-Cooked Meal Marketplace API

Modern, modüler ve scalable bir ev yemekleri platformu backend API'si. Docker ile tam entegre edilmiş, PostgreSQL veritabanı kullanarak geliştirilmiş professional bir home-cooked meal marketplace çözümü.

## 🚀 Özellikler

- ✅ **Modüler Mimari**: Clean Architecture prensiplerine uygun
- ✅ **JWT Authentication**: Güvenli kullanıcı kimlik doğrulama
- ✅ **Docker Support**: Tam Docker entegrasyonu
- ✅ **PostgreSQL**: Güvenilir veritabanı çözümü
- ✅ **RESTful API**: Standart HTTP endpoint'leri
- ✅ **Swagger Documentation**: API dokümantasyonu
- ✅ **Chef Management**: Şef yönetimi ve doğrulama
- ✅ **Meal Catalog**: Ev yemekleri kataloğu
- ✅ **Cart Management**: Sepet yönetimi
- ✅ **Order Processing**: Sipariş işleme sistemi
- ✅ **Review System**: Değerlendirme sistemi
- ✅ **Admin Dashboard**: Kapsamlı admin paneli

## 📁 Proje Yapısı

```
├── cmd/                    # Ana uygulama
├── internal/               # Uygulama kodu
│   ├── api/               # API handlers ve routing
│   ├── service/           # İş mantığı katmanı
│   ├── repository/        # Veri erişim katmanı
│   ├── model/             # Veri modelleri
│   └── auth/              # JWT authentication
├── config/                # Konfigürasyon
├── migrations/            # Veritabanı migration'ları
├── docs/                  # Dokümantasyon
├── tests/                 # Test dosyaları
├── scripts/               # Yardımcı scriptler
│   ├── docker/           # Docker scriptleri
│   └── windows/          # Windows scriptleri
├── api-docs/             # API dokümantasyonu
├── docker-compose.yml    # Docker servis tanımları
├── Dockerfile           # Container tanımı
└── README.md           # Bu dosya
```

## 🛠️ Kurulum

### Gereksinimler

- Docker & Docker Compose
- Git

### Hızlı Başlangıç

```bash
# Repository'yi klonlayın
git clone https://github.com/Yasin4261/food-delivery.git
cd food-delivery

# Docker servislerini başlatın
docker-compose up -d

# API'nin çalışıp çalışmadığını test edin
curl http://localhost:3001/api/v1/meals

# Swagger UI'yi ziyaret edin
# http://localhost:3001/swagger/index.html
```

### Windows Kullanıcıları için

```powershell
# Docker servislerini başlat
.\scripts\docker\docker-start.bat

# API'yi test et
.\scripts\windows\simple-test.ps1

# Logları görüntüle
.\scripts\docker\docker-logs.bat
```

## 🔧 Geliştirme

### Lokal Geliştirme

```bash
# Bağımlılıkları yükle
go mod download

# Uygulamayı çalıştır
go run cmd/main.go

# Veritabanı migration'ını çalıştır
docker exec -it ecommerce_db psql -U postgres -d ecommerce_db -f /migrations/001_initial_schema.sql
```

## 🧪 Test Etme

### Test Çalıştırma Seçenekleri

#### 1. Make ile Test Çalıştırma (Önerilen)
```bash
# Tüm testleri çalıştır
make test

# Sadece unit testleri
make test-unit

# Integration testleri
make test-integration

# Coverage raporu ile
make test-coverage

# Race condition testleri
make test-race

# Benchmark testleri
make test-bench

# Docker ile testleri çalıştır
make test-docker
```

#### 2. Go ile Direkt Test Çalıştırma
```bash
# Tüm testleri çalıştır
go test -v ./...

# Sadece model testleri
go test -v ./internal/model/...

# Coverage ile
go test -v ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Race condition testi
go test -race -v ./internal/service/...

# Benchmark testleri
go test -bench=. -benchmem ./internal/api/handler/...
```

#### 3. Script ile Test Çalıştırma
```bash
# Linux/Mac
./scripts/run-tests.sh

# Windows PowerShell
.\scripts\windows\run-tests.ps1 -Coverage -Race -Bench

# Docker ile
./scripts/docker/run-tests.sh
```

#### 4. CI/CD Test Çalıştırma
```bash
# CI için optimize edilmiş testler
make test-ci

# Test sonuçlarını temizle
make test-clean
```

### Test Kategorileri

| Test Türü | Açıklama | Dosya Yolu |
|-----------|----------|------------|
| **Model Tests** | Veri modellerinin JSON serialization testleri | `internal/model/*_test.go` |
| **Service Tests** | İş mantığı katmanı testleri | `internal/service/*_test.go` |
| **Handler Tests** | HTTP handler testleri | `internal/api/handler/*_test.go` |
| **Integration Tests** | End-to-end API testleri | `tests/integration_test.go` |

### Test Sonuçları

Test sonuçları `test-results/` klasöründe saklanır:
- `coverage.html` - Coverage raporu
- `*-test.log` - Test logları
- `benchmark.log` - Benchmark sonuçları
- `race-test.log` - Race condition test sonuçları

### Postman Collection Kullanma
```bash
# Postman collection'ını kullan
# tests/postman_collection.json dosyasını Postman'e import edin
# tests/postman_environment.json dosyasını environment olarak ekleyin
```

### HTTP Test Dosyası
```bash
# VSCode REST Client ile
# tests/api-test.http dosyasını VSCode'da açın
```

## 📊 API Endpoints

### 🔐 Authentication
- `POST /api/v1/auth/register` - Kullanıcı kaydı (customer/chef)
- `POST /api/v1/auth/login` - Kullanıcı girişi

### 🍽️ Meals (Yemekler)
- `GET /api/v1/meals` - Mevcut yemekleri listele
- `GET /api/v1/meals/:id` - Yemek detayı
- `POST /api/v1/meals` - Yeni yemek ekle (chef)
- `PUT /api/v1/meals/:id` - Yemek güncelle (chef)
- `DELETE /api/v1/meals/:id` - Yemek sil (chef)

### 👨‍🍳 Chefs (Şefler)
- `GET /api/v1/chefs` - Aktif şefleri listele
- `GET /api/v1/chefs/:id` - Şef profili ve yemekleri
- `POST /api/v1/chefs` - Şef profili oluştur
- `PUT /api/v1/chefs/:id` - Şef profili güncelle
- `GET /api/v1/chefs/:id/meals` - Şefin yemekleri

### 👤 User Profile
- `GET /api/v1/users/profile` - Profil bilgisi
- `PUT /api/v1/users/profile` - Profil güncelleme

### 🛒 Cart (Sepet)
- `GET /api/v1/cart` - Sepet görüntüleme
- `POST /api/v1/cart/add` - Sepete yemek ekleme
- `PUT /api/v1/cart/update/:id` - Sepet öğesi güncelleme
- `DELETE /api/v1/cart/remove/:id` - Sepetten yemek çıkarma

### 📦 Orders (Siparişler)
- `GET /api/v1/orders` - Siparişleri listele
- `POST /api/v1/orders` - Yeni sipariş oluştur
- `GET /api/v1/orders/:id` - Sipariş detayı
- `PUT /api/v1/orders/:id/status` - Sipariş durumu güncelle

### ⭐ Reviews (Değerlendirmeler)
- `GET /api/v1/meals/:id/reviews` - Yemek değerlendirmeleri
- `POST /api/v1/reviews` - Değerlendirme yap
- `GET /api/v1/chefs/:id/reviews` - Şef değerlendirmeleri

### 🔧 Admin
- `GET /api/v1/admin/dashboard` - Dashboard istatistikleri
- `GET /api/v1/admin/users` - Kullanıcı yönetimi
- `GET /api/v1/admin/chefs` - Şef yönetimi ve doğrulama
- `GET /api/v1/admin/meals` - Yemek yönetimi
- `GET /api/v1/admin/orders` - Sipariş yönetimi
- `PUT /api/v1/admin/chefs/:id/verify` - Şef doğrulama

## 🗄️ Veritabanı

### Tablolar
- `users` - Kullanıcı bilgileri (customer/chef)
- `chefs` - Şef profilleri ve iş bilgileri
- `meals` - Ev yemekleri kataloğu
- `carts` - Kullanıcı sepetleri
- `cart_items` - Sepet öğeleri
- `orders` - Siparişler
- `order_items` - Sipariş öğeleri
- `reviews` - Yemek ve şef değerlendirmeleri

### Veritabanı Yönetimi

```bash
# pgAdmin: http://localhost:8081
# Email: admin@admin.com
# Password: admin

# Adminer: http://localhost:8082
# Server: ecommerce_db
# Username: postgres
# Password: postgres123
# Database: ecommerce
```

## 📱 Swagger API Dokümantasyonu

API dokümantasyonuna Swagger UI üzerinden erişebilirsiniz:

**URL:** `http://localhost:3001/swagger/index.html`

Swagger dokümantasyonu otomatik olarak güncellenir ve tüm endpoint'leri interaktif olarak test edebilirsiniz.

## 🔍 Monitoring & Logging

### Docker Logs

```bash
# Tüm servislerin logları
docker-compose logs -f

# Sadece API logları
docker logs -f ecommerce_api

# Sadece DB logları
docker logs -f ecommerce_db
```

### Health Check

```bash
# API sağlık kontrolü
curl http://localhost:3001/api/v1/meals

# Swagger UI kontrolü  
curl http://localhost:3001/swagger/index.html

# Veritabanı bağlantı kontrolü
docker exec ecommerce_db pg_isready -U postgres
```

## 📚 Dokümantasyon

- [API Test Guide](docs/API_TEST_GUIDE.md) - API test rehberi
- [API Test Guide (TR)](docs/API_TEST_GUIDE_TR.md) - API test rehberi (Türkçe)
- [Code Structure](docs/CODE_STRUCTURE.md) - Kod yapısı ve mimari
- [Database Schema](docs/DATABASE_SCHEMA.md) - Veritabanı şeması
- [Migration Report](docs/MODULAR_MIGRATION_REPORT.md) - Modüler yapı geçiş raporu

## 🧪 Test Dosyaları

- `tests/api-test.http` - HTTP test dosyası
- `tests/admin-test.http` - Admin endpoint test dosyası
- `postman_collection.json` - Postman koleksiyonu
- `postman_environment.json` - Postman ortam değişkenleri

## 🐳 Docker Servisler

- **ecommerce_api** - Ana API servisi (Port: 3001)
- **ecommerce_db** - PostgreSQL veritabanı (Port: 5432)  
- **ecommerce_pgadmin** - pgAdmin web arayüzü (Port: 8081)
- **ecommerce_adminer** - Adminer web arayüzü (Port: 8082)

## 🏗️ Teknolojiler

- **Go** - Programlama dili
- **Gin** - HTTP web framework  
- **PostgreSQL** - Veritabanı
- **Docker** - Containerization
- **JWT** - Authentication
- **bcrypt** - Password hashing
- **Swagger** - API dokümantasyonu
- **pgAdmin** - Veritabanı yönetimi
- **Adminer** - Hafif veritabanı arayüzü

## 🤝 Katkıda Bulunma

1. Fork edin
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'Add amazing feature'`)
4. Branch'inizi push edin (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## 📄 Lisans

Bu proje MIT lisansı altında lisanslanmıştır.

## 📞 İletişim

- **Proje Sahibi**: Yasin
- **GitHub**: [Yasin4261](https://github.com/Yasin4261)
- **Repository**: [food-delivery](https://github.com/Yasin4261/food-delivery)

---

## 🎯 Geliştirme Roadmap

- [x] **Modüler Mimari** - Clean Architecture yapısı
- [x] **Docker Integration** - Tam Docker desteği
- [x] **JWT Authentication** - Güvenli kimlik doğrulama
- [x] **PostgreSQL Setup** - Veritabanı entegrasyonu
- [ ] **Unit Testing** - Kapsamlı test coverage
- [ ] **API Documentation** - Swagger/OpenAPI entegrasyonu
- [ ] **Performance Optimization** - Caching ve optimizasyon
- [ ] **Mobile API** - Mobil uygulama desteği

---

**🍳 Özgür Mutfak - Home-Cooked Meal Marketplace Platform**