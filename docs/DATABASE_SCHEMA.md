# 🗄️ Database Schema Documentation

## Overview

Bu dokümantasyon, Özgür Mutfak uygulamasının veritabanı şemasını ve veri modelini açıklar.

## 📊 Entity Relationship Diagram

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Users    │    │    Chefs    │    │    Meals    │
│             │    │             │    │             │
│ id (PK)     │ 1  │ id (PK)     │ 1  │ id (PK)     │
│ email       │────│ user_id(FK) │────│ chef_id(FK) │
│ password    │    │ business... │  N │ name        │
│ first_name  │    │ address     │    │ description │
│ last_name   │    │ phone       │    │ price       │
│ role        │    │ is_verified │    │ image_url   │
│ is_active   │    │ created_at  │    │ category    │
│ created_at  │    │ updated_at  │    │ is_active   │
│ updated_at  │    └─────────────┘    │ created_at  │
└─────────────┘                       │ updated_at  │
       │                              └─────────────┘
       │1                                     │
       │                                      │
       │N                                     │N
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Orders    │    │ OrderItems  │    │  CartItems  │
│             │    │             │    │             │
│ id (PK)     │ 1  │ id (PK)     │    │ id (PK)     │
│ user_id(FK) │────│ order_id(FK)│    │ user_id(FK) │
│ total_amount│  N │ meal_id(FK) │    │ meal_id(FK) │
│ status      │    │ quantity    │    │ quantity    │
│ order_date  │    │ unit_price  │    │ created_at  │
│ delivery... │    │ total_price │    │ updated_at  │
│ created_at  │    │ created_at  │    └─────────────┘
│ updated_at  │    │ updated_at  │
└─────────────┘    └─────────────┘
       │
       │1
       │
       │N
┌─────────────┐
│   Reviews   │
│             │
│ id (PK)     │
│ user_id(FK) │
│ chef_id(FK) │
│ meal_id(FK) │
│ rating      │
│ comment     │
│ created_at  │
│ updated_at  │
└─────────────┘
```

## 📋 Table Definitions

### 1. Users Table

**Purpose**: Kullanıcı hesapları ve kimlik doğrulama bilgileri

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(20) DEFAULT 'customer' CHECK (role IN ('customer', 'chef', 'admin')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(is_active);
```

**Fields Description**:
- `id`: Birincil anahtar
- `email`: Benzersiz kullanıcı email adresi
- `password_hash`: bcrypt ile hashlenmiş şifre
- `role`: Kullanıcı rolü (customer, chef, admin)
- `is_active`: Hesap aktif durumu

### 2. Chefs Table

**Purpose**: Aşçı profilleri ve iş bilgileri

```sql
CREATE TABLE chefs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    business_name VARCHAR(255),
    bio TEXT,
    address TEXT,
    phone VARCHAR(20),
    experience_years INTEGER DEFAULT 0,
    specialties TEXT[],
    average_rating DECIMAL(3,2) DEFAULT 0.00,
    total_reviews INTEGER DEFAULT 0,
    is_verified BOOLEAN DEFAULT FALSE,
    delivery_radius INTEGER DEFAULT 10, -- km
    min_order_amount DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX idx_chefs_user_id ON chefs(user_id);
CREATE INDEX idx_chefs_verified ON chefs(is_verified);
CREATE INDEX idx_chefs_rating ON chefs(average_rating DESC);
```

**Fields Description**:
- `user_id`: Users tablosuna foreign key
- `business_name`: İşletme adı
- `specialties`: Uzmanlık alanları (array)
- `is_verified`: Doğrulanmış aşçı durumu
- `delivery_radius`: Teslimat yarıçapı (km)

### 3. Meals Table

**Purpose**: Yemek menüleri ve detayları

```sql
CREATE TABLE meals (
    id SERIAL PRIMARY KEY,
    chef_id INTEGER REFERENCES chefs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    image_url VARCHAR(500),
    category VARCHAR(100),
    ingredients TEXT[],
    allergens TEXT[],
    preparation_time INTEGER, -- minutes
    serving_size INTEGER DEFAULT 1,
    calories INTEGER,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    is_vegan BOOLEAN DEFAULT FALSE,
    is_gluten_free BOOLEAN DEFAULT FALSE,
    is_available BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_meals_chef_id ON meals(chef_id);
CREATE INDEX idx_meals_category ON meals(category);
CREATE INDEX idx_meals_price ON meals(price);
CREATE INDEX idx_meals_available ON meals(is_available);
CREATE INDEX idx_meals_dietary ON meals(is_vegetarian, is_vegan, is_gluten_free);
```

