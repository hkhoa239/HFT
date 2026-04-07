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
  @Output() workspaceChange = new EventEmitter<string>();
  
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
}
