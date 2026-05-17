$loginRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/auth/login' -Method POST -ContentType 'application/json' -Body '{"username":"admin","password":"password123"}'
$tok = $loginRes.data.token
$hdrs = @{Authorization="Bearer $tok"}

Write-Host "=== BACKTEST STATUS ==="
$btRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/backtest/status' -Headers $hdrs
$btRes | ConvertTo-Json -Depth 5

Write-Host "=== PERFORMANCE ==="
$perfRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/analytics/performance' -Headers $hdrs
$perfRes | ConvertTo-Json -Depth 5

Write-Host "=== CORRELATION ==="
$corrRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/analytics/correlation' -Headers $hdrs
$corrRes | ConvertTo-Json -Depth 3
