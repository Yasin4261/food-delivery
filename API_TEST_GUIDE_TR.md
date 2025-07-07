# Ev Yemekleri API Test Rehberi

Bu rehber, ev yemekleri pazaryeri platformu API'sini test etmek için gerekli tüm bilgileri içerir.

## Kurulum

### 1. Postman Collection'ı İçe Aktar
1. Postman'ı açın
2. `Import` butonuna tıklayın
3. `postman_collection.json` dosyasını seçin
4. Collection başarıyla içe aktarılacaktır

### 2. Environment'ı İçe Aktar
1. Postman'da `Environments` sekmesine gidin
2. `Import` butonuna tıklayın
3. `postman_environment.json` dosyasını seçin
4. Environment'ı aktif hale getirin

### 3. Docker Container'ları Çalıştır
```bash
docker-compose up -d
```

## API Endpoint'leri

### Kimlik Doğrulama (Authentication)
- **POST** `/api/v1/auth/register` - Yeni kullanıcı kaydı
- **POST** `/api/v1/auth/login` - Kullanıcı girişi (token alır)
- **POST** `/api/v1/auth/logout` - Kullanıcı çıkışı (token gerekli)

### Kullanıcı Profili (User Profile)
- **GET** `/api/v1/profile` - Profil bilgilerini getir
- **PUT** `/api/v1/profile` - Profil bilgilerini güncelle

### Yemekler (Meals) - Herkese Açık
- **GET** `/api/v1/meals` - Tüm yemekleri listele
- **GET** `/api/v1/meals/:id` - Belirli bir yemeği getir

### Şefler (Chefs) - Herkese Açık
- **GET** `/api/v1/chefs` - Tüm şefleri listele
- **GET** `/api/v1/chefs/:id` - Belirli bir şefi getir
- **GET** `/api/v1/chefs/:id/meals` - Şefin yemeklerini getir

### Sepet Yönetimi (Cart Management)
- **GET** `/api/v1/cart` - Sepeti getir
- **POST** `/api/v1/cart/items` - Sepete ürün ekle
- **DELETE** `/api/v1/cart/items/:id` - Sepetten ürün çıkar

### Sipariş Yönetimi (Order Management)
- **GET** `/api/v1/orders` - Siparişleri listele
- **POST** `/api/v1/orders` - Yeni sipariş oluştur
- **GET** `/api/v1/orders/:id` - Belirli bir siparişi getir

### Şef İşlemleri (Chef Operations)
- **GET** `/api/v1/chef/profile` - Şef profili getir
- **POST** `/api/v1/chef/profile` - Şef profili oluştur
- **PUT** `/api/v1/chef/profile` - Şef profili güncelle
- **GET** `/api/v1/chef/meals` - Şefin yemeklerini getir
- **POST** `/api/v1/chef/meals` - Yeni yemek oluştur
- **PUT** `/api/v1/chef/meals/:id` - Yemek güncelle
- **DELETE** `/api/v1/chef/meals/:id` - Yemek sil
- **PUT** `/api/v1/chef/meals/:id/toggle` - Yemek durumunu değiştir
- **GET** `/api/v1/chef/orders` - Şef siparişleri getir
- **PUT** `/api/v1/chef/orders/:id/status` - Sipariş durumu güncelle

### Admin İşlemleri (Admin Operations)
- **GET** `/api/v1/admin/dashboard` - Admin paneli özet
- **GET** `/api/v1/admin/users` - Tüm kullanıcıları listele
- **GET** `/api/v1/admin/users/:id` - Belirli kullanıcı getir
- **GET** `/api/v1/admin/chefs` - Tüm şefleri listele
- **GET** `/api/v1/admin/chefs/pending` - Bekleyen şefleri listele
- **PUT** `/api/v1/admin/chefs/:id/verify` - Şef doğrula
- **GET** `/api/v1/admin/orders` - Tüm siparişleri listele
- **PUT** `/api/v1/admin/orders/:id/status` - Sipariş durumu güncelle
- **GET** `/api/v1/admin/meals` - Tüm yemekleri listele
- **PUT** `/api/v1/admin/meals/:id/approve` - Yemek onayla
- **DELETE** `/api/v1/admin/meals/:id` - Yemek sil

## Test Senaryoları

### 1. Temel Kullanıcı Akışı
1. Yeni kullanıcı kayıt ol
2. Giriş yap (token al)
3. Profil bilgilerini görüntüle
4. Yemekleri listele
5. Şefleri listele
6. Sepete ürün ekle
7. Sipariş oluştur

### 2. Şef Akışı
1. Kullanıcı olarak kayıt ol
2. Giriş yap
3. Şef profili oluştur
4. Yemek ekle
5. Yemek durumunu güncelle
6. Siparişleri görüntüle
7. Sipariş durumu güncelle

### 3. Admin Akışı
1. Admin olarak giriş yap
2. Dashboard'u görüntüle
3. Kullanıcıları listele
4. Şefleri onayla
5. Yemekleri onayla
6. Siparişleri yönet

## Örnek Veri Formatları

### Kullanıcı Kaydı
```json
{
    "name": "Ahmet Yılmaz",
    "email": "ahmet@example.com",
    "password": "password123",
    "phone": "+905551234567",
    "address": "İstanbul, Turkey"
}
```

### Şef Profili
```json
{
    "bio": "Türk mutfağı uzmanı ev aşçısı",
    "specialties": "Türk, Akdeniz, Vejetaryen",
    "experience_years": 5,
    "kitchen_images": ["mutfak1.jpg", "mutfak2.jpg"],
    "certifications": ["Gıda Güvenliği Sertifikası"]
}
```

### Yemek Oluşturma
```json
{
    "name": "Ev Yapımı Döner Kebap",
    "description": "Taze malzemelerle hazırlanmış otantik Türk döner kebabı",
    "price": 25.50,
    "category": "Türk",
    "ingredients": ["Kuzu eti", "Soğan", "Domates", "Marul", "Yoğurt"],
    "allergens": ["Süt ürünleri"],
    "portion_size": "1 porsiyon",
    "preparation_time": 30,
    "spice_level": "Orta",
    "images": ["doner1.jpg", "doner2.jpg"]
}
```

### Sipariş Oluşturma
```json
{
    "delivery_address": "İstanbul, Turkey",
    "payment_method": "credit_card",
    "notes": "Lütfen akşam 6'dan sonra teslim edin"
}
```

## Durum Kodları

- **200** - Başarılı
- **201** - Oluşturuldu
- **400** - Geçersiz istek
- **401** - Yetkilendirme hatası
- **403** - Erişim yasak
- **404** - Bulunamadı
- **500** - Sunucu hatası

## Notlar

- Tüm token'lar `Bearer` prefix'i ile gönderilmelidir
- Pagination için `page` ve `limit` parametreleri kullanılır
- Tarih formatı ISO 8601 standardındadır
- Fiyatlar ondalık sayı olarak saklanır
- Görsel yüklemeleri için ayrı endpoint'ler geliştirilecektir

## Swagger Dokümantasyonu

API'nin tam dokümantasyonu Swagger UI'da mevcuttur:
- **URL**: `http://localhost:8080/swagger/index.html`
- Docker container çalışırken erişilebilir
- Tüm endpoint'ler, parametreler ve response'lar detaylı şekilde dokumentlanmıştır
