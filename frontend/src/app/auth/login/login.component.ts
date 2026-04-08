import { Component, signal, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="login-bg">
      <div class="login-card">
        <!-- Logo -->
        <div class="logo-area">
          <div class="logo-icon">⚡</div>
          <h1 class="logo-title">QuantAlpha Lab</h1>
          <p class="logo-sub">HFT Research Platform</p>
        </div>

        <!-- Form -->
        <form (ngSubmit)="onSubmit()" class="login-form">
          <div class="field-group">
            <label class="field-label" for="username">Username</label>
            <input
              id="username"
              type="text"
              class="field-input"
              [(ngModel)]="username"
              name="username"
              placeholder="ds1, qr1, pm1, admin"
              autocomplete="username"
              required
            />
          </div>

          <div class="field-group">
            <label class="field-label" for="password">Password</label>
            <input
              id="password"
              type="password"
              class="field-input"
              [(ngModel)]="password"
              name="password"
              placeholder="••••••••"
              autocomplete="current-password"
              required
            />
          </div>

          @if (error()) {
            <div class="error-banner">
              <span class="error-icon">⚠</span> {{ error() }}
            </div>
          }

          <button
            type="submit"
            class="submit-btn"
            [disabled]="loading() || !username || !password"
          >
            @if (loading()) {
              <span class="spinner"></span> Signing in…
            } @else {
              Sign In →
            }
          </button>
        </form>

        <!-- Role hint -->
        <div class="role-hints">
          <span class="hint-label">Roles:</span>
          <code class="role-badge admin">admin</code>
          <code class="role-badge ds">ds</code>
          <code class="role-badge qr">qr</code>
          <code class="role-badge pm">pm</code>
        </div>

        <p class="version-tag">v2.4.1 · Mock Auth</p>
      </div>
    </div>
  `,
  styles: [`
    :host { display: block; position: fixed; inset: 0; z-index: 9999; }
    .login-bg {
      min-height: 100vh;
      background: #0a0e1a;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: 'Inter', 'Segoe UI', sans-serif;
      background-image:
        radial-gradient(ellipse at 20% 50%, rgba(99, 102, 241, 0.08) 0%, transparent 60%),
        radial-gradient(ellipse at 80% 20%, rgba(16, 185, 129, 0.06) 0%, transparent 50%);
    }
    .login-card {
      background: rgba(15, 20, 35, 0.95);
      border: 1px solid rgba(255,255,255,0.08);
      border-radius: 16px;
      padding: 48px 40px 36px;
      width: 100%;
      max-width: 420px;
      box-shadow: 0 24px 80px rgba(0,0,0,0.5), 0 0 0 1px rgba(99,102,241,0.1);
      backdrop-filter: blur(12px);
    }
    .logo-area { text-align: center; margin-bottom: 36px; }
    .logo-icon { font-size: 40px; margin-bottom: 12px; filter: drop-shadow(0 0 12px rgba(99,102,241,0.6)); }
    .logo-title { font-size: 24px; font-weight: 700; color: #f1f5f9; margin: 0 0 6px; letter-spacing: -0.5px; }
    .logo-sub { font-size: 13px; color: #64748b; margin: 0; letter-spacing: 0.5px; text-transform: uppercase; }
    .login-form { display: flex; flex-direction: column; gap: 20px; }
    .field-group { display: flex; flex-direction: column; gap: 8px; }
    .field-label { font-size: 11px; font-weight: 600; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.8px; }
    .field-input {
      background: rgba(255,255,255,0.04);
      border: 1px solid rgba(255,255,255,0.1);
      border-radius: 8px;
      color: #f1f5f9;
      font-size: 14px;
      padding: 12px 16px;
      outline: none;
      transition: border-color 0.2s, box-shadow 0.2s;
    }
    .field-input::placeholder { color: #475569; }
    .field-input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.15); }
    .error-banner {
      background: rgba(239,68,68,0.1);
      border: 1px solid rgba(239,68,68,0.3);
      border-radius: 8px;
      color: #fca5a5;
      font-size: 13px;
      padding: 10px 14px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .submit-btn {
      background: linear-gradient(135deg, #6366f1, #4f46e5);
      border: none;
      border-radius: 8px;
      color: #fff;
      cursor: pointer;
      font-size: 15px;
      font-weight: 600;
      padding: 13px;
      margin-top: 4px;
      transition: opacity 0.2s, transform 0.1s;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
    }
    .submit-btn:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
    .submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }
    .spinner {
      width: 16px;
      height: 16px;
      border: 2px solid rgba(255,255,255,0.3);
      border-top-color: #fff;
      border-radius: 50%;
      animation: spin 0.7s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
    .role-hints { display: flex; align-items: center; gap: 8px; margin-top: 28px; flex-wrap: wrap; }
    .hint-label { font-size: 11px; color: #475569; text-transform: uppercase; }
    .role-badge { font-size: 11px; font-weight: 600; padding: 2px 8px; border-radius: 4px; }
    .role-badge.admin { background: rgba(234,179,8,0.15); color: #fbbf24; }
    .role-badge.ds { background: rgba(99,102,241,0.15); color: #818cf8; }
    .role-badge.qr { background: rgba(16,185,129,0.15); color: #34d399; }
    .role-badge.pm { background: rgba(249,115,22,0.15); color: #fb923c; }
    .version-tag { text-align: center; font-size: 11px; color: #334155; margin: 20px 0 0; }
  `],
})
export class LoginComponent {
  username = '';
  password = '';
  loading = signal(false);
  error = signal<string | null>(null);
  
  loginSuccess = output<{ username: string; role: string }>();

  onSubmit(): void {
    if (!this.username || !this.password) return;
    
    this.loading.set(true);
    this.error.set(null);

    // Mock delay
    setTimeout(() => {
      this.loading.set(false);
      
      // Basic mock check
      const validAdmin = this.username === 'admin' && this.password === 'admin';
      const validDS = this.username.startsWith('ds') && this.password === 'ds';
      const validQR = this.username.startsWith('qr') && this.password === 'qr';
      const validPM = this.username.startsWith('pm') && this.password === 'pm';

      if (validAdmin || validDS || validQR || validPM) {
        let role = 'ADMIN';
        if (validDS) role = 'DS';
        if (validQR) role = 'QR';
        if (validPM) role = 'PM';

        this.loginSuccess.emit({ username: this.username, role });
      } else {
        this.error.set('Invalid credentials. Hint: use admin/admin or ds/ds');
      }
    }, 1200);
  }
}
