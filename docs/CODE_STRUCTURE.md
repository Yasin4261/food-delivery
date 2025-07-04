# 📁 Özgür Mutfak E-Commerce API - Kod Yapısı

## 🏗️ **Tam Modüler Yapı (TAMAMLANDI)**

Tüm katmanlar (handler, service, repository, model) artık işlevsel kategorilerine göre ayrı dosyalara bölündü. Bu sayede kod daha maintainable ve scalable hale geldi.

### 📂 **Yeni Dosya Yapısı:**

```
internal/
├── api/
│   ├── handler/           # HTTP handler'ları (kategori bazlı)
│   │   ├── base.go        # Temel bağımlılıklar
│   │   ├── auth.go        # Kimlik doğrulama
│   │   ├── user.go        # Kullanıcı profil işlemleri
│   │   ├── product.go     # Ürün işlemleri
│   │   ├── cart.go        # Sepet işlemleri
│   │   ├── order.go       # Sipariş işlemleri
│   │   └── admin.go       # Admin panel işlemleri
│   └── router.go          # Route tanımları
├── service/               # İş mantığı katmanı (kategori bazlı)
│   ├── user_service.go    # Kullanıcı iş mantığı
│   ├── product_service.go # Ürün iş mantığı
│   ├── cart_service.go    # Sepet iş mantığı
│   ├── order_service.go   # Sipariş iş mantığı
│   └── admin_service.go   # Admin iş mantığı
├── repository/            # Veri erişim katmanı (kategori bazlı)
│   ├── user_repository.go     # Kullanıcı veri erişimi
│   ├── product_repository.go  # Ürün veri erişimi
│   ├── cart_repository.go     # Sepet veri erişimi
│   └── order_repository.go    # Sipariş veri erişimi
└── model/                 # Veri modelleri (kategori bazlı)
    ├── user.go            # Kullanıcı modelleri
    ├── product.go         # Ürün modelleri
    ├── cart.go            # Sepet modelleri
    ├── order.go           # Sipariş modelleri
    └── common.go          # Ortak yapılar
```

### 🎯 **Katmanlar ve Sorumluluklar:**

#### 1. **Handler Katmanı** - HTTP İsteklerini Yönetir
```go
// auth.go - Kimlik doğrulama
- Login()     // POST /api/v1/auth/login
- Register()  // POST /api/v1/auth/register
- Logout()    // POST /api/v1/auth/logout

// user.go - Kullanıcı profil işlemleri
- GetProfile()    // GET /api/v1/profile
- UpdateProfile() // PUT /api/v1/profile

// product.go - Ürün işlemleri
- GetProducts()   // GET /api/v1/products
- GetProduct()    // GET /api/v1/products/:id

// cart.go - Sepet işlemleri
- GetCart()        // GET /api/v1/cart
- AddToCart()      // POST /api/v1/cart/items
- RemoveFromCart() // DELETE /api/v1/cart/items/:id

// order.go - Sipariş işlemleri
- GetOrders()    // GET /api/v1/orders
- CreateOrder()  // POST /api/v1/orders
- GetOrder()     // GET /api/v1/orders/:id

// admin.go - Admin panel işlemleri
- AdminGetProducts()    // GET /api/v1/admin/products
- AdminCreateProduct()  // POST /api/v1/admin/products
- AdminUpdateProduct()  // PUT /api/v1/admin/products/:id
- AdminDeleteProduct()  // DELETE /api/v1/admin/products/:id
- AdminGetOrders()      // GET /api/v1/admin/orders
- AdminUpdateOrderStatus() // PUT /api/v1/admin/orders/:id/status
```

#### 2. **Service Katmanı** - İş Mantığı Yönetir
```go
// user_service.go - Kullanıcı iş mantığı
- Register()     // Kullanıcı kaydı
- Login()        // Giriş işlemi
- GetProfile()   // Profil bilgisi
- UpdateProfile() // Profil güncelleme

// product_service.go - Ürün iş mantığı
- GetAllProducts()  // Tüm ürünler
- GetProduct()      // Tek ürün
- CreateProduct()   // Ürün oluşturma
- UpdateProduct()   // Ürün güncelleme
- DeleteProduct()   // Ürün silme
- SearchProducts()  // Ürün arama

// cart_service.go - Sepet iş mantığı
- GetOrCreateCart() // Sepet alma/oluşturma
- AddItem()         // Öğe ekleme
- UpdateItem()      // Öğe güncelleme
- RemoveItem()      // Öğe silme
- ClearCart()       // Sepeti temizleme
- GetCartTotal()    // Sepet toplamı

// order_service.go - Sipariş iş mantığı
- CreateOrder()     // Sipariş oluşturma
- GetUserOrders()   // Kullanıcı siparişleri
- GetOrder()        // Tek sipariş
- UpdateOrderStatus() // Sipariş durumu güncelleme
- CancelOrder()     // Sipariş iptal

// admin_service.go - Admin iş mantığı
- GetAllUsers()     // Tüm kullanıcılar
- GetAllOrders()    // Tüm siparişler
- GetDashboardStats() // Dashboard istatistikleri
- UpdateUserRole()  // Kullanıcı rol güncelleme
```

