# 📖 API Reference Documentation

## Overview

Bu dokümantasyon, Özgür Mutfak REST API'sının detaylı referansını içerir.

## 🔗 Base Information

- **Base URL**: `https://api.ozgurmutfak.com/api/v1`
- **API Version**: v1.0.0
- **Authentication**: JWT Bearer Token
- **Content Type**: `application/json`
- **Rate Limit**: 100 requests/minute per user

## 🔐 Authentication

### JWT Token Authentication

API'ya erişim için JWT token gereklidir. Token'ı Authorization header'ında gönderin:

```http
Authorization: Bearer <your-jwt-token>
```

### Token Endpoints

#### Login
```http
POST /auth/login
```

**Request Body:**
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_at": "2024-01-02T15:04:05Z",
        "user": {
            "id": 1,
            "email": "user@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "role": "customer"
        }
    }
}
```

#### Register
```http
POST /auth/register
```

**Request Body:**
```json
{
    "email": "newuser@example.com",
    "password": "password123",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+90555123456"
}
```

#### Refresh Token
```http
POST /auth/refresh
```

**Headers:**
```http
Authorization: Bearer <current-token>
```

## 👤 Users API

### Get Current User
```http
GET /users/me
Authorization: Bearer <token>
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "id": 1,
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "phone": "+90555123456",
        "role": "customer",
        "is_active": true,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z"
    }
}
```

### Update User Profile
```http
PUT /users/me
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "first_name": "John Updated",
    "last_name": "Doe Updated",
    "phone": "+90555654321"
}
```

### Change Password
```http
POST /users/me/change-password
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "current_password": "oldpassword123",
    "new_password": "newpassword123"
}
```

## 👨‍🍳 Chefs API

### List Chefs
```http
GET /chefs
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `search` (string): Search in business name or bio
- `verified` (bool): Filter verified chefs only
- `sort` (string): Sort by (`rating`, `name`, `created_at`)
- `order` (string): Sort order (`asc`, `desc`)

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "user_id": 2,
            "business_name": "Mehmet Usta Mutfağı",
            "bio": "Geleneksel Türk mutfağı uzmanı",
            "address": "İstanbul, Kadıköy",
            "phone": "+90555123456",
            "experience_years": 15,
            "specialties": ["Turkish", "Mediterranean"],
            "average_rating": 4.8,
            "total_reviews": 124,
            "is_verified": true,
            "delivery_radius": 10,
            "min_order_amount": 30.00,
            "created_at": "2024-01-01T10:00:00Z"
        }
    ],
    "meta": {
        "page": 1,
        "limit": 20,
        "total": 50,
        "pages": 3
    }
}
```

### Get Chef Details
```http
GET /chefs/{chef_id}
```

### Become a Chef
```http
POST /chefs/apply
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "business_name": "My Restaurant",
    "bio": "Professional chef with 10 years experience",
    "address": "İstanbul, Beşiktaş",
    "phone": "+90555123456",
    "experience_years": 10,
    "specialties": ["Italian", "Turkish"],
    "delivery_radius": 15,
    "min_order_amount": 25.00
}
```

### Update Chef Profile
```http
PUT /chefs/me
Authorization: Bearer <token>
```

## 🍽️ Meals API

### List Meals
```http
GET /meals
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `chef_id` (int): Filter by chef
- `category` (string): Filter by category
- `search` (string): Search in name/description
- `min_price` (float): Minimum price filter
- `max_price` (float): Maximum price filter
- `vegetarian` (bool): Vegetarian meals only
- `vegan` (bool): Vegan meals only
- `gluten_free` (bool): Gluten-free meals only
- `available` (bool): Available meals only (default: true)
- `sort` (string): Sort by (`price`, `name`, `rating`, `created_at`)

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "chef_id": 1,
            "name": "Ev Yapımı Mantı",
            "description": "El açması hamur ile hazırlanan geleneksel mantı",
            "price": 25.00,
            "image_url": "https://example.com/images/manti.jpg",
            "category": "Ana Yemek",
            "ingredients": ["Un", "Et", "Soğan", "Baharat"],
            "allergens": ["Gluten"],
            "preparation_time": 45,
            "serving_size": 1,
            "calories": 450,
            "is_vegetarian": false,
            "is_vegan": false,
            "is_gluten_free": false,
            "is_available": true,
            "chef": {
                "id": 1,
                "business_name": "Mehmet Usta Mutfağı",
                "average_rating": 4.8
            },
            "created_at": "2024-01-01T10:00:00Z"
        }
    ],
    "meta": {
        "page": 1,
        "limit": 20,
        "total": 150,
        "pages": 8
    }
}
```

### Get Meal Details
```http
GET /meals/{meal_id}
```

### Create Meal (Chef Only)
```http
POST /meals
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "name": "Yeni Yemek",
    "description": "Lezzetli yemek açıklaması",
    "price": 30.00,
    "category": "Ana Yemek",
    "ingredients": ["Malzeme 1", "Malzeme 2"],
    "allergens": ["Gluten", "Süt"],
    "preparation_time": 30,
    "serving_size": 1,
    "calories": 400,
    "is_vegetarian": false,
    "is_vegan": false,
    "is_gluten_free": false,
    "image_url": "https://example.com/image.jpg"
}
```

### Update Meal (Chef Only)
```http
PUT /meals/{meal_id}
Authorization: Bearer <token>
```

### Delete Meal (Chef Only)
```http
DELETE /meals/{meal_id}
Authorization: Bearer <token>
```

### Upload Meal Image
```http
POST /meals/{meal_id}/image
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Form Data:**
- `image`: Image file (JPEG, PNG, max 5MB)

