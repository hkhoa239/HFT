$loginRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/auth/login' -Method POST -ContentType 'application/json' -Body '{"username":"admin","password":"password123"}'
$tok = $loginRes.data.token
$hdrs = @{Authorization="Bearer $tok"}

Write-Host "=== ALPHAS (submitted) ==="
$alpRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/alphas/submitted' -Headers $hdrs
$alpRes | ConvertTo-Json -Depth 3

Write-Host "=== MY ALPHAS ==="
$myAlpRes = Invoke-RestMethod -Uri 'http://127.0.0.1:8080/alphas/me' -Headers $hdrs
$myAlpRes | ConvertTo-Json -Depth 3
