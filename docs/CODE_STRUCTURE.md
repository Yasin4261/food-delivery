# 🏗️ Code Structure Documentation

## Overview

Bu dokümantasyon, Özgür Mutfak API projesinin kod yapısını ve mimari kararlarını açıklar.

## 📁 Project Structure

```
Özgür Mutfak/
├── cmd/                          # Application entry points
│   └── main.go                   # Main application with Swagger setup
├── internal/                     # Private application code
│   ├── api/                      # HTTP layer
│   │   ├── handler/              # HTTP request handlers
│   │   ├── middleware/           # HTTP middlewares
│   │   └── router.go             # Route definitions
│   ├── auth/                     # Authentication logic
│   │   ├── auth.go               # JWT manager
│   │   └── auth_test.go          # Auth tests
│   ├── model/                    # Data models
│   │   ├── *.go                  # Business models
│   │   └── *_test.go             # Model tests
│   ├── repository/               # Data access layer
│   │   ├── *_repository.go       # Database operations
│   │   └── *_repository_test.go  # Repository tests
│   └── service/                  # Business logic layer
│       ├── *_service.go          # Business logic
│       └── *_service_test.go     # Service tests
├── config/                       # Configuration
│   ├── config.go                 # Config structure
│   └── config.docker.yaml       # Docker config
├── migrations/                   # Database migrations
│   └── *.sql                     # SQL migration files
├── docs/                         # Documentation
├── tests/                        # Integration tests
├── pkg/                          # Public packages
├── scripts/                      # Build/deployment scripts
├── docker-compose.yml            # Docker services
├── Dockerfile                    # Container definition
├── go.mod                        # Go modules
└── README.md                     # Project documentation
```

## 🏛️ Architecture Pattern

### Clean Architecture

Proje Clean Architecture prensiplerine uygun olarak tasarlanmıştır:

```
┌─────────────────────────────────────────┐
│               Frameworks                │  ← HTTP, Database, External APIs
├─────────────────────────────────────────┤
│            Interface Adapters           │  ← Handlers, Repositories
├─────────────────────────────────────────┤
│             Use Cases                   │  ← Services (Business Logic)
├─────────────────────────────────────────┤
│               Entities                  │  ← Models, Domain Objects
└─────────────────────────────────────────┘
```

### Dependency Flow

```
Handler → Service → Repository → Database
   ↓        ↓         ↓
 HTTP    Business   Data
Layer    Logic     Access
```

## 📋 Layer Responsibilities

### 1. Handler Layer (`internal/api/handler/`)

**Purpose**: HTTP request/response handling
- HTTP request parsing
- Input validation
- Response formatting
- Error handling
- Authentication checks

**Dependencies**: 
- Services (business logic)
- Models (data structures)

```go
// Example handler function
func GetMeals(c *gin.Context) {
    // 1. Parse request parameters
    // 2. Call service layer
    // 3. Format response
    // 4. Handle errors
}
```

### 2. Service Layer (`internal/service/`)

**Purpose**: Business logic implementation
- Core business rules
- Data validation
- Cross-entity operations
- Transaction coordination

**Dependencies**:
- Repositories (data access)
- Models (data structures)
- External services

```go
type MealService struct {
    mealRepo repository.MealRepository
    chefRepo repository.ChefRepository
}

func (s *MealService) CreateMeal(meal *model.Meal) error {
    // 1. Validate business rules
    // 2. Check permissions
    // 3. Save to database
    // 4. Handle side effects
}
```

### 3. Repository Layer (`internal/repository/`)

**Purpose**: Data access abstraction
- Database operations
- Query building
- Data mapping
- Connection management

**Dependencies**:
- Database driver
- Models (for mapping)

```go
type MealRepository interface {
    Create(meal *model.Meal) error
    GetByID(id int) (*model.Meal, error)
    Update(meal *model.Meal) error
    Delete(id int) error
}
```

### 4. Model Layer (`internal/model/`)

**Purpose**: Data structure definitions
- Business entities
- Data transfer objects
- Validation rules
- JSON serialization

```go
type Meal struct {
    ID          int     `json:"id" db:"id"`
    Name        string  `json:"name" db:"name" validate:"required"`
    Price       float64 `json:"price" db:"price" validate:"min=0"`
    ChefID      int     `json:"chef_id" db:"chef_id"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

## 🔧 Key Design Patterns

### 1. Repository Pattern

**Purpose**: Abstract data access logic

```go
// Interface definition
type UserRepository interface {
    Create(user *model.User) error
    GetByEmail(email string) (*model.User, error)
    Update(user *model.User) error
}

// Implementation
type userRepository struct {
    db *sql.DB
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
    // SQL query implementation
}
```

### 2. Dependency Injection

**Purpose**: Loose coupling between components

```go
// Service receives dependencies via constructor
func NewMealService(mealRepo repository.MealRepository, chefRepo repository.ChefRepository) service.MealService {
    return &mealService{
        mealRepo: mealRepo,
        chefRepo: chefRepo,
    }
}
```

### 3. Middleware Pattern

**Purpose**: Cross-cutting concerns

