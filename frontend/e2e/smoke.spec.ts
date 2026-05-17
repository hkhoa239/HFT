import { test, expect } from '@playwright/test';

test.describe('QuantAlpha Smoke Test', () => {
  test.beforeEach(async ({ page }) => {
    // 1. Login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'quant');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/qr');
  });

  test('Alpha Creation and Backtest Flow', async ({ page }) => {
    // 2. Alpha Creation
    await page.click('text=New Alpha');
    await page.fill('input[placeholder="Alpha Name"]', 'E2E Test Alpha');
    await page.fill('textarea[placeholder="Description"]', 'Created by Playwright');
    // CodeMirror is tricky to fill, assuming a standard textarea or finding the hidden one
    // In a real app we might need to click and type
    await page.click('.cm-content');
    await page.keyboard.type('def signal(row): return 1');
    await page.click('button:has-text("Save")');
    await expect(page.locator('text=Alpha saved successfully')).toBeVisible();

    // 3. Backtest Execution
    await page.click('button:has-text("Run Backtest")');
    await page.click('button:has-text("Confirm Run")');
    await expect(page.locator('text=Backtest queued')).toBeVisible();
    
    // Polling check (wait for status to change from pending to completed or failed)
    await expect(page.locator('text=Completed')).toBeVisible({ timeout: 30000 });
  });

  test('Dashboard Rendering', async ({ page }) => {
    // 4. Portfolio Dashboard
    await page.goto('/pm');
    await expect(page.locator('h1')).toContainText('Portfolio Manager');
    await expect(page.locator('canvas')).toBeVisible(); // Correlation matrix or PnL chart
  });

  test('Login Failure Scenario', async ({ page }) => {
    // 5. Failure Scenario
    await page.goto('/login');
    await page.fill('input[name="username"]', 'wronguser');
    await page.fill('input[name="password"]', 'wrongpass');
    await page.click('button[type="submit"]');
    await expect(page.locator('.error-message')).toBeVisible();
  });
});
