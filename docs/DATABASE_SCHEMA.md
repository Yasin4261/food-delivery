# 📊 Özgür Mutfak E-Commerce Database Schema

## 🗂️ Database Visualization Tools

### 1. **Adminer (Yeni Eklendi)**
- **URL:** http://localhost:8082
- **Giriş:** 
  - Server: `db`
  - Username: `postgres` 
  - Password: `password`
  - Database: `ecommerce_db`

### 2. **pgAdmin (Mevcut)**
- **URL:** http://localhost:8081
- **Email:** admin@ecommerce.com
- **Password:** admin123

---

## 🏗️ Database Schema Diagram

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│     USERS       │     │   CATEGORIES     │     │    PRODUCTS     │
├─────────────────┤     ├──────────────────┤     ├─────────────────┤
│ 🔑 id (PK)      │     │ 🔑 id (PK)       │     │ 🔑 id (PK)      │
│ 📧 email (UQ)   │     │ 📝 name          │◄────┤ 🔗 category_id  │
│ 🔒 password     │     │ 📅 created_at    │     │ 📝 name         │
│ 👤 first_name   │     │ 📅 updated_at    │     │ 📄 description  │
│ 👤 last_name    │     └──────────────────┘     │ 💰 price        │
│ 🎭 role         │                              │ 📦 stock        │
│ 📅 created_at   │                              │ 🖼️ image_url    │
│ 📅 updated_at   │                              │ 📅 created_at   │
└─────────────────┘                              │ 📅 updated_at   │
         │                                       └─────────────────┘
         │                                                │
         ▼                                                │
┌─────────────────┐                                       │
│     CARTS       │                                       │
├─────────────────┤                                       │
│ 🔑 id (PK)      │                                       │
│ 🔗 user_id (FK) │                                       │
│ 📅 created_at   │                                       │
│ 📅 updated_at   │                                       │
└─────────────────┘                                       │
         │                                                │
         ▼                                                ▼
┌─────────────────┐                              ┌─────────────────┐
│   CART_ITEMS    │                              │   ORDER_ITEMS   │
├─────────────────┤                              ├─────────────────┤
│ 🔑 id (PK)      │                              │ 🔑 id (PK)      │
│ 🔗 cart_id (FK) │                              │ 🔗 order_id (FK)│
│ 🔗 product_id   │◄─────────────────────────────┤ 🔗 product_id   │
│ 🔢 quantity     │                              │ 🔢 quantity     │
│ 📅 created_at   │                              │ 💰 price        │
│ 📅 updated_at   │                              │ 📅 created_at   │
└─────────────────┘                              │ 📅 updated_at   │
                                                 └─────────────────┘
         ┌─────────────────┐                              ▲
         │     ORDERS      │                              │
         ├─────────────────┤                              │
         │ 🔑 id (PK)      │──────────────────────────────┘
         │ 🔗 user_id (FK) │◄─────────────────────────────┐
         │ 💰 total        │                              │
         │ 📊 status       │                              │
         │ 📍 address      │                              │
         │ 📅 created_at   │                              │
         │ 📅 updated_at   │                              │
         └─────────────────┘                              │
                                                         │
                            ┌─────────────────────────────┘
                            │
                    ┌─────────────────┐
                    │     USERS       │
                    │   (Reference)   │
                    └─────────────────┘
```

## 🔗 Table Relations

### **Primary Relationships:**
- 👤 **Users** → 🛒 **Carts** (1:N)
- 👤 **Users** → 📦 **Orders** (1:N)
- 🏷️ **Categories** → 🛍️ **Products** (1:N)
- 🛒 **Carts** → 📋 **Cart Items** (1:N)
- 📦 **Orders** → 📋 **Order Items** (1:N)
- 🛍️ **Products** → 📋 **Cart Items** (1:N)
- 🛍️ **Products** → 📋 **Order Items** (1:N)

### **Business Logic:**
1. **User Registration/Login** ✅ Implemented
2. **Product Catalog** 🚧 Structure ready
3. **Shopping Cart** 🚧 Structure ready
4. **Order Management** 🚧 Structure ready

---

## 📋 Table Details

### 🧑‍💼 **USERS Table**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) DEFAULT 'customer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 🏷️ **CATEGORIES Table**
```sql
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 🛍️ **PRODUCTS Table**
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INTEGER DEFAULT 0,
    category_id INTEGER REFERENCES categories(id),
    image_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 🛒 **CARTS Table**
```sql
CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 📦 **ORDERS Table**
```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    address TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🛠️ Advanced Visualization Tools

### **Online Tools:**
1. **dbdiagram.io** - Database design tool
2. **drawSQL** - Visual database designer
3. **Lucidchart** - Professional diagrams

### **Desktop Tools:**
1. **DBeaver** - Universal database tool
2. **DataGrip** - JetBrains database IDE
3. **TablePlus** - Modern database client

### **VS Code Extensions:**
1. **PostgreSQL** - Database management
2. **Database Client** - Multi-database support
3. **ERD Editor** - Entity relationship diagrams
