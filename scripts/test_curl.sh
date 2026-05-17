#!/bin/bash
# QuantAlpha HFT Platform - Comprehensive API Test using curl (Bash/WSL/POSIX)

BASE_URL="http://127.0.0.1:8080"
USERNAME="admin"
PASSWORD="password123"

echo "==================================================="
echo "QuantAlpha HFT Platform - Comprehensive Admin API Test"
echo "==================================================="
echo "Logging in as '$USERNAME'..."

# Perform login and capture response
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# Parse token using python3
TOKEN=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  TOKEN=$(echo "$RESPONSE" | jq -r '.data.token' 2>/dev/null)
fi

if [ -z "$TOKEN" ]; then
  TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
fi

if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to extract token from login response!"
  echo "Response was: $RESPONSE"
  exit 1
fi

echo "Login SUCCESS. Extracted Token (first 30 chars): ${TOKEN:0:30}..."
echo ""

run_curl_get() {
  local endpoint="$1"
  local description="$2"
  echo "---------------------------------------------------"
  echo "Querying: $description ($endpoint)"
  echo "---------------------------------------------------"
  local res=$(curl -s -X GET "$BASE_URL$endpoint" -H "Authorization: Bearer $TOKEN")
  echo "$res" | python3 -m json.tool 2>/dev/null || echo "$res" | jq . 2>/dev/null || echo "$res"
  echo ""
}

# 1. User Profile
run_curl_get "/me/profile" "Current User Profile"

# 2. General Registries
run_curl_get "/models" "Trained Models List"
run_curl_get "/factors" "Published Factors Registry"

# 3. Quant Researcher (QR) Workspace Endpoints
run_curl_get "/alphas/me" "My Created Alphas"
run_curl_get "/alphas/me/submitted" "My Submitted Alphas"
run_curl_get "/backtest/me/status" "My Backtest Status Summary"

# 4. Data Scientist (DS) Analytics & Overview
run_curl_get "/analytics/ds/overview" "DS Overview Statistics"
run_curl_get "/analytics/ds/models" "DS Model Training Performance Metrics"

# 5. Portfolio Manager (PM) Portfolios & Risks
run_curl_get "/alphas/submitted" "PM Submitted Alphas List"
run_curl_get "/backtest/status" "PM All Backtests Status"
run_curl_get "/analytics/correlation" "PM Strategy Correlation Matrix"
run_curl_get "/analytics/performance" "PM Strategy Performance Metrics"
run_curl_get "/audit-logs" "PM System Audit Logs"

# 6. Admin Panel User Management
run_curl_get "/admin/users" "Admin User Directory"

echo "==================================================="
echo "COMPLETED COMPREHENSIVE ADMIN API TEST"
echo "==================================================="
