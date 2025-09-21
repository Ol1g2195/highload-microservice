param(
  [string]$BaseUrl = "http://localhost:8080"
)

Write-Host "Smoke start => $BaseUrl"

# health
$code = (Invoke-WebRequest -Uri "$BaseUrl/health" -Method GET -UseBasicParsing).StatusCode
if ($code -ne 200) { throw "Health failed: $code" }

# create user
$u = @{ email = "user@example.com"; first_name = "John"; last_name = "Doe" } | ConvertTo-Json
$created = Invoke-RestMethod -Uri "$BaseUrl/api/v1/users/" -Method POST -Body $u -ContentType "application/json"
if (-not $created.id) { throw "User create failed" }
Write-Host "User: $($created.id)"

# get user twice (cache warmup)
Invoke-RestMethod -Uri "$BaseUrl/api/v1/users/$($created.id)" -Method GET | Out-Null
Invoke-RestMethod -Uri "$BaseUrl/api/v1/users/$($created.id)" -Method GET | Out-Null

# create event
$e = @{ user_id = "$($created.id)"; type = "smoke_event"; data = '{"action":"login"}' } | ConvertTo-Json
Invoke-RestMethod -Uri "$BaseUrl/api/v1/events/" -Method POST -Body $e -ContentType "application/json" | Out-Null

# list events
$events = Invoke-RestMethod -Uri "$BaseUrl/api/v1/events?page=1&limit=10" -Method GET
if (-not $events.events) { throw "No events returned" }

Write-Host "Smoke OK"


