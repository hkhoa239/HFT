import { Component, Output, EventEmitter, OnInit, OnDestroy, Input } from '@angular/core';

@Component({
  selector: 'app-topbar',
  standalone: true,
  host: { 'class': 'topbar' },
  templateUrl: './topbar.component.html',
  styleUrl: './topbar.component.scss'
})
export class TopbarComponent implements OnInit, OnDestroy {
  @Input() workspace = 'QR';
  @Input() user: { username: string; role: string } | null = null;
  @Output() workspaceChange = new EventEmitter<string>();
  @Output() logout = new EventEmitter<void>();
  
  timeString = '';
  private timer: any;

  ngOnInit() {
    this.updateTime();
    this.timer = setInterval(() => this.updateTime(), 1000);
  }

  private updateTime() {
    const now = new Date();
    const time = now.toLocaleTimeString('en-GB', { hour12: false });
    const date = now.toISOString().split('T')[0];
    this.timeString = `${time} · ${date}`;
  }

  ngOnDestroy() {
    if (this.timer) clearInterval(this.timer);
  }

  setWorkspace(ws: string) {
    this.workspace = ws;
    this.workspaceChange.emit(ws);
  }

  onLogout() {
    this.logout.emit();
  }

  getInitials(): string {
    if (!this.user?.username) return '??';
    const parts = this.user.username.split(/[._-]/);
    if (parts.length > 1) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return this.user.username.slice(0, 2).toUpperCase();
  }
}