## 🛒 Cart API

### Get Cart
```http
GET /cart
Authorization: Bearer <token>
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "items": [
            {
                "id": 1,
                "meal_id": 1,
                "quantity": 2,
                "meal": {
                    "id": 1,
                    "name": "Ev Yapımı Mantı",
                    "price": 25.00,
                    "image_url": "https://example.com/manti.jpg",
                    "chef": {
                        "id": 1,
                        "business_name": "Mehmet Usta Mutfağı"
                    }
                },
                "created_at": "2024-01-01T10:00:00Z"
            }
        ],
        "total_items": 2,
        "total_amount": 50.00
    }
}
```

### Add to Cart
```http
POST /cart/items
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "meal_id": 1,
    "quantity": 2
}
```

### Update Cart Item
```http
PUT /cart/items/{item_id}
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "quantity": 3
}
```

### Remove from Cart
```http
DELETE /cart/items/{item_id}
Authorization: Bearer <token>
```

### Clear Cart
```http
DELETE /cart
Authorization: Bearer <token>
```

## 📦 Orders API

### List Orders
```http
GET /orders
Authorization: Bearer <token>
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `status` (string): Filter by status
- `start_date` (date): Filter from date (YYYY-MM-DD)
- `end_date` (date): Filter to date (YYYY-MM-DD)

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "user_id": 1,
            "total_amount": 75.00,
            "status": "delivered",
            "order_date": "2024-01-01T12:00:00Z",
            "delivery_address": "İstanbul, Beşiktaş, Örnek Mahalle",
            "delivery_phone": "+90555123456",
            "delivery_notes": "Kapı kodu: 1234",
            "estimated_delivery": "2024-01-01T13:30:00Z",
            "actual_delivery": "2024-01-01T13:25:00Z",
            "payment_method": "card",
            "payment_status": "paid",
            "items": [
                {
                    "id": 1,
                    "meal_id": 1,
                    "quantity": 2,
                    "unit_price": 25.00,
                    "total_price": 50.00,
                    "meal": {
                        "id": 1,
                        "name": "Ev Yapımı Mantı",
                        "chef": {
                            "id": 1,
                            "business_name": "Mehmet Usta Mutfağı"
                        }
                    }
                }
            ],
            "created_at": "2024-01-01T12:00:00Z"
        }
    ],
    "meta": {
        "page": 1,
        "limit": 20,
        "total": 25,
        "pages": 2
    }
}
```