#### 3. **Repository Katmanı** - Veri Erişimi Yönetir
```go
// user_repository.go - Kullanıcı veri erişimi
- Create()      // Kullanıcı oluşturma
- GetByEmail()  // Email ile kullanıcı bulma
- GetByID()     // ID ile kullanıcı bulma
- GetAll()      // Tüm kullanıcılar
- Update()      // Kullanıcı güncelleme
- Delete()      // Kullanıcı silme

// product_repository.go - Ürün veri erişimi
- Create()         // Ürün oluşturma
- GetAll()         // Tüm ürünler
- GetByID()        // ID ile ürün bulma
- GetByCategory()  // Kategori ile ürün bulma
- Search()         // Ürün arama
- Update()         // Ürün güncelleme
- Delete()         // Ürün silme
- GetLowStock()    // Stok azalan ürünler

// cart_repository.go - Sepet veri erişimi
- Create()         // Sepet oluşturma
- GetByUserID()    // Kullanıcı sepeti
- CreateCartItem() // Sepet öğesi ekleme
- GetCartItem()    // Sepet öğesi alma
- GetCartItems()   // Sepet öğeleri
- UpdateCartItem() // Sepet öğesi güncelleme
- DeleteCartItem() // Sepet öğesi silme
- ClearCart()      // Sepeti temizleme

// order_repository.go - Sipariş veri erişimi
- Create()           // Sipariş oluşturma
- CreateOrderItem()  // Sipariş öğesi oluşturma
- GetByUserID()      // Kullanıcı siparişleri
- GetByID()          // ID ile sipariş bulma
- GetAll()           // Tüm siparişler
- GetByStatus()      // Duruma göre siparişler
- GetRecent()        // Son siparişler
- UpdateStatus()     // Sipariş durumu güncelleme
- GetOrderItems()    // Sipariş öğeleri
```

