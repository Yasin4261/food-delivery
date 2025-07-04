# 🍽️ Özgür Mutfak - Food Delivery API

Modern, modüler ve scalable bir yemek teslimat backend API'si. Docker ile tam entegre edilmiş, PostgreSQL veritabanı kullanarak geliştirilmiş professional bir e-commerce çözümü.

## 🚀 Özellikler

- ✅ **Modüler Mimari**: Clean Architecture prensiplerine uygun
- ✅ **JWT Authentication**: Güvenli kullanıcı kimlik doğrulama
- ✅ **Docker Support**: Tam Docker entegrasyonu
- ✅ **PostgreSQL**: Güvenilir veritabanı çözümü
- ✅ **RESTful API**: Standart HTTP endpoint'leri
- ✅ **Admin Panel**: Admin yönetim arayüzü
- ✅ **Cart Management**: Sepet yönetimi
- ✅ **Order Processing**: Sipariş işleme sistemi
- ✅ **Product Catalog**: Ürün kataloğu yönetimi

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
curl http://localhost:8080/api/v1/products
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

### Test Etme

```bash
# Unit testleri çalıştır
go test ./...

# API testlerini çalıştır
.\scripts\windows\test-api.ps1

# Postman collection'ını kullan
# tests/postman_collection.json dosyasını Postman'e import edin
```

## 📊 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Kullanıcı kaydı
- `POST /api/v1/auth/login` - Kullanıcı girişi
- `POST /api/v1/auth/logout` - Kullanıcı çıkışı

### Products
- `GET /api/v1/products` - Ürünleri listele
- `GET /api/v1/products/:id` - Ürün detayı

### User Profile
- `GET /api/v1/profile` - Profil bilgisi
- `PUT /api/v1/profile` - Profil güncelleme

### Cart (Sepet)
- `GET /api/v1/cart` - Sepet görüntüleme
- `POST /api/v1/cart/items` - Sepete ürün ekleme
- `DELETE /api/v1/cart/items/:id` - Sepetten ürün çıkarma

### Orders (Siparişler)
- `GET /api/v1/orders` - Siparişleri listele
- `POST /api/v1/orders` - Yeni sipariş oluştur
- `GET /api/v1/orders/:id` - Sipariş detayı

### Admin
- `GET /api/v1/admin/products` - Ürün yönetimi
- `POST /api/v1/admin/products` - Ürün oluşturma
- `PUT /api/v1/admin/products/:id` - Ürün güncelleme
- `DELETE /api/v1/admin/products/:id` - Ürün silme
- `GET /api/v1/admin/orders` - Sipariş yönetimi
- `PUT /api/v1/admin/orders/:id/status` - Sipariş durumu güncelleme

## 🗄️ Veritabanı

### Tablolar
- `users` - Kullanıcı bilgileri
- `products` - Ürün kataloğu
- `categories` - Ürün kategorileri
- `carts` - Kullanıcı sepetleri
- `cart_items` - Sepet öğeleri
- `orders` - Siparişler
- `order_items` - Sipariş öğeleri

### Veritabanı Yönetimi

```bash
# pgAdmin: http://localhost:5050
# Email: admin@admin.com
# Password: admin

# Adminer: http://localhost:8081
# Server: ecommerce_db
# Username: postgres
# Password: postgres123
# Database: ecommerce_db
```

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
curl http://localhost:8080/api/v1/products

# Veritabanı bağlantı kontrolü
docker exec ecommerce_db pg_isready -U postgres
```

## 📚 Dokümantasyon

- [API Test Guide](api-docs/API_TEST_GUIDE.md) - API test rehberi
- [Code Structure](docs/CODE_STRUCTURE.md) - Kod yapısı ve mimari
- [Database Schema](docs/DATABASE_SCHEMA.md) - Veritabanı şeması
- [Migration Report](docs/MODULAR_MIGRATION_REPORT.md) - Modüler yapı geçiş raporu

## 🧪 Test Dosyaları

- `tests/api-test.http` - HTTP test dosyası
- `tests/postman_collection.json` - Postman koleksiyonu
- `tests/postman_environment.json` - Postman ortam değişkenleri

## 🐳 Docker Servisler

- **ecommerce_api** - Ana API servisi (Port: 8080)
- **ecommerce_db** - PostgreSQL veritabanı (Port: 5432)
- **ecommerce_pgadmin** - pgAdmin web arayüzü (Port: 5050)
- **ecommerce_adminer** - Adminer web arayüzü (Port: 8081)

## 🏗️ Teknolojiler

- **Go** - Programlama dili
- **Gin** - HTTP web framework
- **PostgreSQL** - Veritabanı
- **Docker** - Containerization
- **JWT** - Authentication
- **bcrypt** - Password hashing

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

**🚀 Özgür Mutfak - Modern Food Delivery Platform**