### Get Order Details
```http
GET /orders/{order_id}
Authorization: Bearer <token>
```

### Create Order
```http
POST /orders
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "delivery_address": "İstanbul, Beşiktaş, Örnek Mahalle No:123",
    "delivery_phone": "+90555123456",
    "delivery_notes": "Kapı kodu: 1234",
    "payment_method": "card"
}
```

### Update Order Status (Chef Only)
```http
PATCH /orders/{order_id}/status
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "status": "preparing"
}
```

**Allowed Status Transitions:**
- `pending` → `confirmed`
- `confirmed` → `preparing`
- `preparing` → `ready`
- `ready` → `delivered`
- Any status → `cancelled` (within time limit)

### Cancel Order
```http
POST /orders/{order_id}/cancel
Authorization: Bearer <token>
```

## ⭐ Reviews API

### List Reviews
```http
GET /reviews
```

**Query Parameters:**
- `chef_id` (int): Filter by chef
- `meal_id` (int): Filter by meal
- `user_id` (int): Filter by user
- `rating` (int): Filter by rating (1-5)
- `verified` (bool): Filter verified reviews only

### Get Review Details
```http
GET /reviews/{review_id}
```

### Create Review
```http
POST /reviews
Authorization: Bearer <token>
```

**Request Body:**
```json
{
    "chef_id": 1,
    "meal_id": 1,
    "order_id": 1,
    "rating": 5,
    "comment": "Harika bir yemekti, kesinlikle tavsiye ederim!"
}
```

### Update Review
```http
PUT /reviews/{review_id}
Authorization: Bearer <token>
```

### Delete Review
```http
DELETE /reviews/{review_id}
Authorization: Bearer <token>
```

## 🔍 Search API

### Global Search
```http
GET /search
```

**Query Parameters:**
- `q` (string, required): Search query
- `type` (string): Filter by type (`meals`, `chefs`, `all`)
- `category` (string): Filter meals by category
- `location` (string): Filter chefs by location

**Response (200):**
```json
{
    "success": true,
    "data": {
        "meals": [
            {
                "id": 1,
                "name": "Ev Yapımı Mantı",
                "price": 25.00,
                "chef": {
                    "business_name": "Mehmet Usta Mutfağı"
                }
            }
        ],
        "chefs": [
            {
                "id": 1,
                "business_name": "Mehmet Usta Mutfağı",
                "average_rating": 4.8
            }
        ]
    },
    "meta": {
        "query": "mantı",
        "results_count": {
            "meals": 5,
            "chefs": 2,
            "total": 7
        }
    }
}
```

## 📊 Analytics API (Admin Only)

### Dashboard Stats
```http
GET /admin/stats/dashboard
Authorization: Bearer <admin-token>
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "totals": {
            "users": 1250,
            "chefs": 85,
            "meals": 450,
            "orders": 2340,
            "revenue": 45670.50
        },
        "growth": {
            "users_this_month": 124,
            "orders_this_month": 234,
            "revenue_this_month": 5670.25
        },
        "popular_meals": [
            {
                "id": 1,
                "name": "Ev Yapımı Mantı",
                "order_count": 145
            }
        ],
        "top_chefs": [
            {
                "id": 1,
                "business_name": "Mehmet Usta Mutfağı",
                "total_orders": 89,
                "average_rating": 4.8
            }
        ]
    }
}
```

## 🏥 Health Check

### System Health
```http
GET /health
```

**Response (200):**
```json
{
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.0.0",
    "database": "connected",
    "uptime": "72h30m45s"
}
```

### API Version
```http
GET /version
```

**Response (200):**
```json
{
    "version": "1.0.0",
    "build_date": "2024-01-01T10:00:00Z",
    "commit": "abc123def456",
    "go_version": "1.21.0"
}
```

## 📝 Error Responses

### Error Format

