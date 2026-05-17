@echo off
set BASE_URL=http://127.0.0.1:8080
set USERNAME=admin
set PASSWORD=password123

echo ===================================================
echo QuantAlpha HFT Platform - Comprehensive API Test (Windows)
echo ===================================================
echo.
echo Logging in as %USERNAME% to get JWT Access Token...
curl -s -X POST %BASE_URL%/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"%USERNAME%\",\"password\":\"%PASSWORD%\"}" > login_response.json

echo Login response saved.
echo.
echo Trying to extract token using python...
for /f "tokens=*" %%i in ('python -c "import json; print(json.load(open('login_response.json'))['data']['token'])" 2^>nul') do set TOKEN=%%i

if "%TOKEN%"=="" (
  echo Error: Python not found or token extraction failed!
  echo Please inspect login_response.json for details.
  del login_response.json 2>nul
  exit /b 1
)

echo Login SUCCESS. Token loaded.
echo.

call :run_curl "/me/profile" "Current User Profile"
call :run_curl "/models" "Trained Models List"
call :run_curl "/factors" "Published Factors Registry"
call :run_curl "/alphas/me" "My Created Alphas"
call :run_curl "/alphas/me/submitted" "My Submitted Alphas"
call :run_curl "/backtest/me/status" "My Backtest Status Summary"
call :run_curl "/analytics/ds/overview" "DS Overview Statistics"
call :run_curl "/analytics/ds/models" "DS Model Training Performance Metrics"
call :run_curl "/alphas/submitted" "PM Submitted Alphas List"
call :run_curl "/backtest/status" "PM All Backtests Status"
call :run_curl "/analytics/correlation" "PM Strategy Correlation Matrix"
call :run_curl "/analytics/performance" "PM Strategy Performance Metrics"
call :run_curl "/audit-logs" "PM System Audit Logs"
call :run_curl "/admin/users" "Admin User Directory"

echo ===================================================
echo COMPLETED COMPREHENSIVE ADMIN API TEST
echo ===================================================
del login_response.json 2>nul
exit /b 0

:run_curl
echo ---------------------------------------------------
echo Querying: %~2 (%~1)
echo ---------------------------------------------------
curl -s -X GET %BASE_URL%%~1 -H "Authorization: Bearer %TOKEN%"
echo.
echo.
goto :eof