**Fields Description**:
- `chef_id`: Chefs tablosuna foreign key
- `ingredients`: Malzeme listesi (array)
- `allergens`: Alerjen bilgileri (array)
- `dietary flags`: Diyet tercihleri için boolean alanlar

### 4. Orders Table

**Purpose**: Sipariş yönetimi

```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'delivered', 'cancelled')),
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_address TEXT NOT NULL,
    delivery_phone VARCHAR(20),
    delivery_notes TEXT,
    estimated_delivery TIMESTAMP,
    actual_delivery TIMESTAMP,
    payment_method VARCHAR(20) DEFAULT 'cash' CHECK (payment_method IN ('cash', 'card', 'online')),
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_date ON orders(order_date DESC);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
```

**Fields Description**:
- `status`: Sipariş durumu (pending → confirmed → preparing → ready → delivered)
- `payment_status`: Ödeme durumu
- `estimated_delivery`: Tahmini teslimat zamanı

### 5. Order Items Table

**Purpose**: Sipariş detayları

```sql
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    meal_id INTEGER REFERENCES meals(id) ON DELETE SET NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    special_instructions TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_meal_id ON order_items(meal_id);

-- Check constraint
ALTER TABLE order_items ADD CONSTRAINT chk_total_price 
CHECK (total_price = unit_price * quantity);
```

**Fields Description**:
- `unit_price`: Sipariş anındaki birim fiyat (fiyat değişikliklerinden korunmak için)
- `special_instructions`: Özel talimatlar

### 6. Cart Items Table

**Purpose**: Sepet yönetimi

```sql
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    meal_id INTEGER REFERENCES meals(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
CREATE INDEX idx_cart_items_meal_id ON cart_items(meal_id);

-- Unique constraint: Bir kullanıcı aynı meal'den sadece bir item'a sahip olabilir
CREATE UNIQUE INDEX idx_cart_items_user_meal ON cart_items(user_id, meal_id);
```

### 7. Reviews Table

**Purpose**: Değerlendirme ve yorumlar

```sql
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    chef_id INTEGER REFERENCES chefs(id) ON DELETE CASCADE,
    meal_id INTEGER REFERENCES meals(id) ON DELETE SET NULL,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    is_verified BOOLEAN DEFAULT FALSE, -- Doğrulanmış satın alma
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_reviews_chef_id ON reviews(chef_id);
CREATE INDEX idx_reviews_meal_id ON reviews(meal_id);
CREATE INDEX idx_reviews_rating ON reviews(rating DESC);
CREATE INDEX idx_reviews_verified ON reviews(is_verified);

-- Unique constraint: Bir kullanıcı aynı sipariş için sadece bir review verebilir
CREATE UNIQUE INDEX idx_reviews_user_order ON reviews(user_id, order_id);
```

## 🔍 Views and Queries

### Chef Statistics View

```sql
CREATE VIEW chef_stats AS
SELECT 
    c.id,
    c.business_name,
    c.average_rating,
    c.total_reviews,
    COUNT(DISTINCT m.id) as total_meals,
    COUNT(DISTINCT o.id) as total_orders,
    SUM(oi.total_price) as total_revenue
FROM chefs c
LEFT JOIN meals m ON c.id = m.chef_id AND m.is_active = TRUE
LEFT JOIN order_items oi ON m.id = oi.meal_id
LEFT JOIN orders o ON oi.order_id = o.id AND o.status = 'delivered'
GROUP BY c.id, c.business_name, c.average_rating, c.total_reviews;
```

### Popular Meals View

```sql
CREATE VIEW popular_meals AS
SELECT 
    m.id,
    m.name,
    m.price,
    c.business_name as chef_name,
    COUNT(oi.id) as order_count,
    AVG(r.rating) as average_rating,
    COUNT(r.id) as review_count
FROM meals m
JOIN chefs c ON m.chef_id = c.id
LEFT JOIN order_items oi ON m.id = oi.meal_id
LEFT JOIN reviews r ON m.id = r.meal_id
WHERE m.is_active = TRUE AND m.is_available = TRUE
GROUP BY m.id, m.name, m.price, c.business_name
ORDER BY order_count DESC, average_rating DESC;
```