```go
func AuthRequired(jwtManager *auth.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token
        // 2. Validate token
        // 3. Set user context
        // 4. Continue or abort
    }
}
```

## 🔐 Authentication Architecture

### JWT Token Flow

```
1. User Login → Handler
2. Handler → AuthService.Login()
3. AuthService → UserRepository.GetByEmail()
4. Validate password (bcrypt)
5. Generate JWT token
6. Return token to client
7. Client includes token in Authorization header
8. Middleware validates token on protected routes
```

### Role-Based Access Control

```go
// Roles hierarchy
const (
    RoleCustomer = "customer"
    RoleChef     = "chef"
    RoleAdmin    = "admin"
)

// Middleware checks
func RoleRequired(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if userRole != role {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 📊 Database Schema Design

### Table Relationships

```
Users (1) ←→ (0..1) Chefs
  ↓
  (1) ←→ (N) Orders
  ↓
  (1) ←→ (N) CartItems

Chefs (1) ←→ (N) Meals
  ↓
  (1) ←→ (N) Reviews

Orders (1) ←→ (N) OrderItems
OrderItems (N) ←→ (1) Meals
```

### Migration Strategy

1. **Sequential numbering**: `001_`, `002_`, etc.
2. **Descriptive names**: `001_initial_schema.sql`
3. **Up migrations only**: Forward compatibility
4. **Idempotent scripts**: Can be run multiple times

## 🧪 Testing Strategy

### Test Pyramid

```
                🔺
               /E2E\          ← Integration tests (few)
              /-----\
             /Unit  \         ← Unit tests (many)
            /_______\
```

### Test Categories

1. **Unit Tests**: Test individual functions/methods
   - Model validation
   - Service business logic
   - Repository data access

2. **Integration Tests**: Test component interactions
   - API endpoints
   - Database operations
   - External service calls

3. **Mock Strategy**: Use mocks for external dependencies
   ```go
   type MockUserRepository struct{}
   
   func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
       // Mock implementation
   }
   ```

## 🔄 Error Handling Strategy

### Error Types

1. **Business Logic Errors**: Validation, business rules
2. **Data Access Errors**: Database, network issues
3. **Authentication Errors**: Invalid tokens, permissions
4. **System Errors**: Internal server errors

### Error Response Format

```go
type ErrorResponse struct {
    Error struct {
        Code    string `json:"code"`
        Message string `json:"message"`
        Details []struct {
            Field   string `json:"field"`
            Message string `json:"message"`
        } `json:"details,omitempty"`
    } `json:"error"`
}
```

## 🚀 Deployment Architecture

### Docker Setup

```yaml
# docker-compose.yml
services:
  api:
    build: .
    ports:
      - "3001:8080"
    depends_on:
      - db
    environment:
      - DB_HOST=db
      
  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=ecommerce
```

### Environment Configuration

```go
type Config struct {
    Server struct {
        Port string `yaml:"port"`
        Host string `yaml:"host"`
    } `yaml:"server"`
    
    Database struct {
        Host     string `yaml:"host"`
        Port     string `yaml:"port"`
        Name     string `yaml:"name"`
        User     string `yaml:"user"`
        Password string `yaml:"password"`
    } `yaml:"database"`
}
```

## 📈 Performance Considerations

### Database Optimization

1. **Indexing Strategy**:
   - Primary keys: Auto-indexed
   - Foreign keys: Indexed for joins
   - Search fields: Composite indexes

2. **Connection Pooling**:
   ```go
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

3. **Query Optimization**:
   - Use prepared statements
   - Avoid N+1 queries
   - Implement pagination

### API Performance

1. **Response Caching**: HTTP cache headers
2. **JSON Optimization**: Minimal payloads
3. **Compression**: Gzip middleware
4. **Rate Limiting**: Prevent abuse

## 🔍 Monitoring & Observability

### Logging Strategy

```go
// Structured logging
log.WithFields(log.Fields{
    "user_id": userID,
    "action":  "create_order",
    "order_id": orderID,
}).Info("Order created successfully")
```

### Health Checks

```go
// Health check endpoint
func HealthCheck(c *gin.Context) {
    // Check database connection
    // Check external services
    // Return status
}
```

## 🔄 Development Workflow

### Git Workflow

1. **Feature branches**: `feature/meal-management`
2. **Commit convention**: `feat: add meal creation endpoint`
3. **Pull requests**: Required for main branch
4. **Testing**: All tests must pass

### Code Quality

1. **Linting**: `golangci-lint`
2. **Formatting**: `gofmt`
3. **Testing**: Minimum 80% coverage
4. **Documentation**: Godoc comments

## 📋 Future Considerations

### Scalability

1. **Microservices**: Split by domain
2. **Caching**: Redis integration
3. **Message Queues**: Async processing
4. **Load Balancing**: Multiple API instances

### Security Enhancements

1. **Rate limiting**: API throttling
2. **Input sanitization**: SQL injection prevention
3. **HTTPS only**: TLS enforcement
4. **Security headers**: CORS, CSP

---

**Note**: Bu dokümantasyon proje gelişimi ile birlikte güncellenmelidir.