Tüm error response'lar aşağıdaki formatı takip eder:

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message",
        "details": {
            "field": "error details"
        }
    }
}
```

### Common Error Codes

#### Authentication Errors (401)
- `UNAUTHORIZED`: Token missing or invalid
- `TOKEN_EXPIRED`: JWT token has expired
- `INVALID_CREDENTIALS`: Wrong email/password

#### Authorization Errors (403)
- `FORBIDDEN`: Insufficient permissions
- `CHEF_ONLY`: Endpoint requires chef role
- `ADMIN_ONLY`: Endpoint requires admin role

#### Validation Errors (422)
- `VALIDATION_ERROR`: Request data validation failed
- `MISSING_FIELD`: Required field is missing
- `INVALID_FORMAT`: Field format is incorrect

#### Resource Errors (404)
- `NOT_FOUND`: Requested resource not found
- `USER_NOT_FOUND`: User does not exist
- `MEAL_NOT_FOUND`: Meal does not exist

#### Business Logic Errors (400)
- `MEAL_NOT_AVAILABLE`: Meal is not available for order
- `INSUFFICIENT_STOCK`: Not enough items in stock
- `ORDER_CANNOT_BE_CANCELLED`: Order cancellation not allowed

#### Server Errors (500)
- `INTERNAL_ERROR`: Internal server error
- `DATABASE_ERROR`: Database connection or query error

### Example Error Responses

**Validation Error (422):**
```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Request validation failed",
        "details": {
            "name": "Name is required",
            "price": "Price must be greater than 0"
        }
    }
}
```

**Unauthorized (401):**
```json
{
    "success": false,
    "error": {
        "code": "UNAUTHORIZED",
        "message": "Authentication required"
    }
}
```

## 📏 Rate Limiting

### Limits

- **General API**: 100 requests per minute per user
- **Authentication**: 5 requests per minute per IP
- **File Upload**: 10 requests per minute per user
- **Admin API**: 200 requests per minute per admin

### Headers

Rate limit bilgileri response header'larında döner:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 85
X-RateLimit-Reset: 1609459200
```

### Rate Limit Exceeded (429)

```json
{
    "success": false,
    "error": {
        "code": "RATE_LIMIT_EXCEEDED",
        "message": "Too many requests. Please try again later.",
        "details": {
            "retry_after": 60
        }
    }
}
```

## 🔄 Pagination

### Query Parameters

```http
GET /meals?page=2&limit=50
```

- `page`: Page number (starts from 1)
- `limit`: Items per page (default: 20, max: 100)

### Response Format

```json
{
    "success": true,
    "data": [...],
    "meta": {
        "page": 2,
        "limit": 50,
        "total": 500,
        "pages": 10,
        "has_prev": true,
        "has_next": true,
        "prev_page": 1,
        "next_page": 3
    }
}
```

## 📤 File Upload

### Image Upload

**Endpoint:**
```http
POST /meals/{meal_id}/image
Content-Type: multipart/form-data
Authorization: Bearer <token>
```

**Form Data:**
- `image`: Image file

**Constraints:**
- **Allowed formats**: JPEG, PNG, GIF
- **Max file size**: 5MB
- **Max dimensions**: 2048x2048px

**Response (200):**
```json
{
    "success": true,
    "data": {
        "url": "https://cdn.ozgurmutfak.com/images/meals/123/image.jpg",
        "size": 245760,
        "width": 800,
        "height": 600
    }
}
```

## 🌐 CORS

### Allowed Origins

- `https://ozgurmutfak.com`
- `https://www.ozgurmutfak.com`
- `https://admin.ozgurmutfak.com`
- `http://localhost:3000` (development)

### Allowed Methods

- `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`

### Allowed Headers

- `Content-Type`
- `Authorization`
- `X-Requested-With`

## 📱 API Versioning

### Version Header

```http
Accept: application/vnd.ozgurmutfak.v1+json
```

### URL Versioning (Current)

```
/api/v1/meals
/api/v1/orders
```

### Deprecation

Deprecated endpoints will include warning headers:

```http
Warning: 299 - "This API version is deprecated. Please migrate to v2."
Sunset: Wed, 31 Dec 2024 23:59:59 GMT
```

---

**Bu API dokümantasyonu düzenli olarak güncellenmektedir. Son güncellemeler için GitHub repository'sini kontrol ediniz.**