## 🔄 Triggers and Functions

### Update Chef Rating Trigger

```sql
-- Function to update chef rating
CREATE OR REPLACE FUNCTION update_chef_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chefs SET 
        average_rating = (
            SELECT ROUND(AVG(rating), 2) 
            FROM reviews 
            WHERE chef_id = COALESCE(NEW.chef_id, OLD.chef_id)
        ),
        total_reviews = (
            SELECT COUNT(*) 
            FROM reviews 
            WHERE chef_id = COALESCE(NEW.chef_id, OLD.chef_id)
        ),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.chef_id, OLD.chef_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger
CREATE TRIGGER trigger_update_chef_rating
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_chef_rating();
```

### Update Timestamps Trigger

```sql
-- Generic function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chefs_updated_at BEFORE UPDATE ON chefs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_meals_updated_at BEFORE UPDATE ON meals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

## 💾 Data Migration Strategy

### Migration Files Structure

```
migrations/
├── 001_initial_schema.sql           # Base tables
├── 002_home_cooked_meals.sql        # Meal-specific features
├── 003_create_home_cooked_meals_schema.sql
├── 004_create_tables_only.sql       # Clean table creation
├── 005_insert_test_data.sql         # Sample data
├── 006_add_is_active_to_users.sql   # User activation
├── 007_add_multi_chef_support.sql   # Multi-chef features
└── 008_multi_vendor_orders.sql      # Order improvements
```

### Sample Data Scripts

```sql
-- Insert test users
INSERT INTO users (email, password_hash, first_name, last_name, role) VALUES
('chef1@test.com', '$2b$10$...', 'Mehmet', 'Yılmaz', 'chef'),
('customer1@test.com', '$2b$10$...', 'Ayşe', 'Demir', 'customer'),
('admin@test.com', '$2b$10$...', 'Admin', 'User', 'admin');

-- Insert test chefs
INSERT INTO chefs (user_id, business_name, bio, address) VALUES
(1, 'Mehmet Usta Mutfağı', 'Geleneksel Türk mutfağı uzmanı', 'İstanbul, Kadıköy');

-- Insert test meals
INSERT INTO meals (chef_id, name, description, price, category) VALUES
(1, 'Ev Yapımı Mantı', 'El açması hamur ile hazırlanan mantı', 25.00, 'Ana Yemek'),
(1, 'Mercimek Çorbası', 'Taze sebzelerle yapılan mercimek çorbası', 8.00, 'Çorba');
```

## 🔒 Security Considerations

### Data Protection

1. **Password Security**: bcrypt hashing (min cost 10)
2. **Email Uniqueness**: Unique constraint + validation
3. **Role-based Access**: Enum constraints for roles
4. **Soft Delete**: is_active flags instead of hard delete
5. **Audit Trail**: created_at, updated_at timestamps

### Performance Optimizations

1. **Indexing Strategy**:
   - Primary keys: Auto-indexed
   - Foreign keys: Indexed for joins
   - Search fields: Composite indexes
   - Query-specific indexes

2. **Query Optimization**:
   - Use prepared statements
   - Implement pagination
   - Avoid N+1 queries
   - Use joins instead of multiple queries

3. **Connection Pooling**:
   ```go
   // Go database configuration
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

## 📊 Performance Monitoring

### Key Metrics to Track

1. **Query Performance**:
   - Slow query log
   - Query execution plans
   - Index usage statistics

2. **Database Health**:
   - Connection pool utilization
   - Lock wait times
   - Table size growth

3. **Business Metrics**:
   - Order completion rates
   - Average order value
   - Chef activity levels

### Monitoring Queries

```sql
-- Find slow queries
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## 🔄 Backup and Recovery

### Backup Strategy

1. **Daily Full Backups**: Complete database dump
2. **Continuous WAL Archiving**: Point-in-time recovery
3. **Regular Testing**: Restore procedures validation

```bash
# Full backup
pg_dump -h localhost -U postgres ozgur_mutfak > backup_$(date +%Y%m%d).sql

# Restore
psql -h localhost -U postgres -d ozgur_mutfak < backup_20240101.sql
```

---

**Note**: Bu şema dokümantasyonu, veritabanı değişiklikleri ile birlikte güncellenmelidir.
