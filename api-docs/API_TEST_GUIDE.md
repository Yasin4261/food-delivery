# 📋 Özgür Mutfak E-Commerce API - Postman Test Guide

Bu dokümantasyon, Özgür Mutfak e-ticaret API'sinin Postman ve HTTP client'lar ile nasıl test edileceğini detaylı olarak açıklar.

## 🚀 Hızlı Başlangıç

### Gereksinimler
- ✅ Docker ve Docker Compose kurulu olmalı
- ✅ Postman uygulaması (opsiyonel)
- ✅ VS Code + REST Client extension (opsiyonel)

### API'yi Çalıştırma
```bash
# Projeyi başlat
docker-compose up -d

# Servis durumunu kontrol et
docker ps

# API sağlık kontrolü
curl http://localhost:3001/api/v1/products
```

## 📁 Dosya Yapısı

- `postman_collection.json` - Postman collection dosyası
- `postman_environment.json` - Postman environment değişkenleri
- `api-test.http` - VS Code REST Client test dosyası

## 🔧 Postman Kurulumu

### 1. Collection ve Environment İmport Etme

1. **Postman'i aç**
2. **Import** butonuna tıkla
3. **File** sekmesini seç
4. Şu dosyaları sırayla import et:
   - `postman_collection.json`
   - `postman_environment.json`

### 2. Environment Aktivasyonu

1. Sağ üst köşedeki **Environment dropdown**'dan seç
2. **"Özgür Mutfak Environment"**'ı seç
3. Environment değişkenlerini kontrol et:
   - `base_url`: http://localhost:3001
   - `auth_token`: (otomatik doldurulacak)

## 🔐 Authentication Endpoint'leri

### 📝 1. REGISTER - Kullanıcı Kaydı

**Endpoint:** `POST /api/v1/auth/register`

**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "email": "john.doe@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Başarılı Response (200):**
```json
{
  "message": "Hesap başarıyla oluşturuldu",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "customer"
    }
  }
}
```

**Hata Response (400) - Duplicate Email:**
```json
{
  "error": "bu email adresi zaten kullanımda"
}
```

### 🔑 2. LOGIN - Giriş

**Endpoint:** `POST /api/v1/auth/login`

**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "email": "john.doe@example.com",
  "password": "password123"
}
```

**Başarılı Response (200):**
```json
{
  "message": "Başarıyla giriş yapıldı",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "customer"
    }
  }
}
```

**Hata Response (401) - Yanlış Kimlik:**
```json
{
  "error": "email veya şifre hatalı"
}
```

### 🚪 3. LOGOUT - Çıkış

**Endpoint:** `POST /api/v1/auth/logout`

**Headers:**
```
Authorization: Bearer YOUR_JWT_TOKEN_HERE
Content-Type: application/json
```

**Body:** Boş

**Başarılı Response (200):**
```json
{
  "message": "Başarıyla çıkış yapıldı"
}
```

**Hata Response (401) - Geçersiz Token:**
```json
{
  "error": "Geçersiz token"
}
```

## 🛍️ Products Endpoint'leri

### 📦 1. GET PRODUCTS - Ürünleri Listele

**Endpoint:** `GET /api/v1/products`

**Headers:** Yok (Authentication opsiyonel)

**Body:** Yok

**Response (200):**
```json
{
  "message": "Get products endpoint - henüz implement edilmedi"
}
```

## 🔄 Test Senaryoları

### ✅ Tam Test Sırası (Postman)

1. **Health Check**
   - `GET /api/v1/products` → 200 OK

2. **Kullanıcı Kaydı**
   - `POST /api/v1/auth/register` → 200/201
   - Token otomatik kaydedilir

3. **Duplicate Email Testi**
   - Aynı email ile register → 400 Error

4. **Login Test**
   - `POST /api/v1/auth/login` → 200
   - Yeni token otomatik kaydedilir

5. **Yanlış Şifre Testi**
   - Yanlış password ile login → 401 Error

6. **Logout Test**
   - `POST /api/v1/auth/logout` → 200
   - Token otomatik temizlenir

### 🧪 Validation Test Cases

#### Register Endpoint:
- ✅ **Başarılı kayıt:** Tüm alanlar dolu → 200/201
- ✅ **Email duplicate:** Mevcut email → 400
- ❌ **Eksik alan:** first_name eksik → 400
- ❌ **Geçersiz email:** "invalid-email" → 400
- ❌ **Kısa şifre:** "123" → 400

#### Login Endpoint:
- ✅ **Başarılı giriş:** Doğru kimlik → 200
- ✅ **Yanlış şifre:** → 401
- ✅ **Olmayan email:** → 401
- ❌ **Eksik alan:** password eksik → 400

#### Logout Endpoint:
- ✅ **Başarılı çıkış:** Geçerli token → 200
- ✅ **Geçersiz token:** → 401
- ✅ **Eksik token:** Authorization header yok → 401

## 🔧 VS Code REST Client Kullanımı

### Kurulum
1. VS Code'da **REST Client** extension'ını yükle
2. `api-test.http` dosyasını aç
3. Her `###` üstündeki **"Send Request"** linkine tıkla

### Token Yönetimi
1. Register/Login response'undan token'ı kopyala
2. `YOUR_JWT_TOKEN_HERE` yazan yerlere yapıştır
3. Logout test'ini çalıştır

## 💡 İpuçları

### Postman Scripts
Collection'da otomatik scripts var:
- ✅ **Register/Login:** Token otomatik kaydedilir
- ✅ **Response validation:** Status code kontrolü
- ✅ **Logout:** Token otomatik temizlenir

### Environment Variables
```
base_url = http://localhost:3001
auth_token = (otomatik set ediliyor)
user_email = test@example.com
user_password = password123
```

### Hata Giderme

#### API bağlanamıyor:
```bash
# Container durumunu kontrol et
docker ps

# API loglarını kontrol et
docker-compose logs api

# Portu kontrol et
curl http://localhost:3001/api/v1/products
```

#### Database hataları:
```bash
# Database container'ını kontrol et
docker exec -it ecommerce_db psql -U postgres -d ecommerce_db -c "\dt"

# Migration'ı kontrol et
docker-compose logs db
```

## 🚀 Gelecek Özellikler

### Planlanacak Endpoint'ler:
- `POST /api/v1/products` - Ürün ekleme (Admin)
- `PUT /api/v1/products/{id}` - Ürün güncelleme (Admin)
- `DELETE /api/v1/products/{id}` - Ürün silme (Admin)
- `GET /api/v1/categories` - Kategoriler
- `POST /api/v1/cart/add` - Sepete ekleme
- `GET /api/v1/cart` - Sepet görüntüleme
- `POST /api/v1/orders` - Sipariş oluşturma

### Authentication Geliştirmeleri:
- Refresh token mekanizması
- Role-based authorization (admin/customer)
- Password reset functionality
- Email verification

## 🆘 Destek

Test sırasında sorun yaşarsanız:
1. `docker-compose logs` ile servislerin durumunu kontrol edin
2. Environment değişkenlerinin doğru set edildiğini kontrol edin
3. API endpoint'lerinin aktif olduğunu `/products` ile test edin

---

📞 **İletişim:** API ile ilgili sorularınız için geliştirme ekibi ile iletişime geçin.
🔄 **Update:** Bu dokümantasyon API geliştirmesi ile birlikte güncellenecektir.
