# E-Commerce API Documentation

Bu dokümantasyon E-Commerce API'sinin tüm endpoint'lerini ve kullanım örneklerini içerir.

## Base URL

- **Development**: `http://localhost:8080/api/v1`
- **Docker**: `http://localhost:8080/api/v1`

## Authentication

API, JWT (JSON Web Token) tabanlı authentication kullanır. Korumalı endpoint'lere erişim için Authorization header'ında token göndermeniz gerekir:

```
Authorization: Bearer <your-jwt-token>
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - Kullanıcı girişi
- `POST /api/v1/auth/register` - Kullanıcı kaydı

### Products
- `GET /api/v1/products` - Tüm ürünleri listele
- `GET /api/v1/products/:id` - Ürün detayı

### User Profile (Auth Required)
- `GET /api/v1/profile` - Kullanıcı profili
- `PUT /api/v1/profile` - Profil güncelle

### Cart (Auth Required)
- `GET /api/v1/cart` - Sepeti görüntüle
- `POST /api/v1/cart/items` - Sepete ürün ekle
- `DELETE /api/v1/cart/items/:id` - Sepetten ürün sil

### Orders (Auth Required)
- `GET /api/v1/orders` - Siparişleri listele
- `POST /api/v1/orders` - Yeni sipariş oluştur
- `GET /api/v1/orders/:id` - Sipariş detayı

### Admin Endpoints (Admin Auth Required)
- `GET /api/v1/admin/products` - Tüm ürünleri yönet
- `POST /api/v1/admin/products` - Yeni ürün oluştur
- `PUT /api/v1/admin/products/:id` - Ürün güncelle
- `DELETE /api/v1/admin/products/:id` - Ürün sil
- `GET /api/v1/admin/orders` - Tüm siparişleri görüntüle
- `PUT /api/v1/admin/orders/:id/status` - Sipariş durumu güncelle

## Authentication

API'ye erişim için JWT token kullanılır. Token, `Authorization` header'ında `Bearer <token>` formatında gönderilmelidir.

## Response Format

Tüm API yanıtları JSON formatındadır:

```json
{
  "success": true,
  "data": {},
  "message": "İşlem başarılı"
}
```

Hata durumunda:

```json
{
  "success": false,
  "error": "Hata mesajı",
  "code": "ERROR_CODE"
}
```
