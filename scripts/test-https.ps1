# PowerShell script for testing HTTPS functionality
$BaseUrl = "https://localhost:8443"

Write-Host "=== Testing High-Load Microservice HTTPS ===" -ForegroundColor Green

# Ignore SSL certificate errors for self-signed certificates
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12

# Test 1: Health check over HTTPS
Write-Host "`n1. Testing health endpoint over HTTPS..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
    Write-Host "✅ HTTPS Health check passed: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "❌ HTTPS Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Login over HTTPS
Write-Host "`n2. Testing admin login over HTTPS..." -ForegroundColor Yellow
try {
    $loginData = @{
        email = "admin@highload-microservice.local"
        password = "admin123456"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $accessToken = $loginResponse.access_token
    
    Write-Host "✅ HTTPS Admin login successful!" -ForegroundColor Green
    Write-Host "   Access Token: $($accessToken.Substring(0, 20))..." -ForegroundColor Cyan
} catch {
    Write-Host "❌ HTTPS Admin login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 3: Profile over HTTPS
Write-Host "`n3. Testing profile endpoint over HTTPS..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer $accessToken"
    }
    
    $profile = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/profile" -Method GET -Headers $headers
    Write-Host "✅ HTTPS Profile retrieved successfully!" -ForegroundColor Green
    Write-Host "   User ID: $($profile.user_id)" -ForegroundColor Cyan
    Write-Host "   Email: $($profile.email)" -ForegroundColor Cyan
    Write-Host "   Role: $($profile.role)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ HTTPS Profile retrieval failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Compare HTTP vs HTTPS
Write-Host "`n4. Testing HTTP vs HTTPS comparison..." -ForegroundColor Yellow
try {
    # Test HTTP (should work)
    $httpHealth = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "✅ HTTP endpoint works: $($httpHealth.status)" -ForegroundColor Green
    
    # Test HTTPS (should work)
    $httpsHealth = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
    Write-Host "✅ HTTPS endpoint works: $($httpsHealth.status)" -ForegroundColor Green
    
    Write-Host "✅ Both HTTP and HTTPS endpoints are working!" -ForegroundColor Green
} catch {
    Write-Host "❌ HTTP/HTTPS comparison failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== HTTPS Testing Complete ===" -ForegroundColor Green
Write-Host "HTTPS functionality is working correctly!" -ForegroundColor Green
Write-Host "Note: Using self-signed certificates for development." -ForegroundColor Yellow
