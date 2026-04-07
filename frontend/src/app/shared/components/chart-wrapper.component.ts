import { Component, ElementRef, Input, OnChanges, OnInit, SimpleChanges, ViewChild } from '@angular/core';
import { Chart, ChartConfiguration, registerables } from 'chart.js';

Chart.register(...registerables);

@Component({
  selector: 'app-chart',
  standalone: true,
  templateUrl: './chart-wrapper.component.html',
  styleUrl: './chart-wrapper.component.scss'
})
export class ChartWrapperComponent implements OnInit, OnChanges {
  @ViewChild('chartCanvas', { static: true }) canvas!: ElementRef<HTMLCanvasElement>;
  @Input() config!: ChartConfiguration;
  
  private chart?: Chart;

  ngOnInit() {
    this.createChart();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes['config'] && !changes['config'].firstChange) {
      this.updateChart();
    }
  }

  private createChart() {
    if (this.chart) {
      this.chart.destroy();
    }
    this.chart = new Chart(this.canvas.nativeElement, this.config);
  }

  private updateChart() {
    if (this.chart) {
      this.chart.data = this.config.data;
      this.chart.options = this.config.options || {};
      this.chart.update('none');
    }
  }
}
