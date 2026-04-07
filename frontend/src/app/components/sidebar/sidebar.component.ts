import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule],
  host: { 'class': 'sidebar' },
  templateUrl: './sidebar.component.html',
  styleUrl: './sidebar.component.scss'
})
export class SidebarComponent {
  @Input() workspace = 'QR';
  @Output() workspaceChange = new EventEmitter<string>();

  setWorkspace(ws: string) {
    this.workspace = ws;
    this.workspaceChange.emit(this.workspace);
  }
}
