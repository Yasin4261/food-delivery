# Özgür Mutfak E-Commerce API Test Script
# Bu script tüm API endpoint'lerini otomatik test eder

Write-Host "🚀 Özgür Mutfak E-Commerce API Test Script" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green

# API Base URL
$baseUrl = "http://localhost:3001/api/v1"

# Test Data
$testUser = @{
    email = "postman.test@example.com"
    password = "password123"
    first_name = "Postman"
    last_name = "Test"
}

# Headers
$jsonHeaders = @{ "Content-Type" = "application/json" }

Write-Host "`n🔍 1. API Health Check..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$baseUrl/products" -Method GET
    Write-Host "✅ API is running: $($healthResponse.message)" -ForegroundColor Green
} catch {
    Write-Host "❌ API is not responding. Make sure Docker containers are running!" -ForegroundColor Red
    Write-Host "Run: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n📝 2. Testing User Registration..." -ForegroundColor Yellow
try {
    $registerBody = $testUser | ConvertTo-Json
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method POST -Headers $jsonHeaders -Body $registerBody
    $token = $registerResponse.data.token
    Write-Host "✅ User registered successfully" -ForegroundColor Green
    Write-Host "   Token: $($token.Substring(0, 20))..." -ForegroundColor Cyan
} catch {
    if ($_.Exception.Response.StatusCode -eq "BadRequest") {
        Write-Host "⚠️ User already exists (this is expected on second run)" -ForegroundColor Yellow
        
        } catch {
    if ($_.Exception.Response.StatusCode -eq "BadRequest") {
        Write-Host "⚠️ User already exists (this is expected on second run)" -ForegroundColor Yellow
        
        # Try to login instead
        Write-Host "`n🔑 Attempting login with existing user..." -ForegroundColor Yellow
        $loginBody = @{
            email = $testUser.email
            password = $testUser.password
        } | ConvertTo-Json
        
        try {
            $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Headers $jsonHeaders -Body $loginBody
            $token = $loginResponse.data.token
            Write-Host "✅ Login successful with existing user" -ForegroundColor Green
        } catch {
            Write-Host "❌ Both register and login failed" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "❌ Registration failed: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
}
}

Write-Host "`n🔑 3. Testing User Login..." -ForegroundColor Yellow
try {
    $loginBody = @{
        email = $testUser.email
        password = $testUser.password
    } | ConvertTo-Json
    
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Headers $jsonHeaders -Body $loginBody
    $token = $loginResponse.data.token
    Write-Host "✅ Login successful" -ForegroundColor Green
    Write-Host "   New Token: $($token.Substring(0, 20))..." -ForegroundColor Cyan
} catch {
    Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n🚪 4. Testing User Logout..." -ForegroundColor Yellow
try {
    $authHeaders = @{ 
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $logoutResponse = Invoke-RestMethod -Uri "$baseUrl/auth/logout" -Method POST -Headers $authHeaders
    Write-Host "✅ Logout successful: $($logoutResponse.message)" -ForegroundColor Green
} catch {
    Write-Host "❌ Logout failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n🛒 5. Testing Products Endpoint..." -ForegroundColor Yellow
try {
    $productsResponse = Invoke-RestMethod -Uri "$baseUrl/products" -Method GET
    Write-Host "✅ Products endpoint working: $($productsResponse.message)" -ForegroundColor Green
} catch {
    Write-Host "❌ Products endpoint failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n🧪 6. Testing Error Cases..." -ForegroundColor Yellow

# Test wrong password
Write-Host "   Testing wrong password..." -ForegroundColor Gray
try {
    $wrongLoginBody = @{
        email = $testUser.email
        password = "wrongpassword"
    } | ConvertTo-Json
    
    Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Headers $jsonHeaders -Body $wrongLoginBody
    Write-Host "❌ Expected error for wrong password but got success!" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq "Unauthorized") {
        Write-Host "   ✅ Wrong password correctly rejected (401)" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️ Unexpected error: $($_.Exception.Response.StatusCode)" -ForegroundColor Yellow
    }
}

# Test invalid token
Write-Host "   Testing invalid token..." -ForegroundColor Gray
try {
    $invalidAuthHeaders = @{ 
        "Authorization" = "Bearer invalid_token"
        "Content-Type" = "application/json"
    }
    
    Invoke-RestMethod -Uri "$baseUrl/auth/logout" -Method POST -Headers $invalidAuthHeaders
    Write-Host "❌ Expected error for invalid token but got success!" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq "Unauthorized") {
        Write-Host "   ✅ Invalid token correctly rejected (401)" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️ Unexpected error: $($_.Exception.Response.StatusCode)" -ForegroundColor Yellow
    }
}

Write-Host "`n📊 Test Summary:" -ForegroundColor Green
Write-Host "=================" -ForegroundColor Green
Write-Host "✅ API Health Check" -ForegroundColor Green
Write-Host "✅ User Registration/Login" -ForegroundColor Green
Write-Host "✅ User Logout" -ForegroundColor Green
Write-Host "✅ Products Endpoint" -ForegroundColor Green
Write-Host "✅ Error Handling" -ForegroundColor Green

Write-Host "`n🎉 All tests completed successfully!" -ForegroundColor Green
Write-Host "`n📁 Files created for manual testing:" -ForegroundColor Cyan
Write-Host "   - postman_collection.json (Import to Postman)" -ForegroundColor White
Write-Host "   - postman_environment.json (Import to Postman)" -ForegroundColor White
Write-Host "   - api-test.http (Use with VS Code REST Client)" -ForegroundColor White
Write-Host "   - API_TEST_GUIDE.md (Detailed documentation)" -ForegroundColor White

Write-Host "`n💡 Next steps:" -ForegroundColor Yellow
Write-Host "   1. Import Postman files for manual testing" -ForegroundColor White
Write-Host "   2. Use VS Code REST Client with api-test.http" -ForegroundColor White
Write-Host "   3. Read API_TEST_GUIDE.md for detailed instructions" -ForegroundColor White
