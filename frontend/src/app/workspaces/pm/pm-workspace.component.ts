import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DataService, Strategy } from '../../services/data.service';
import { ChartWrapperComponent } from '../../shared/components/chart-wrapper.component';
import { ChartConfiguration } from 'chart.js';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

@Component({
  selector: 'app-pm-workspace',
  standalone: true,
  imports: [CommonModule, ChartWrapperComponent],
  templateUrl: './pm-workspace.component.html',
  styleUrl: './pm-workspace.component.scss'
})
export class PmWorkspaceComponent implements OnInit {
  // Leaderboard
  strategies: Strategy[] = [];
  
  // Heatmap
  stratCorr: number[][] = [];
  stratLabels: string[] = [];
  heatmapCells: any[] = [];
  hoveredCell: any = null;
  sparklines: { [key: string]: SafeHtml } = {};

  // Comparison
  modelsPerf: any[] = [];
  
  // Allocation
  deployment: any[] = [];
  
  // Depth Chart
  depthConfig?: ChartConfiguration;

  private originalStrategies: any[] = [];

  constructor(
    public dataService: DataService,
    public sanitizer: DomSanitizer
  ) { }

  ngOnInit() {
    this.strategies = this.dataService.getStrategies();
    this.originalStrategies = [...this.strategies];
    this.stratCorr = this.dataService.getStratCorr();
    this.stratLabels = this.strategies.map(s => s.id);
    this.modelsPerf = this.dataService.getModelsPerf();
    this.deployment = this.dataService.getDeployment();

    this.generateHeatmap();
    this.initDepthChart();
    this.generateSparklines();
  }

  filterStrats(query: string) {
    if (!query) {
      this.strategies = [...this.originalStrategies];
    } else {
      const q = query.toLowerCase();
      this.strategies = this.originalStrategies.filter(s => 
        s.id.toLowerCase().includes(q) || 
        s.author.toLowerCase().includes(q) ||
        s.model.toLowerCase().includes(q)
      );
    }
  }

  sortStrats(col: string) {
    this.strategies.sort((a, b) => {
      const v1 = (a as any)[col];
      const v2 = (b as any)[col];
      if (typeof v1 === 'string') return v1.localeCompare(v2);
      return (v2 as number) - (v1 as number);
    });
  }

  generateSparklines() {
    this.strategies.forEach(s => {
      this.sparklines[s.id] = this.getSparkline(s.rollPnL);
    });
  }

  getSparkline(data: number[]): SafeHtml {
    const W = 70, H = 22;
    const mn = Math.min(...data), mx = Math.max(...data);
    const rng = mx - mn || 1;
    const pts = data.map((v, i) => {
      const x = (i / (data.length - 1)) * W;
      const y = H - ((v - mn) / rng) * H;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    }).join(' ');
    
    const last = data[data.length - 1];
    const col = last >= 0 ? '#22d3a0' : '#f45c5c';
    const zy = H - ((0 - mn) / rng) * H;

    const svg = `<svg width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" style="display:block">
      <line x1="0" y1="${zy.toFixed(1)}" x2="${W}" y2="${zy.toFixed(1)}" stroke="rgba(107,114,128,0.3)" stroke-width="0.5" stroke-dasharray="2,2"/>
      <polyline points="${pts}" fill="none" stroke="${col}" stroke-width="1.5"/>
      <circle cx="${W}" cy="${(H-((last-mn)/rng)*H).toFixed(1)}" r="2" fill="${col}"/>
    </svg>`;
    
    return this.sanitizer.bypassSecurityTrustHtml(svg);
  }

  private generateHeatmap() {
    const n = this.stratLabels.length;
    this.heatmapCells = [];
    for (let i = 0; i < n; i++) {
      for (let j = 0; j < n; j++) {
        const v = this.stratCorr[i][j];
        this.heatmapCells.push({
          row: i, col: j, val: v,
          labelI: this.stratLabels[i],
          labelJ: this.stratLabels[j],
          style: {
            background: this.getHMColor(v),
            borderRadius: '2px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '6px',
            color: '#fff',
            fontWeight: '600'
          }
        });
      }
    }
  }

  getHMColor(v: number): string {
    const r = v > 0 ? 34 : 244, g = v > 0 ? 211 : 92, b = v > 0 ? 160 : 92;
    const a = Math.abs(v) * 0.75 + 0.06;
    return `rgba(${r},${g},${b},${a})`;
  }

  showTooltip(cell: any, event: MouseEvent) {
    this.hoveredCell = {
      val: cell.val,
      x: event.clientX + 14,
      y: event.clientY - 32,
      labelI: cell.labelI,
      labelJ: cell.labelJ,
      modelI: this.strategies[cell.row]?.model,
      modelJ: this.strategies[cell.col]?.model,
      weightsI: this.strategies[cell.row]?.weights,
      weightsJ: this.strategies[cell.col]?.weights
    };
  }

  hideTooltip() {
    this.hoveredCell = null;
  }

  private initDepthChart() {
    const ticks = this.dataService.getTicks();
    const last = ticks[ticks.length - 1];
    
    // Reverse ask price/qty for the chart
    const bidPrices = [last.bid1, last.bid2, last.bid3].map(v => v.toFixed(1));
    const askPrices = [last.ask1, last.ask2, last.ask3].map(v => v.toFixed(1));
    const askPricesRev = [...askPrices].reverse();
    
    const bidQtys = [last.bq1, last.bq2, last.bq3];
    const askQtys = [last.aq1, last.aq2, last.aq3];
    const askQtysRev = [...askQtys].reverse();

    this.depthConfig = {
      type: 'bar',
      data: {
        labels: [...askPricesRev, ...bidPrices],
        datasets: [
          {
            label: 'Ask Depth',
            data: [...askQtysRev, 0, 0, 0],
            backgroundColor: 'rgba(244,92,92,0.55)',
            borderColor: '#f45c5c',
            borderWidth: 1
          },
          {
            label: 'Bid Depth',
            data: [0, 0, 0, ...bidQtys],
            backgroundColor: 'rgba(34,211,160,0.55)',
            borderColor: '#22d3a0',
            borderWidth: 1
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { 
            display: true, 
            labels: { color: '#6b7280', font: { size: 9, family: 'Roboto Mono' }, boxWidth: 10 } 
          },
          tooltip: { 
            backgroundColor: '#1a1d28', borderColor: '#252836', borderWidth: 1 
          },
          title: {
            display: true,
            text: `Order Book Depth — ${new Date().toLocaleTimeString()}`,
            color: '#6b7280',
            font: { size: 9, family: 'Roboto Mono' }
          }
        },
        scales: {
          x: { 
            grid: { color: '#151820' }, 
            ticks: { color: '#3d4460', font: { size: 8 } },
            title: { display: true, text: 'Price', color: '#6b7280', font: { size: 8 } }
          },
          y: { 
            grid: { color: '#151820' }, 
            ticks: { color: '#3d4460', font: { size: 8 } },
            title: { display: true, text: 'Quantity', color: '#6b7280', font: { size: 8 } }
          }
        }
      }
    };
  }
}
