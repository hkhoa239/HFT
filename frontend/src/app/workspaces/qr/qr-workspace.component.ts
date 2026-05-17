import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { interval, Subscription } from 'rxjs';
import { switchMap, takeWhile, take } from 'rxjs/operators';
import { DataService } from '../../services/data.service';
import { NotificationService } from '../../services/notification.service';
import { ChartWrapperComponent } from '../../shared/components/chart-wrapper.component';
import { EditorWrapperComponent } from '../../shared/components/editor-wrapper.component';
import { ChartConfiguration } from 'chart.js';
import { APP_CONFIG } from '../../app.constants';

@Component({
  selector: 'app-qr-workspace',
  standalone: true,
  imports: [CommonModule, ChartWrapperComponent, EditorWrapperComponent, FormsModule],
  templateUrl: './qr-workspace.component.html',
  styleUrl: './qr-workspace.component.scss'
})
export class QrWorkspaceComponent implements OnInit, OnDestroy {
  qrCode = '';
  pnlConfig?: ChartConfiguration;
  private pollSub?: Subscription;
  private apiUrl = APP_CONFIG.apiUrl;
  
  // Variable Library
  variables: any = {};
  varCategories: string[] = [];
  varSearch = '';

  // Simulation
  isSimulating = false;
  simProgress = 0;
  simStatus = '';
  private currentJobId: string | null = null;
  alphaId: string | null = null;
  alphaStatus = 'draft';
  runParams = {
    instrument: 'VN30F2112',
    lookbackSec: 60,
    predWindowSec: 10,
    capital: 1000000,
    start: '2021-01-01',
    end: '2021-12-31'
  };

  // Metrics
  metrics: any[] = [];
  pnlSummary: any = {
    total: '—',
    wr: '—',
    lastTrace: 'Waiting for run...',
    obi: '—',
    rr: '—'
  };

  constructor(
    public dataService: DataService,
    public ns: NotificationService,
    private http: HttpClient
  ) { }

  ngOnInit() {
    this.variables = { ...this.dataService.getVariables() };
    this.varCategories = Object.keys(this.variables);
    
    import('../../services/code-templates').then(m => {
      this.qrCode = m.QR_CODE;
    });

    // Initial empty chart
    this.updatePnlChart([]);
    this.loadMyAlpha();
    this.loadLatestBacktest();
    this.loadActiveFactors();
  }

  private loadActiveFactors() {
    this.dataService.getFactors().subscribe({
      next: (factors) => {
        if (factors && factors.length > 0) {
          this.variables['ACTIVE FACTORS'] = factors.map((f: any) => ({
            name: f.name,
            desc: f.description || 'Active published factor',
            ds: true
          }));
          this.varCategories = Object.keys(this.variables);
        }
      }
    });
  }

  private loadMyAlpha() {
    this.http.get<any>(`${this.apiUrl}/alphas/me`).pipe(take(1)).subscribe({
      next: (res) => {
        if (res.success && res.data && res.data.length > 0) {
          const a = res.data[0];
          this.alphaId = a.id;
          this.alphaStatus = a.status || 'draft';
          if (!this.qrCode) this.qrCode = a.code_content || '';
        } else {
          this.alphaId = null;
          this.alphaStatus = 'draft';
        }
      }
    });
  }

  private loadLatestBacktest() {
    this.http.get<any>(`${this.apiUrl}/backtest/me/status`).pipe(take(1)).subscribe({
      next: (res) => {
        if (res.success && res.data && res.data.length > 0) {
          // Get the most recent completed backtest
          const latest = res.data.find((b: any) => b.status === 'completed');
          if (latest && latest.metrics) {
            this.handleSuccess(latest.metrics);
          }
        }
      }
    });
  }

  ngOnDestroy() {
    this.stopPolling();
  }

