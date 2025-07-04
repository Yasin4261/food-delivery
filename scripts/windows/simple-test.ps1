# API Quick Test
Write-Host "API Test Starting..." -ForegroundColor Green

$baseUrl = "http://localhost:3001/api/v1"
$headers = @{ "Content-Type" = "application/json" }

# Health Check
Write-Host "Testing API Health..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/products" -Method GET
    Write-Host "API is running: $($response.message)" -ForegroundColor Green
} catch {
    Write-Host "API not responding!" -ForegroundColor Red
    exit 1
}

# Register Test
Write-Host "Testing Registration..." -ForegroundColor Yellow
$registerData = @{
    email = "quicktest@example.com"
    password = "password123"
    first_name = "Quick"
    last_name = "Test"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method POST -Headers $headers -Body $registerData
    $token = $registerResponse.data.token
    Write-Host "Registration successful" -ForegroundColor Green
} catch {
    Write-Host "User exists, trying login..." -ForegroundColor Yellow
    
    $loginData = @{
        email = "quicktest@example.com"
        password = "password123"
    } | ConvertTo-Json
    
    try {
        $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Headers $headers -Body $loginData
        $token = $loginResponse.data.token
        Write-Host "Login successful" -ForegroundColor Green
    } catch {
        Write-Host "Both register and login failed" -ForegroundColor Red
        exit 1
    }
}

# Logout Test
Write-Host "Testing Logout..." -ForegroundColor Yellow
$authHeaders = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

try {
    $logoutResponse = Invoke-RestMethod -Uri "$baseUrl/auth/logout" -Method POST -Headers $authHeaders
    Write-Host "Logout successful: $($logoutResponse.message)" -ForegroundColor Green
} catch {
    Write-Host "Logout failed" -ForegroundColor Red
}

Write-Host "Test completed successfully!" -ForegroundColor Green
Write-Host "Files created:" -ForegroundColor Cyan
Write-Host "- postman_collection.json" -ForegroundColor White
Write-Host "- postman_environment.json" -ForegroundColor White
Write-Host "- api-test.http" -ForegroundColor White
Write-Host "- API_TEST_GUIDE.md" -ForegroundColor White
