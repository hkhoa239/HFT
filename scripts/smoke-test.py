import requests
import time
import sys
import os

# Configuration
BASE_URL = os.environ.get("API_URL", "http://127.0.0.1:8080")
MAX_RETRIES = 30
RETRY_INTERVAL = 2

# Seed credentials from init.sql
ADMIN_USER = "admin"
ADMIN_PASS = "password123"

def wait_for_ready():
    print(f"Waiting for backend readiness at {BASE_URL}/ready ...")
    for i in range(MAX_RETRIES):
        try:
            resp = requests.get(f"{BASE_URL}/ready", timeout=5)
            if resp.status_code == 200:
                data = resp.json()
                if data.get("success"):
                    print("Backend is READY.")
                    return True
            print(f"[{i+1}/{MAX_RETRIES}] Backend not ready: {resp.status_code} {resp.text}")
        except Exception as e:
            print(f"[{i+1}/{MAX_RETRIES}] Backend unreachable: {e}")
        
        time.sleep(RETRY_INTERVAL)
    return False

def test_login():
    print(f"Attempting login for user: {ADMIN_USER} ...")
    login_url = f"{BASE_URL}/auth/login"
    try:
        resp = requests.post(login_url, json={
            "username": ADMIN_USER,
            "password": ADMIN_PASS
        }, timeout=5)
        
        if resp.status_code != 200:
            print(f"Login FAILED: {resp.status_code} {resp.text}")
            return None
        
        data = resp.json()
        if not data.get("success"):
            print(f"Login Success flag FALSE: {data}")
            return None
        
        token = data["data"]["token"]
        print("Login SUCCESS. JWT extracted.")
        return token
    except Exception as e:
        print(f"Login request ERROR: {e}")
        return None

def test_authenticated_endpoint(token):
    print("Verifying /alphas/me endpoint ...")
    headers = {"Authorization": f"Bearer {token}"}
    try:
        resp = requests.get(f"{BASE_URL}/alphas/me", headers=headers, timeout=5)
        if resp.status_code != 200:
            print(f"API Call FAILED: {resp.status_code} {resp.text}")
            return False
        
        data = resp.json()
        if not data.get("success"):
            print(f"API Success flag FALSE: {data}")
            return False
        
        if not isinstance(data.get("data"), list):
            print(f"Unexpected JSON structure: 'data' should be a list, got {type(data.get('data'))}")
            return False
        
        print(f"API Call SUCCESS. Found {len(data['data'])} alphas.")
        return True
    except Exception as e:
        print(f"API request ERROR: {e}")
        return False

def main():
    print("=== QuantAlpha Smoke Test Baseline ===")
    
    if not wait_for_ready():
        print("CRITICAL: Backend failed to become ready in time.")
        sys.exit(1)
    
    token = test_login()
    if not token:
        print("CRITICAL: Authentication failed.")
        sys.exit(1)
        
    if not test_authenticated_endpoint(token):
        print("CRITICAL: Authenticated endpoint validation failed.")
        sys.exit(1)
        
    print("=== SMOKE TEST PASSED ===")
    sys.exit(0)

if __name__ == "__main__":
    main()