#### 4. **Model Katmanı** - Veri Yapıları Tanımlar
```go
// user.go - Kullanıcı modelleri
- User                    // Kullanıcı struct
- LoginRequest           // Giriş isteği
- RegisterRequest        // Kayıt isteği
- AuthResponse          // Kimlik doğrulama yanıtı
- UpdateProfileRequest  // Profil güncelleme isteği

// product.go - Ürün modelleri
- Product               // Ürün struct
- Category             // Kategori struct
- CreateProductRequest // Ürün oluşturma isteği
- UpdateProductRequest // Ürün güncelleme isteği
- ProductResponse      // Ürün yanıtı
- ProductListResponse  // Ürün listesi yanıtı

// cart.go - Sepet modelleri
- Cart                   // Sepet struct
- CartItem              // Sepet öğesi struct
- AddToCartRequest      // Sepete ekleme isteği
- UpdateCartItemRequest // Sepet öğesi güncelleme isteği
- CartResponse          // Sepet yanıtı
- CartItemResponse      // Sepet öğesi yanıtı

// order.go - Sipariş modelleri
- Order                    // Sipariş struct
- OrderItem               // Sipariş öğesi struct
- CreateOrderRequest      // Sipariş oluşturma isteği
- OrderItemInput          // Sipariş öğesi girdisi
- UpdateOrderStatusRequest // Sipariş durumu güncelleme isteği
- OrderResponse           // Sipariş yanıtı
- OrderItemResponse       // Sipariş öğesi yanıtı

// common.go - Ortak yapılar
- DashboardStats        // Dashboard istatistikleri
- ErrorResponse         // Hata yanıtı
- SuccessResponse       // Başarı yanıtı
- PaginationResponse    // Sayfalama yanıtı
- HealthCheckResponse   // Sağlık kontrolü yanıtı
- JWTClaims            // JWT claims
```
- RemoveFromCart() // DELETE /api/v1/cart/item/:id
- ClearCart()      // DELETE /api/v1/cart/clear
```

#### 6. **`order.go`** - Sipariş İşlemleri
```go
// Order management endpoints
- GetOrders()         // GET /api/v1/orders
- CreateOrder()       // POST /api/v1/orders
- GetOrder()          // GET /api/v1/orders/:id
- UpdateOrderStatus() // PUT /api/v1/orders/:id/status (Admin)
- CancelOrder()       // PUT /api/v1/orders/:id/cancel
```

#### 7. **`admin.go`** - Admin İşlemleri
```go
// Admin-only endpoints
- AdminGetProducts()       // GET /api/v1/admin/products
- AdminCreateProduct()     // POST /api/v1/admin/products
- AdminUpdateProduct()     // PUT /api/v1/admin/products/:id
- AdminDeleteProduct()     // DELETE /api/v1/admin/products/:id
- AdminGetOrders()         // GET /api/v1/admin/orders
- AdminUpdateOrderStatus() // PUT /api/v1/admin/orders/:id/status
- AdminGetUsers()          // GET /api/v1/admin/users
- AdminGetUser()           // GET /api/v1/admin/users/:id
- AdminGetDashboard()      // GET /api/v1/admin/dashboard
```

---

## 🔗 **API Route Yapısı:**

### **Public Routes (Authentication gerektirmez):**
```
POST   /api/v1/auth/login
POST   /api/v1/auth/register
GET    /api/v1/products
GET    /api/v1/products/:id
```

### **Protected Routes (JWT Token gerekir):**
```
POST   /api/v1/auth/logout
GET    /api/v1/user/profile
PUT    /api/v1/user/profile
GET    /api/v1/cart
POST   /api/v1/cart/add
PUT    /api/v1/cart/item/:id
DELETE /api/v1/cart/item/:id
DELETE /api/v1/cart/clear
GET    /api/v1/orders
POST   /api/v1/orders
GET    /api/v1/orders/:id
PUT    /api/v1/orders/:id/cancel
```

### **Admin Routes (JWT + Admin Role gerekir):**
```
GET    /api/v1/admin/dashboard
GET    /api/v1/admin/products
POST   /api/v1/admin/products
PUT    /api/v1/admin/products/:id
DELETE /api/v1/admin/products/:id
GET    /api/v1/admin/orders
PUT    /api/v1/admin/orders/:id/status
GET    /api/v1/admin/users
GET    /api/v1/admin/users/:id
```

---

## 🎯 **Avantajları:**

### ✅ **Daha İyi Organizasyon:**
- Her dosya tek bir sorumluluğa odaklanır
- Kod bulmak ve düzenlemek daha kolay
- Takım çalışmasında çakışma riski azalır

### ✅ **Bakım Kolaylığı:**
- İlgili işlevler bir arada
- Test yazma daha kolay
- Hata ayıklama daha hızlı

### ✅ **Ölçeklenebilirlik:**
- Yeni özellikler eklenmesi kolay
- Kod tekrarı azalır
- Clean Architecture prensiplerine uygun

---

## 🔄 **Gelecek Geliştirmeler:**

### **Service Layer Ayrımı:**
```
internal/service/
├── user_service.go
├── product_service.go
├── cart_service.go
├── order_service.go
└── admin_service.go
```

### **Repository Layer Ayrımı:**
```
internal/repository/
├── user_repository.go
├── product_repository.go
├── cart_repository.go
└── order_repository.go
```

### **Model Ayrımı:**
```
internal/model/
├── user.go
├── product.go
├── cart.go
├── order.go
└── requests.go
```

---

## 🧪 **Test Durumu:**

### ✅ **Çalışan Endpoint'ler:**
- Authentication (Login, Register, Logout)
- Products (Basic listing)

### 🚧 **Geliştirme Aşamasında:**
- Cart Management
- Order Processing
- Admin Panel
- User Profile Management

---

## 📝 **Kod Kalitesi:**

### **Before (Tek Dosya):**
```
handlers.go - 222 satır
├── Auth handlers
├── Product handlers
├── User handlers
├── Cart handlers
├── Order handlers
└── Admin handlers
```

### **After (Modüler Yapı):**
```
base.go    - 20 satır  (Dependencies)
auth.go    - 60 satır  (Authentication)
user.go    - 40 satır  (User Profile)
product.go - 45 satır  (Products)
cart.go    - 85 satır  (Shopping Cart)
order.go   - 75 satır  (Orders)
admin.go   - 95 satır  (Admin Panel)
```

### **Sonuç:**
- ✅ Daha temiz kod
- ✅ Daha iyi okunabilirlik
- ✅ Kolay bakım
- ✅ Daha az karmaşıklık

---

Bu yeni yapı ile kod daha profesyonel, sürdürülebilir ve genişletilebilir hale geldi! 🚀
