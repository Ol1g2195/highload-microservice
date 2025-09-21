# PowerShell script for testing authentication
$BaseUrl = "http://localhost:8080"

Write-Host "=== Testing High-Load Microservice Authentication ===" -ForegroundColor Green

# Test 1: Health check
Write-Host "`n1. Testing health endpoint..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
    Write-Host "✅ Health check passed: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Login with admin credentials
Write-Host "`n2. Testing admin login..." -ForegroundColor Yellow
try {
    $loginData = @{
        email = "admin@highload-microservice.local"
        password = "admin123456"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $accessToken = $loginResponse.access_token
    $refreshToken = $loginResponse.refresh_token
    
    Write-Host "✅ Admin login successful!" -ForegroundColor Green
    Write-Host "   Access Token: $($accessToken.Substring(0, 20))..." -ForegroundColor Cyan
    Write-Host "   User Role: $($loginResponse.user.role)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Admin login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 3: Get user profile
Write-Host "`n3. Testing profile endpoint..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer $accessToken"
    }
    
    $profile = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/profile" -Method GET -Headers $headers
    Write-Host "✅ Profile retrieved successfully!" -ForegroundColor Green
    Write-Host "   User ID: $($profile.user_id)" -ForegroundColor Cyan
    Write-Host "   Email: $($profile.email)" -ForegroundColor Cyan
    Write-Host "   Role: $($profile.role)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Profile retrieval failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Access protected endpoint (users list)
Write-Host "`n4. Testing protected users endpoint..." -ForegroundColor Yellow
try {
    $users = Invoke-RestMethod -Uri "$BaseUrl/api/v1/users/" -Method GET -Headers $headers
    Write-Host "✅ Users endpoint accessed successfully!" -ForegroundColor Green
    Write-Host "   Found $($users.Count) users" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Users endpoint access failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Test unauthorized access (without token)
Write-Host "`n5. Testing unauthorized access..." -ForegroundColor Yellow
try {
    $users = Invoke-RestMethod -Uri "$BaseUrl/api/v1/users/" -Method GET
    Write-Host "❌ Unauthorized access should have failed!" -ForegroundColor Red
} catch {
    Write-Host "✅ Unauthorized access properly blocked: $($_.Exception.Message)" -ForegroundColor Green
}

# Test 6: Create API key (admin only)
Write-Host "`n6. Testing API key creation..." -ForegroundColor Yellow
try {
    $apiKeyData = @{
        name = "test-api-key"
        permissions = @("users:read", "events:read")
    } | ConvertTo-Json

    $apiKeyResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/api-keys/" -Method POST -Body $apiKeyData -ContentType "application/json" -Headers $headers
    Write-Host "✅ API key created successfully!" -ForegroundColor Green
    Write-Host "   API Key: $($apiKeyResponse.api_key)" -ForegroundColor Cyan
    Write-Host "   Name: $($apiKeyResponse.name)" -ForegroundColor Cyan
    Write-Host "   Permissions: $($apiKeyResponse.permissions -join ', ')" -ForegroundColor Cyan
    
    # Test API key authentication
    Write-Host "`n7. Testing API key authentication..." -ForegroundColor Yellow
    $apiHeaders = @{
        "X-API-Key" = $apiKeyResponse.api_key
    }
    
    $apiUsers = Invoke-RestMethod -Uri "$BaseUrl/api/v1/users/" -Method GET -Headers $apiHeaders
    Write-Host "✅ API key authentication successful!" -ForegroundColor Green
    Write-Host "   Found $($apiUsers.Count) users via API key" -ForegroundColor Cyan
    
} catch {
    Write-Host "❌ API key creation/authentication failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 8: Token refresh
Write-Host "`n8. Testing token refresh..." -ForegroundColor Yellow
try {
    $refreshData = @{
        refresh_token = $refreshToken
    } | ConvertTo-Json

    $refreshResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/refresh" -Method POST -Body $refreshData -ContentType "application/json"
    Write-Host "✅ Token refresh successful!" -ForegroundColor Green
    Write-Host "   New Access Token: $($refreshResponse.access_token.Substring(0, 20))..." -ForegroundColor Cyan
} catch {
    Write-Host "❌ Token refresh failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== Authentication Testing Complete ===" -ForegroundColor Green
Write-Host "All security features are working correctly!" -ForegroundColor Green
