$loginRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/auth/login' -Method POST -ContentType 'application/json' -Body '{"username":"admin","password":"password123"}'
$tok = $loginRes.data.token
$hdrs = @{Authorization="Bearer $tok"}

Write-Host "=== SEED ==="
$seedRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/admin/seed' -Method POST -ContentType 'application/json' -Headers $hdrs -Body '{}'
$seedRes | ConvertTo-Json

Write-Host "=== DS OVERVIEW ==="
$dsRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/analytics/ds/overview' -Headers $hdrs
$dsRes | ConvertTo-Json -Depth 5

Write-Host "=== DS MODELS ==="
$modRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/analytics/ds/models' -Headers $hdrs
$modRes | ConvertTo-Json -Depth 5

Write-Host "=== FACTORS ==="
$facRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/factors' -Headers $hdrs
$facRes | ConvertTo-Json -Depth 3

Write-Host "=== PERFORMANCE ==="
$perfRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/analytics/performance' -Headers $hdrs
$perfRes | ConvertTo-Json -Depth 3