  private stopPolling() {
    if (this.pollSub) {
      this.pollSub.unsubscribe();
      this.pollSub = undefined;
    }
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

  saveAlpha() {
    if (!this.qrCode) return;
    
    // For MVP, we check if an alpha already exists, otherwise create one.
    this.http.get<any>(`${this.apiUrl}/alphas/me`).pipe(take(1)).subscribe({
      next: (res) => {
        if (res.success && res.data && res.data.length > 0) {
          // Update existing
          const alphaId = res.data[0].id;
          this.http.put<any>(`${this.apiUrl}/alphas/${alphaId}`, {
            name: "My Alpha Strategy",
            code_content: this.qrCode
          }).subscribe({
            next: () => { this.ns.success('Alpha updated successfully'); this.loadMyAlpha(); },
            error: () => this.ns.error('Failed to update alpha')
          });
        } else {
          // Create new
          this.http.post<any>(`${this.apiUrl}/alphas`, {
            name: "My Alpha Strategy",
            code_content: this.qrCode
          }).subscribe({
            next: () => { this.ns.success('Alpha created and saved'); this.loadMyAlpha(); },
            error: () => this.ns.error('Failed to create alpha')
          });
        }
      },
      error: () => this.ns.error('Failed to access registry')
    });
  }

  runBacktest() {
    if (this.isSimulating) return;
    
    this.simStatus = 'Checking registry...';
    this.isSimulating = true;

    this.http.get<any>(`${this.apiUrl}/alphas/me`).pipe(take(1)).subscribe({
      next: (res) => {
        if (res.success && res.data && res.data.length > 0) {
          const alphaId = res.data[0].id;
          this.startBacktest(alphaId);
        } else {
          this.resetSim();
          this.ns.error('No Alpha found. Click "Save" to register your code first.');
        }
      },
      error: () => {
        this.resetSim();
        this.ns.error('Failed to fetch Alphas');
      }
    });
  }

  private startBacktest(alphaId: string) {
    this.isSimulating = true;
    this.simStatus = 'Initializing...';
    this.simProgress = 10;

    const payload = {
      alpha_id: alphaId,
      params: {
        start: this.runParams.start,
        end: this.runParams.end,
        capital: this.runParams.capital,
        instrument: this.runParams.instrument,
        lookback_sec: this.runParams.lookbackSec,
        prediction_sec: this.runParams.predWindowSec
      }
    };

    this.http.post<any>(`${this.apiUrl}/backtest/run`, payload).subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.currentJobId = res.data.id;
          this.startPolling(this.currentJobId!);
        } else {
          this.resetSim();
          this.ns.error(res.error || 'Failed to start backtest');
        }
      },
      error: (err) => {
        this.resetSim();
        this.ns.error('Server error: ' + (err.error?.error || 'Connection failed'));
      }
    });
  }

  private startPolling(jobId: string) {
    this.simStatus = 'Job Queued...';
    this.simProgress = 30;
    const startTime = Date.now();
    const timeoutMs = 5 * 60 * 1000; // 5 minute timeout

    this.pollSub = interval(2000).pipe(
      switchMap(() => this.http.get<any>(`${this.apiUrl}/backtest/${jobId}`)),
      takeWhile(res => {
        const isRunning = res.data.status === 'pending' || res.data.status === 'running';
        const isTimedOut = Date.now() - startTime > timeoutMs;
        if (isTimedOut && isRunning) {
          this.ns.error('Backtest timed out after 5 minutes');
          this.resetSim();
          return false;
        }
        return isRunning;
      }, true)
    ).subscribe({
      next: (res) => {
        const run = res.data;
        if (run.status === 'running') {
          this.simStatus = 'Running';
          this.simProgress = 50;
        } else if (run.status === 'completed') {
          this.handleSuccess(run.metrics);
        } else if (run.status === 'failed') {
          this.handleFailure(run.error_log);
        }
      },
      error: () => {
        this.ns.error('Polling failed');
        this.resetSim();
      }
    });
  }

  submitAlpha() {
    if (!this.alphaId) {
      this.ns.error('No alpha available to submit. Save first.');
      return;
    }
    this.http.post<any>(`${this.apiUrl}/alphas/${this.alphaId}/submit`, {}).pipe(take(1)).subscribe({
      next: (res) => {
        if (res.success) {
          this.alphaStatus = 'submitted';
          this.ns.success('Alpha submitted');
          this.loadMyAlpha();
        } else {
          this.ns.error(res.error || 'Submit failed');
        }
      },
      error: () => this.ns.error('Submit failed')
    });
  }

  private handleSuccess(metrics: any) {
    this.stopPolling();
    this.isSimulating = false;
    this.simProgress = 100;
    this.simStatus = 'Completed';
    
    if (!metrics) {
      this.ns.error('Received empty metrics from backtest');
      return;
    }

    // Update UI with real metrics - with safe defaults
    const totalPnl = metrics.total_pnl ?? 0;
    const winRate = (metrics.win_rate ?? 0) * 100;
    const sharpe = metrics.sharpe_ratio ?? 0;
    const trades = metrics.trade_count ?? 0;

    this.pnlSummary = {
      total: (totalPnl >= 0 ? '+' : '') + totalPnl.toFixed(1) + ' bps',
      wr: `Win Rate ${winRate.toFixed(1)}%`,
      lastTrace: `Trades: ${trades}`,
      obi: sharpe.toFixed(2),
      rr: 'Backtest OK'
    };

    this.metrics = [
      { l: 'Total PnL', v: this.pnlSummary.total, s: 'Real result from engine', cls: 'cg' },
      { l: 'Sharpe', v: sharpe.toFixed(2), s: 'Risk-adjusted return', cls: 'ca' },
      { l: 'Win Rate', v: winRate.toFixed(1) + '%', s: 'Correct signals', cls: 'cg' },
      { l: 'Trade Count', v: trades, s: 'Execution frequency', cls: 'cb' }
    ];

    const rawCurve = metrics.pnl_curve || [];
    const pnlCurvePoints = rawCurve.map((p: any) => 
      typeof p === 'number' ? p : (p && typeof p.cumPnL === 'number' ? p.cumPnL : 0)
    );
    this.updatePnlChart(pnlCurvePoints);
    this.ns.success('Backtest complete!');
  }

  private handleFailure(errorLog: string) {
    this.stopPolling();
    this.resetSim();
    this.simStatus = 'Failed';
    this.ns.error('Backtest failed: ' + (errorLog || 'Unknown error'));
  }

  private resetSim() {
    this.isSimulating = false;
    this.simProgress = 0;
    this.simStatus = '';
  }

  private updatePnlChart(pnlCurve: number[]) {
    const labels = Array.from({length: pnlCurve.length}, (_,i) => `W${i+1}`);
    
    // Calculate PnL deltas for the bar chart
    const perTrade = pnlCurve.map((v, i) => i === 0 ? v : v - pnlCurve[i-1]);

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
            data: pnlCurve as any,
            borderColor: '#4f8ef7',
            borderWidth: 2,
            pointRadius: 0,
            fill: true,
            backgroundColor: 'rgba(79,142,247,0.1)',
            tension: 0.3,
            yAxisID: 'y',
            order: 1
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
          }
        }
      }
    };
  }
}

