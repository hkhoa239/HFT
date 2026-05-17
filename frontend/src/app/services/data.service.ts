import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { APP_CONFIG } from '../app.constants';
import { VARIABLES } from './code-templates';

export interface Strategy {
  id: string;
  author: string;
  model: string;
  weights: string;
  lb: number;
  status: string;
  totalPnL: number;
  sharpe: number;
  winRate: number;
  maxDD: number;
  trades: number;
  rollPnL: number[];
}

@Injectable({
  providedIn: 'root'
})
export class DataService {
  private http = inject(HttpClient);
  private apiUrl = APP_CONFIG.apiUrl;

  // Real API calls
  getPerformance(): Observable<Strategy[]> {
    return this.http.get<any>(`${this.apiUrl}/analytics/performance`).pipe(
      map(res => {
        if (!res.success || !res.data) return [];
        return res.data.map((item: any) => ({
          id: item.alpha_name || 'Unknown',
          author: item.author_name || 'System',
          model: 'HFT-Alpha',
          weights: 'dynamic',
          lb: 60,
          totalPnL: item.total_return ?? item.total_pnl ?? 0,
          sharpe: item.sharpe ?? item.sharpe_ratio ?? 0,
          winRate: item.win_rate ?? 0,
          maxDD: Math.abs(item.max_drawdown ?? 0),
          status: item.status === 'completed' ? 'Active' : 'Archived',
          rollPnL: this.extractPnLArray(item.pnl_curve)
        }));
      }),
      catchError(() => of([]))
    );
  }

  getCorrelation(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/analytics/correlation`).pipe(
      map(res => res.success ? res.data : null),
      catchError(() => of(null))
    );
  }

  getModels(): Observable<any[]> {
    return this.http.get<any>(`${this.apiUrl}/models`).pipe(
      map(res => {
        if (!res.success || !res.data) return [];
        return res.data.map((m: any, i: number) => ({
          model: m.name,
          accMean: m.training_metrics?.accuracy ?? 0.64,
          accStd: m.training_metrics?.acc_std ?? 0.02,
          f1Mean: m.training_metrics?.f1_score ?? 0.63,
          bestWin: m.version || ('W' + (i + 1)),
          color: ['#4f8ef7', '#22d3a0', '#f5a623', '#f45c5c', '#a78bfa'][i % 5]
        }));
      }),
      catchError(() => of([]))
    );
  }

  getFactors(): Observable<any[]> {
    return this.http.get<any>(`${this.apiUrl}/factors`).pipe(
      map(res => (res.success && res.data) ? res.data : []),
      catchError(() => of([]))
    );
  }

  getAuditLogs(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/audit-logs`).pipe(
      map(res => res.success ? res.data : { data: [], total: 0 }),
      catchError(() => of({ data: [], total: 0 }))
    );
  }

  seedData(): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/seed`, {});
  }

  getVariables() {
    return VARIABLES;
  }

  // Utils
  private extractPnLArray(curve: any[] | undefined): number[] {
    if (!curve || curve.length === 0) return [0, 0, 1, 2, 3];
    return curve.map(p => p.v ?? p.cumPnL ?? 0);
  }

  // Visual-only helpers for non-core widgets (outside A-F core flows).
  getTicks(): any[] {
    return Array.from({ length: 50 }, (_, i) => ({
      ts: `10:00:${i.toString().padStart(2, '0')}`,
      bid1: 1450.5 + Math.sin(i / 8) * 0.6,
      bid2: 1450.4 + Math.sin(i / 8) * 0.5,
      bid3: 1450.3 + Math.sin(i / 8) * 0.4,
      ask1: 1451.0 + Math.sin(i / 8) * 0.6,
      ask2: 1451.1 + Math.sin(i / 8) * 0.5,
      ask3: 1451.2 + Math.sin(i / 8) * 0.4,
      bq1: 20 + (i % 9),
      bq2: 18 + (i % 7),
      bq3: 16 + (i % 5),
      aq1: 21 + (i % 8),
      aq2: 19 + (i % 6),
      aq3: 17 + (i % 4),
      mid: 1450.75,
      spread: 0.5
    }));
  }

  getTicksSample(): any[] { return this.getTicks().slice(0, 20); }

  getDsScalars() {
    return [
      { l: 'Precision', v: '0.652', s: 'UP signal accuracy' },
      { l: 'Recall', v: '0.613', s: 'UP signal capture' },
      { l: 'F1-Score', v: '0.631', s: 'Harmonic mean' },
      { l: 'Log-Loss', v: '0.584', s: 'Entropy penalty' }
    ];
  }

  getGenDatasets() {
    return [
      { name: 'VN30F2112_clean', type: 'CSV', rows: 1421802, date: '2024-05-10', status: 'Published' },
      { name: 'Alpha_Factors_v1', type: 'Parquet', rows: 840200, date: '2024-05-12', status: 'Draft' }
    ];
  }

  getRollWindows() { return Array.from({ length: 20 }, (_, i) => `W${i + 1}`); }

  getRollAcc() {
    return {
      'RandomForest': Array.from({ length: 20 }, () => 0.6 + Math.random() * 0.1),
      'XGBoost': Array.from({ length: 20 }, () => 0.62 + Math.random() * 0.08)
    };
  }

  getDsPredActual() {
    return {
      probs: Array.from({ length: 80 }, () => Math.random()),
      labels: Array.from({ length: 80 }, () => Math.random() > 0.5 ? 1 : 0)
    };
  }
}
export { QR_CODE, DS_CODE, DS_ANALY_CODE, DS_GEN_CODE } from './code-templates';
