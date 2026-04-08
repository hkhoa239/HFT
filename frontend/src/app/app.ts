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

  onLoginSuccess(data: { username: string; role: string }): void {
    this.currentUser.set(data);
    this.isLoggedIn.set(true);
    
    // Set active workspace based on role if needed
    if (data.role === 'DS') this.activeWorkspace = 'DS';
    if (data.role === 'PM') this.activeWorkspace = 'PM';
    if (data.role === 'QR') this.activeWorkspace = 'QR';
  }

  logout(): void {
    this.isLoggedIn.set(false);
    this.currentUser.set(null);
  }
}
