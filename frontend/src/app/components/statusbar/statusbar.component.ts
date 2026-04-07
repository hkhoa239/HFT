import { Component } from '@angular/core';

@Component({
  selector: 'app-statusbar',
  standalone: true,
  host: { 'class': 'statusbar' },
  templateUrl: './statusbar.component.html',
  styleUrl: './statusbar.component.scss'
})
export class StatusbarComponent {
  public lastBacktest = '—';
}
