import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TopbarComponent } from './components/topbar/topbar.component';
import { SidebarComponent } from './components/sidebar/sidebar.component';
import { StatusbarComponent } from './components/statusbar/statusbar.component';
import { QrWorkspaceComponent } from './workspaces/qr/qr-workspace.component';
import { DsWorkspaceComponent } from './workspaces/ds/ds-workspace.component';
import { PmWorkspaceComponent } from './workspaces/pm/pm-workspace.component';
import { ToastComponent } from './components/toast/toast.component';

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
    ToastComponent
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  activeWorkspace = 'QR';
}
