# 🤝 Contributing Guidelines

## Welcome

Özgür Mutfak projesine katkıda bulunduğunuz için teşekkür ederiz! Bu döküman, projeye nasıl katkı yapabileceğinizi açıklar.

## 📋 Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Workflow](#development-workflow)
4. [Coding Standards](#coding-standards)
5. [Commit Guidelines](#commit-guidelines)
6. [Pull Request Process](#pull-request-process)
7. [Testing Requirements](#testing-requirements)
8. [Documentation](#documentation)
9. [Issue Reporting](#issue-reporting)
10. [Community](#community)

## 🤝 Code of Conduct

### Our Pledge

Özgür Mutfak topluluğu olarak, yaş, vücut ölçüsü, engellilik, etnik köken, cinsiyet kimliği ve ifadesi, deneyim seviyesi, milliyetiçelik, kişisel görünüm, ırk, din veya cinsel kimlik ve yönelim fark etmeksizin herkes için açık ve misafirperver bir ortam yaratmayı taahhüt ediyoruz.

### Our Standards

Pozitif bir ortam yaratmaya katkıda bulunan davranış örnekleri:

- ✅ Misafirperver ve kapsayıcı dil kullanmak
- ✅ Farklı bakış açılarına ve deneyimlere saygı göstermek
- ✅ Yapıcı eleştirileri nezaketle kabul etmek
- ✅ Topluluk için en iyisine odaklanmak
- ✅ Diğer topluluk üyelerine empati göstermek

Kabul edilemez davranışlar:

- ❌ Cinselleştirilmiş dil veya görüntü kullanımı
- ❌ Trolleme, hakaret edici/aşağılayıcı yorumlar ve kişisel veya politik saldırılar
- ❌ Açık veya özel taciz
- ❌ Başkalarının fiziksel veya e-posta adresi gibi özel bilgilerini izinsiz yayınlamak
- ❌ Profesyonel ortamda makul sayılamayacak diğer davranışlar

## 🚀 Getting Started

### Prerequisites

1. **Development Environment**
   - Go 1.21+ yüklü
   - PostgreSQL 15+ yüklü
   - Git yapılandırılmış
   - VS Code (önerilen) + Go extension

2. **Hesap Requirements**
   - GitHub hesabı
   - Git global config ayarlanmış:
     ```bash
     git config --global user.name "Your Name"
     git config --global user.email "your.email@example.com"
     ```

### First Time Setup

```bash
# 1. Repository'yi fork edin (GitHub web arayüzünde)

# 2. Fork'u clone edin
git clone https://github.com/YOUR_USERNAME/ozgur-mutfak.git
cd ozgur-mutfak

# 3. Original repository'yi upstream olarak ekleyin
git remote add upstream https://github.com/ORIGINAL_OWNER/ozgur-mutfak.git

# 4. Development setup'ı tamamlayın
cp .env.example .env.development
# .env.development dosyasını düzenleyin

# 5. Dependencies'leri yükleyin
go mod download

# 6. Database setup
createdb ozgur_mutfak_dev
goose -dir migrations postgres "user=postgres dbname=ozgur_mutfak_dev sslmode=disable" up

# 7. Tests'in çalıştığını doğrulayın
go test ./...
```

## 🔄 Development Workflow

### Branch Strategy

Projede **GitFlow** benzeri bir workflow kullanıyoruz:

```
main (production)
├── develop (development)
│   ├── feature/meal-management
│   ├── feature/user-authentication
│   ├── bugfix/order-calculation
│   └── hotfix/security-patch
```

### Working on Features

```bash
# 1. Latest changes'leri pull edin
git checkout develop
git pull upstream develop

# 2. Feature branch oluşturun
git checkout -b feature/your-feature-name

# 3. Changes'lerinizi yapın
# ... kod yazın ...

# 4. Regular commits yapın
git add .
git commit -m "feat: add user registration endpoint"

# 5. Upstream'den updates alın (gerekirse)
git fetch upstream
git rebase upstream/develop

# 6. Push to your fork
git push origin feature/your-feature-name

# 7. Pull Request oluşturun
```

### Branch Naming Convention

- **Features**: `feature/short-description`
  - `feature/meal-crud-operations`
  - `feature/jwt-authentication`
  - `feature/order-status-tracking`

- **Bug Fixes**: `bugfix/short-description`
  - `bugfix/order-total-calculation`
  - `bugfix/email-validation`

- **Hot Fixes**: `hotfix/short-description`
  - `hotfix/security-vulnerability`
  - `hotfix/database-connection`

- **Documentation**: `docs/short-description`
  - `docs/api-documentation`
  - `docs/deployment-guide`

## 📝 Coding Standards

### Go Code Style

#### 1. Formatting

```bash
# Her commit'ten önce çalıştırın
go fmt ./...
goimports -w .
```

#### 2. Naming Conventions

```go
// ✅ Good - PascalCase for exported functions
func CreateMeal(meal *model.Meal) error

// ✅ Good - camelCase for unexported functions
func calculateOrderTotal(items []OrderItem) float64

// ✅ Good - Interface naming
type MealRepository interface {
    Create(meal *model.Meal) error
    GetByID(id int) (*model.Meal, error)
}

// ✅ Good - Struct naming
type MealService struct {
    repo repository.MealRepository
}

// ❌ Bad - Unclear naming
func DoStuff() error
func Process(x interface{}) interface{}
```

#### 3. Error Handling

```go
// ✅ Good - Proper error handling
func GetMealByID(id int) (*model.Meal, error) {
    meal, err := repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get meal with id %d: %w", id, err)
    }
    return meal, nil
}

// ❌ Bad - Ignoring errors
func GetMealByID(id int) *model.Meal {
    meal, _ := repo.GetByID(id) // Don't ignore errors!
    return meal
}
```

#### 4. Comments and Documentation

```go
// ✅ Good - Package documentation
// Package service implements business logic for the Özgür Mutfak application.
// It provides meal management, order processing, and user authentication services.
package service

// ✅ Good - Function documentation
// CreateMeal creates a new meal in the system.
// It validates the meal data and saves it to the database.
// Returns the created meal with assigned ID or an error if validation fails.
func CreateMeal(meal *model.Meal) (*model.Meal, error) {
    // Implementation...
}

// ✅ Good - Complex logic explanation
// calculateDeliveryFee determines the delivery fee based on:
// 1. Distance from chef to customer
// 2. Time of day (peak hours have higher fees)
// 3. Order value (free delivery over threshold)
func calculateDeliveryFee(distance float64, orderTime time.Time, orderValue float64) float64 {
    // Implementation...
}
```

### API Design Standards

#### 1. HTTP Methods

```go
// ✅ Correct HTTP method usage
// GET /api/v1/meals          - List meals
// GET /api/v1/meals/:id      - Get specific meal
// POST /api/v1/meals         - Create new meal
// PUT /api/v1/meals/:id      - Update entire meal
// PATCH /api/v1/meals/:id    - Partial update
// DELETE /api/v1/meals/:id   - Delete meal
```

#### 2. Response Format

```go
// ✅ Consistent response structure
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Details map[string]string `json:"details,omitempty"`
}

// ✅ Successful response
{
    "success": true,
    "data": {
        "id": 1,
        "name": "Mantı",
        "price": 25.00
    }
}

// ✅ Error response
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid input data",
        "details": {
            "name": "Name is required",
            "price": "Price must be positive"
        }
    }
}
```

#### 3. Status Codes

```go
// ✅ Appropriate status codes
// 200 - OK (successful GET, PUT, PATCH)
// 201 - Created (successful POST)
// 204 - No Content (successful DELETE)
// 400 - Bad Request (client error)
// 401 - Unauthorized (authentication required)
// 403 - Forbidden (insufficient permissions)
// 404 - Not Found (resource doesn't exist)
// 409 - Conflict (duplicate resource)
// 422 - Unprocessable Entity (validation error)
// 500 - Internal Server Error (server error)
```

### Database Standards

#### 1. Migration Naming

```sql
-- ✅ Good migration naming
-- 001_create_users_table.sql
-- 002_add_email_index_to_users.sql
-- 003_create_meals_table.sql
-- 004_add_chef_verification_status.sql
```

#### 2. Table and Column Naming

```sql
-- ✅ Good naming conventions
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ✅ Good index naming
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_orders_user_id_status ON orders(user_id, status);
```

## 📝 Commit Guidelines

### Commit Message Format

Conventional Commits formatını kullanıyoruz:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Types

- **feat**: Yeni feature ekleme
- **fix**: Bug fix
- **docs**: Sadece documentation değişiklikleri
- **style**: Code formatting, missing semi colons, etc.
- **refactor**: Code refactoring (ne feature ne bug fix)
- **perf**: Performance improvements
- **test**: Test ekleme veya düzeltme
- **chore**: Build process, dependency updates, etc.

#### Examples

```bash
# ✅ Good commits
feat: add user registration endpoint
fix: resolve order total calculation error
docs: update API documentation for meals endpoint
refactor: extract user validation logic to separate function
test: add unit tests for meal service
chore: update Go dependencies

# ✅ With scope
feat(auth): implement JWT token refresh
fix(orders): correct status transition validation
docs(api): add examples to Swagger documentation

# ✅ With body
feat: add meal image upload functionality

Implement image upload for meal listings with:
- File validation (type, size)
- Image compression and resizing
- S3 storage integration
- Proper error handling

Closes #123

# ❌ Bad commits
Update stuff
Fix bug
WIP
asdfsadf
Fixed it finally!!!
```

### Commit Best Practices

1. **Make atomic commits** - Her commit tek bir logical change içermeli
2. **Write meaningful messages** - Commit'in ne yaptığını açık olarak belirtin
3. **Use imperative mood** - "Add feature" not "Added feature"
4. **Limit first line to 50 characters**
5. **Provide context in body** - Gerekirse açıklama ekleyin

## 🔄 Pull Request Process

### Before Creating PR

```bash
# 1. Code quality checks
go fmt ./...
goimports -w .
golangci-lint run

# 2. Run all tests
go test ./... -race -cover

# 3. Update documentation (if needed)
swag init -g cmd/main.go -o docs/

# 4. Rebase from develop
git fetch upstream
git rebase upstream/develop

# 5. Push to your fork
git push origin feature/your-feature-name
```

### PR Template

Pull Request oluştururken şu template'i kullanın:

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

## Screenshots (if applicable)
Add screenshots here if the change includes UI modifications.

## Related Issues
Closes #issue_number
```

### PR Review Process

1. **Automated Checks**
   - ✅ All tests must pass
   - ✅ Code coverage maintained (>80%)
   - ✅ Linting checks pass
   - ✅ Security scans pass

2. **Manual Review**
   - Code quality ve best practices
   - Business logic correctness
   - Performance implications
   - Security considerations

3. **Approval Requirements**
   - At least 1 approval from maintainer
   - All conversations resolved
   - No conflicting files

## 🧪 Testing Requirements

### Test Coverage

- **Minimum**: 80% code coverage
- **Target**: 90% code coverage
- **Critical paths**: 100% coverage (auth, payments, orders)

### Test Types

#### 1. Unit Tests

```go
// ✅ Good unit test
func TestMealService_CreateMeal(t *testing.T) {
    tests := []struct {
        name    string
        meal    *model.Meal
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid meal creation",
            meal: &model.Meal{
                Name:        "Test Meal",
                Description: "Test Description",
                Price:       25.00,
                ChefID:      1,
            },
            wantErr: false,
        },
        {
            name: "missing name should fail",
            meal: &model.Meal{
                Description: "Test Description",
                Price:       25.00,
                ChefID:      1,
            },
            wantErr: true,
            errMsg:  "name is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewMealService(mockRepo)
            _, err := service.CreateMeal(tt.meal)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateMeal() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("CreateMeal() error = %v, expected to contain %v", err, tt.errMsg)
            }
        })
    }
}
```

#### 2. Integration Tests

```go
// ✅ Good integration test
func TestMealAPI_CreateMeal_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    router := setupTestRouter(db)
    
    meal := map[string]interface{}{
        "name":        "Integration Test Meal",
        "description": "Test Description",
        "price":       25.00,
        "chef_id":     1,
    }
    
    jsonData, _ := json.Marshal(meal)
    req := httptest.NewRequest("POST", "/api/v1/meals", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+getTestJWT())
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
    
    var response APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response.Success)
}
```

### Test Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/service/...

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

## 📚 Documentation

### Code Documentation

1. **Package Documentation**
   ```go
   // Package service provides business logic for meal management.
   // It includes meal creation, validation, and retrieval operations.
   package service
   ```

2. **Function Documentation**
   ```go
   // CreateMeal creates a new meal with validation.
   // It returns the created meal with assigned ID or validation error.
   func CreateMeal(meal *model.Meal) (*model.Meal, error)
   ```

3. **Complex Logic**
   ```go
   // Calculate delivery fee based on multiple factors:
   // - Base fee: $2.99
   // - Distance multiplier: $0.50 per km
   // - Peak hour surcharge: 20% (6-9 PM)
   // - Free delivery: orders over $50
   func calculateDeliveryFee(distance float64, orderTime time.Time, total float64) float64
   ```

### API Documentation

Swagger/OpenAPI 3.0 kullanıyoruz:

```go
// @Summary Create a new meal
// @Description Create a new meal with the provided data
// @Tags meals
// @Accept json
// @Produce json
// @Param meal body model.CreateMealRequest true "Meal data"
// @Success 201 {object} APIResponse{data=model.Meal}
// @Failure 400 {object} APIResponse{error=APIError}
// @Failure 401 {object} APIResponse{error=APIError}
// @Security BearerToken
// @Router /meals [post]
func CreateMeal(c *gin.Context) {
    // Implementation...
}
```

### README Updates

Eğer major feature ekliyorsanız, README.md'yi güncelleyin:

- Features listesi
- API endpoints
- Setup instructions (gerekirse)

## 🐛 Issue Reporting

### Bug Reports

Bug bulduğunuzda şu template'i kullanın:

```markdown
**Bug Description**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected Behavior**
A clear and concise description of what you expected to happen.

**Actual Behavior**
What actually happened.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**
- OS: [e.g. Ubuntu 20.04]
- Go version: [e.g. 1.21.0]
- PostgreSQL version: [e.g. 15.3]

**Additional Context**
Add any other context about the problem here.

**Logs**
```
Include relevant log output here
```
```

### Feature Requests

Yeni feature önerisi için:

```markdown
**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is.

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.

**Implementation Ideas**
If you have ideas about how this could be implemented, please share them.
```

## 👥 Community

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: General discussions, questions
- **Email**: ozgur.mutfak.dev@gmail.com (maintainers)

### Getting Help

1. **Documentation**: Önce docs/ klasöründeki dökümanları kontrol edin
2. **Search Issues**: Mevcut issues'larda similar problems arayın
3. **GitHub Discussions**: Genel sorular için discussions kullanın
4. **Create Issue**: Spesifik bug/feature için yeni issue açın

### Recognition

Katkılarınız şu şekillerde tanınacak:

1. **Contributors List**: README.md'de contributor olarak listelenme
2. **Release Notes**: Major contributions için release notes'ta bahsedilme
3. **Recognition Badge**: GitHub profile'ınızda gösterilebilecek badge

## 📋 Maintainer Responsibilities

### For Maintainers

1. **Code Review Standards**
   - Functionality correctness
   - Code quality and style
   - Performance implications
   - Security considerations
   - Test coverage

2. **Response Times**
   - Issues: 48 hours içinde response
   - PRs: 72 hours içinde initial review
   - Bug fixes: Priority'ye göre escalation

3. **Release Management**
   - Semantic versioning
   - Change log maintenance
   - Backward compatibility
   - Migration guides

### Becoming a Maintainer

Maintainer olmak için:

1. **Consistent Contributions**: 3+ months regular contributions
2. **Code Quality**: High-quality code ve best practices
3. **Community Involvement**: Issues, reviews, discussions'larda aktif katılım
4. **Technical Knowledge**: Codebase'in tamamına deep understanding

## 🏷️ Labels and Milestones

### Issue Labels

- **Type**:
  - `bug` - Something isn't working
  - `enhancement` - New feature or request
  - `documentation` - Improvements or additions to documentation
  - `question` - Further information is requested

- **Priority**:
  - `priority: critical` - Critical bug, security issue
  - `priority: high` - Important feature or serious bug
  - `priority: medium` - Standard priority
  - `priority: low` - Nice to have

- **Status**:
  - `status: needs-triage` - Needs maintainer review
  - `status: accepted` - Approved for development
  - `status: in-progress` - Currently being worked on
  - `status: blocked` - Blocked by external dependency

- **Difficulty**:
  - `good first issue` - Good for newcomers
  - `help wanted` - Extra attention is needed
  - `difficulty: easy` - Can be completed quickly
  - `difficulty: medium` - Requires moderate effort
  - `difficulty: hard` - Complex implementation

## ✅ Checklist for Contributors

### Before Starting Work

- [ ] Issue exists and is approved
- [ ] Development environment setup
- [ ] Tests are passing
- [ ] Dependencies are up to date

### During Development

- [ ] Follow coding standards
- [ ] Write/update tests
- [ ] Update documentation
- [ ] Regular commits with good messages

### Before Submitting PR

- [ ] All tests pass locally
- [ ] Code is formatted and linted
- [ ] Documentation updated
- [ ] Self-reviewed code
- [ ] Rebased from develop branch

---

**Thank you for contributing to Özgür Mutfak! 🎉**

Bu guidelines'a uyarak, code quality'yi yüksek tutmaya ve pozitif bir development experience sağlamaya yardımcı oluyorsunuz.
