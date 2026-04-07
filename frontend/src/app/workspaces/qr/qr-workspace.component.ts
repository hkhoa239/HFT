import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DataService, Tick, Strategy, PnLData } from '../../services/data.service';
import { NotificationService } from '../../services/notification.service';
import { ChartWrapperComponent } from '../../shared/components/chart-wrapper.component';
import { EditorWrapperComponent } from '../../shared/components/editor-wrapper.component';
import { ChartConfiguration } from 'chart.js';

@Component({
  selector: 'app-qr-workspace',
  standalone: true,
  imports: [CommonModule, ChartWrapperComponent, EditorWrapperComponent],
  templateUrl: './qr-workspace.component.html',
  styleUrl: './qr-workspace.component.scss'
})
export class QrWorkspaceComponent implements OnInit, OnDestroy {
  qrCode = '';
  pnlConfig?: ChartConfiguration;
  
  // Variable Library
  variables: any = {};
  varCategories: string[] = [];
  varSearch = '';

  // Simulation
  isSimulating = false;
  simProgress = 0;
  simEta = '';
  private simTimer: any;

  // Metrics
  metrics: any[] = [];
  pnlSummary: any = {
    total: '+42.3 bps',
    wr: 'Win Rate 56.4%',
    lastTrace: '↑ BUY +5.0 bps',
    obi: '+0.312',
    rr: '+0.0021%'
  };

  constructor(
    public dataService: DataService,
    public ns: NotificationService
  ) { }

  ngOnInit() {
    this.variables = this.dataService.getVariables();
    this.varCategories = Object.keys(this.variables);
    
    import('../../services/data.service').then(m => {
      this.qrCode = m.QR_CODE;
    });

    this.updateMetrics();
    this.updatePnlChart();
  }

  ngOnDestroy() {
    if (this.simTimer) clearInterval(this.simTimer);
  }

  filterVars(q: string) {
    this.varSearch = q.toLowerCase();
  }

  getFilteredVars(cat: string) {
    if (!this.varSearch) return this.variables[cat];
    return this.variables[cat].filter((v: any) => 
      v.name.toLowerCase().includes(this.varSearch) || 
      v.desc.toLowerCase().includes(this.varSearch)
    );
  }

  runBacktest() {
    if (this.isSimulating) return;
    
    this.isSimulating = true;
    this.simProgress = 0;
    this.simEta = 'Calculating…';
    
    this.simTimer = setInterval(() => {
      this.simProgress += Math.random() * 7 + 3;
      if (this.simProgress >= 100) {
        this.simProgress = 100;
        clearInterval(this.simTimer);
        this.simEta = 'Done!';
        
        setTimeout(() => {
          this.isSimulating = false;
          this.simProgress = 0;
          this.ns.success('Backtest complete — Total PnL: +42.3 bps, Win Rate: 56.4%');
          this.updatePnlChart();
        }, 400);
      } else {
        this.simEta = 'ETA: ' + Math.ceil((100 - this.simProgress) / 8) + 's';
      }
    }, 200);
  }

  private updateMetrics() {
    this.metrics = this.dataService.getQrMetrics();
  }

  private updatePnlChart() {
    const pnl = this.dataService.getPnlInit();
    const { perTrade, equity } = pnl;
    const n = perTrade.length;
    const labels = Array.from({length: n}, (_,i) => `T${i+1}`);
    const cumPnL = perTrade.reduce((acc: number[], v) => {
      acc.push((acc.length ? acc[acc.length-1] : 0) + v);
      return acc;
    }, []);

    const barColors = perTrade.map(v => 
      v > 0 ? 'rgba(34,211,160,0.60)' : 
      v < 0 ? 'rgba(244,92,92,0.60)'  : 
              'rgba(107,114,128,0.15)');

    this.pnlConfig = {
      type: 'bar',
      data: {
        labels,
        datasets: [
          {
            type: 'bar',
            label: 'P&L per window',
            data: perTrade as any,
            backgroundColor: barColors,
            yAxisID: 'y',
            order: 2
          },
          {
            type: 'line',
            label: 'Cum P&L (bps)',
            data: cumPnL as any,
            borderColor: '#4f8ef7',
            borderWidth: 2,
            pointRadius: 0,
            fill: true,
            backgroundColor: 'rgba(79,142,247,0.1)',
            tension: 0.3,
            yAxisID: 'y',
            order: 1
          },
          {
            type: 'line',
            label: 'Equity Curve',
            data: equity.slice(0, n) as any,
            borderColor: '#22d3a0',
            borderWidth: 1.5,
            pointRadius: 0,
            fill: false,
            borderDash: [4, 3],
            tension: 0.3,
            yAxisID: 'y2',
            order: 0
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: true, labels: { color:'#6b7280', font:{size:9,family:'Roboto Mono'}, boxWidth:12 } },
          tooltip: { mode: 'index', intersect: false, backgroundColor:'#1a1d28', borderColor:'#252836', borderWidth:1 }
        },
        scales: {
          x: { grid:{color:'#151820'}, ticks:{color:'#3d4460', maxTicksLimit:12, font:{size:8}} },
          y: {
            position: 'left',
            grid:{color:'#151820'},
            ticks:{color:'#3d4460', font:{size:8}},
            title:{display:true, text:'P&L (bps)', color:'#6b7280', font:{size:8}}
          },
          y2: {
            position: 'right',
            grid:{drawOnChartArea:false},
            ticks:{color:'#3d4460', font:{size:8}},
            title:{display:true, text:'Equity', color:'#6b7280', font:{size:8}}
          }
        }
      }
    };
  }
}

