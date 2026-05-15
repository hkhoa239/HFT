import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TopbarComponent } from './components/topbar/topbar.component';
import { SidebarComponent } from './components/sidebar/sidebar.component';
import { StatusbarComponent } from './components/statusbar/statusbar.component';
import { QrWorkspaceComponent } from './workspaces/qr/qr-workspace.component';
import { DsWorkspaceComponent } from './workspaces/ds/ds-workspace.component';
import { PmWorkspaceComponent } from './workspaces/pm/pm-workspace.component';
import { ToastComponent } from './components/toast/toast.component';
import { LoginComponent } from './auth/login/login.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    TopbarComponent,
    SidebarComponent,
    StatusbarComponent,
    QrWorkspaceComponent,
    DsWorkspaceComponent,
    PmWorkspaceComponent,
    ToastComponent,
    LoginComponent
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  activeWorkspace = 'QR';
  isLoggedIn = signal(false);
  currentUser = signal<{ username: string; role: string } | null>(null);

  constructor() {
    this.hydrateAuth();
  }

  private hydrateAuth(): void {
    try {
      const token = localStorage.getItem('auth_token');
      const userJson = localStorage.getItem('auth_user');
      if (token && userJson) {
        const user = JSON.parse(userJson);
        this.currentUser.set(user);
        this.isLoggedIn.set(true);
        this.activeWorkspace = user.role;
      }
    } catch (e) {
      console.error('Failed to hydrate auth state', e);
      this.logout();
    }
  }

  onLoginSuccess(data: { username: string; role: string }): void {
    this.currentUser.set(data);
    this.isLoggedIn.set(true);
    this.activeWorkspace = data.role;
  }

  logout(): void {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
    this.isLoggedIn.set(false);
    this.currentUser.set(null);
  }
}
