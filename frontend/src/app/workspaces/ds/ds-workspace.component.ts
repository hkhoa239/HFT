import { Component, OnInit, OnDestroy, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DataService } from '../../services/data.service';
import { NotificationService } from '../../services/notification.service';
import { ChartWrapperComponent } from '../../shared/components/chart-wrapper.component';
import { EditorWrapperComponent } from '../../shared/components/editor-wrapper.component';
import { ChartConfiguration } from 'chart.js';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-ds-workspace',
  standalone: true,
  imports: [CommonModule, ChartWrapperComponent, EditorWrapperComponent, FormsModule],
  templateUrl: './ds-workspace.component.html',
  styleUrl: './ds-workspace.component.scss'
})
export class DsWorkspaceComponent implements OnInit, OnDestroy {
  activeDsTab = 'analyze';

  // Data
  dsStats: { label: string, val: string }[] = [];

  // Editors
  analyCode = '';
  trainCode = '';
  genCode = '';

  // Charts
  obConfig?: ChartConfiguration;
  qtyConfig?: ChartConfiguration;
  rollConfig?: ChartConfiguration;
  predConfig?: ChartConfiguration;

  // Train Eval
  confusion: any[] = [];
  scalars: any[] = [];

  // Console
  consoleLines: { m: string, c: string }[] = [];
  analyLines: { m: string, c: string }[] = [];
  genLines: { m: string, c: string }[] = [];

  // Registry
  registry: any[] = [];
  selectedDataset: any = null;
  registryStats = '0 datasets · 0 published';
  showDialog = false;
  ticks: any[] = [];

  // Console/Logs
  logs: string[] = [];
  private logTimer: any;

  constructor(
    public dataService: DataService,
    public ns: NotificationService,
    private cd: ChangeDetectorRef
  ) { }

  ngOnInit() {
    this.dsStats = [
      { label: 'Symbol', val: 'VN30F2112' },
      { label: 'Record Count', val: '1,421,802' },
      { label: 'Mean Spread', val: '0.12 ticks' },
      { label: 'Tick Size', val: '0.1 index pts' },
      { label: 'L3 Depth', val: '42 levels' },
      { label: 'Avg Latency', val: '0.34ms' },
      { label: 'Update Freq', val: '2.4/sec' }
    ];

    import('../../services/data.service').then(m => {
      this.analyCode = m.DS_ANALY_CODE;
      this.trainCode = m.DS_CODE;
      this.genCode = m.DS_GEN_CODE;
    });

    this.confusion = [
      { l: 'TP', v: '3,214', bg: 'rgba(34,211,160,0.20)' },
      { l: 'FP', v: '1,892', bg: 'rgba(244,92,92,0.15)' },
      { l: 'FN', v: '2,108', bg: 'rgba(244,92,92,0.15)' },
      { l: 'TN', v: '2,786', bg: 'rgba(34,211,160,0.20)' },
    ];

    this.scalars = this.dataService.getDsScalars();
    this.registry = this.dataService.getGenDatasets();
    this.updateRegistryStats();

    this.initConsole();
    this.initCharts();
  }

  updateRegistryStats() {
    const pub = this.registry.filter(d => d.status === 'Published').length;
    this.registryStats = `${this.registry.length} datasets · ${pub} published`;
  }

  initConsole() {
    this.consoleLines = [{ m: 'QuantAlpha HFT Console v1.0.0', c: '#22d3a0' }, { m: 'Ready — click Train to start rolling-window pipeline.', c: '#6b7280' }];
    this.analyLines = [{ m: 'Ready — click Run to start analysis.', c: '#6b7280' }];
    this.genLines = [{ m: 'Data Generator v1.0', c: '#22d3a0' }, { m: 'Set DATASET_NAME then click ▶ Run.', c: '#6b7280' }];
  }

  ngOnDestroy() {
    if (this.logTimer) clearInterval(this.logTimer);
  }

  setDsTab(tab: string) {
    this.activeDsTab = tab;
  }

  viewData(d: any) {
    this.selectedDataset = d;
    this.showDialog = true;
  }

  publishFactor(f: any) {
    f.status = 'Published';
    this.ns.success(`"${f.name}" published to QR variable library`);
    this.analyLines.push({ m: `> Publishing factors to QR Variable Library…`, c: '#4f8ef7' });
    setTimeout(() => {
      this.analyLines.push({ m: `  ✓ ${f.name} published. VERSION: v2.`, c: '#22d3a0' });
    }, 400);
  }

  publishRegistry(r: any) {
    r.status = 'Published';
    this.updateRegistryStats();
    this.ns.success(`'${r.name}' published to QR`);
  }

  runGenerator() {
    this.genLines = [
      { m: 'Data Generator v1.0', c: '#22d3a0' },
      { m: 'Set DATASET_NAME then click ▶ Run.', c: '#6b7280' }
    ];
    const steps = [
      ['> Running generator script…', '#4f8ef7'],
      ['  Loading VN30F2112_ob3.csv (41,496 rows)…', '#6b7280'],
      ['  Computing OBI_532_v2…', '#6b7280'],
      ['  Validating output shape (41,496 × 1)…', '#a8d8ea'],
      ['  mean: 0.009  std: 0.315  min: -0.979  max: 0.986', '#22d3a0'],
      ['  Generated 41,496 rows for \'OBI_532_v2\'', '#22d3a0'],
      ['✓ Dataset registered — click name or View to preview.', '#22d3a0'],
    ];

    let i = 0;
    const iv = setInterval(() => {
      if (i < steps.length) {
        this.genLines.push({ m: steps[i][0], c: steps[i][1] });
        i++;
        this.cd.detectChanges();
      } else {
        clearInterval(iv);
        if (!this.registry.find(x => x.name === 'OBI_532_v2')) {
          this.registry.push({
            name: 'OBI_532_v2', type: 'OBI Factor', rows: 41496,
            date: new Date().toISOString().split('T')[0], status: 'Draft',
            mean: 0.009, std: 0.315, min: -0.979, max: 0.986, author: 'You'
          });
          this.updateRegistryStats();
        }
        this.ns.success("'OBI_532_v2' generated — Draft registered");
      }
    }, 200);
  }

  runAnalysis() {
    this.analyLines = [
      { m: 'Ready — click Run to start analysis.', c: '#6b7280' }
    ];
    const steps = [
      ['> Loading VN30F2112_ob3.csv…', '#4f8ef7'],
      ['  41,496 snapshots × 12 columns loaded.', '#6b7280'],
      ['  Computing OBI_111, OBI_910, OBI_532…', '#6b7280'],
      ['  Computing rise_10s, rise_30s, rise_60s…', '#6b7280'],
      ['  Labelling UP/DOWN (next 10 s)…', '#a8d8ea'],
      ['', ''],
      ['  UP rate: 51.82%', '#22d3a0'],
      ['  Spread mean: 1.4823', '#22d3a0'],
      ['  OBI_111 std : 0.3218', '#22d3a0'],
      ['', ''],
      ['✓ Factors computed — click "To QR" to send.', '#22d3a0'],
    ];

    let i = 0;
    const iv = setInterval(() => {
      if (i < steps.length) {
        if (steps[i][0] || steps[i][1]) {
          this.analyLines.push({ m: steps[i][0], c: steps[i][1] });
        } else {
          this.analyLines.push({ m: '', c: '' });
        }
        i++;
        this.cd.detectChanges();
      } else {
        clearInterval(iv);
        this.ns.success('Analysis complete');
      }
    }, 160);
  }

  trainModel() {
    this.consoleLines = [
      { m: 'QuantAlpha HFT Console v1.0.0', c: '#22d3a0' },
      { m: 'Ready — click Train to start rolling-window pipeline.', c: '#6b7280' }
    ];
    const steps = [
      ['> Initializing HFT rolling-window pipeline…', '#4f8ef7'],
      ['  Loading VN30F2112.csv (41,496 snapshots, 300 sampled)…', '#6b7280'],
      ['  Computing 64 features (30 rise_ratio + 34 OBI/depth)…', '#6b7280'],
      ['  Rolling windows: 840 iterations (30min train, 10s pred)…', '#a8d8ea'],
      ['  RandomForest   — mean acc: 0.6483, F1: 0.6310', '#22d3a0'],
      ['  ExtraTrees     — mean acc: 0.6381, F1: 0.6194', '#22d3a0'],
      ['  SVC (RBF)      — mean acc: 0.5884, F1: 0.5672', '#f5a623'],
      ['✓ Training complete.', '#22d3a0'],
    ];

    let i = 0;
    const iv = setInterval(() => {
      if (i < steps.length) {
        this.consoleLines.push({ m: steps[i][0], c: steps[i][1] });
        i++;
        this.cd.detectChanges();
      } else {
        clearInterval(iv);
        this.confusion = [
          { l: 'TP', v: '3,214', bg: 'rgba(34,211,160,0.20)' },
          { l: 'FP', v: '1,892', bg: 'rgba(244,92,92,0.15)' },
          { l: 'FN', v: '2,108', bg: 'rgba(244,92,92,0.15)' },
          { l: 'TN', v: '2,786', bg: 'rgba(34,211,160,0.20)' },
        ];
        this.scalars = this.dataService.getDsScalars();
        this.cd.detectChanges();
        this.ns.success('HFT model trained — RF Accuracy: 64.8%');
      }
    }, 280);
  }

  evaluateModel() {
    this.consoleLines.push({ m: '> Evaluating RandomForest on held-out windows…', c: '#4f8ef7' });
    setTimeout(() => {
      this.consoleLines.push({ m: '  AUC: 0.706  Precision: 0.652  Recall: 0.613  F1: 0.631', c: '#22d3a0' });
      this.consoleLines.push({ m: '  Top-5 features: rise_30s, OBI_111, OBI_910, rise_10s, depth_532', c: '#a8d8ea' });
      this.consoleLines.push({ m: '✓ Evaluation complete.', c: '#22d3a0' });
      this.ns.success('Evaluation sequence complete');
    }, 500);
  }

  exportModel() {
    this.consoleLines.push({ m: '> Exporting model to registry…', c: '#4f8ef7' });
    setTimeout(() => {
      this.consoleLines.push({ m: '✓ Saved: /registry/models/rf_hft_ob3_v1.pkl', c: '#22d3a0' });
      this.ns.info('Model exported to registry');
    }, 500);
  }

  private initCharts() {
    const ticks = this.dataService.getTicksSample();

    // OB Chart
    this.obConfig = {
      type: 'line',
      data: {
        labels: ticks.map(t => t.ts),
        datasets: [
          { label:'Ask2', data: ticks.map(t=>t.ask2), borderColor:'rgba(244,92,92,0.35)', borderWidth:1, pointRadius:0, fill:false, borderDash:[2,3], tension:0.2 },
          { label:'Ask1', data: ticks.map(t=>t.ask1), borderColor:'#f45c5c', borderWidth:1.5, pointRadius:0, fill:true, backgroundColor:'rgba(244,92,92,0.05)', tension:0.2 },
          { label:'Bid1', data: ticks.map(t=>t.bid1), borderColor:'#22d3a0', borderWidth:1.5, pointRadius:0, fill:true, backgroundColor:'rgba(34,211,160,0.05)', tension:0.2 },
          { label:'Bid2', data: ticks.map(t=>t.bid2), borderColor:'rgba(34,211,160,0.35)', borderWidth:1, pointRadius:0, fill:false, borderDash:[2,3], tension:0.2 }
        ]
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: { display: true, labels: { color: '#6b7280', font: { size: 8, family: 'Roboto Mono' }, boxWidth: 10, padding: 8 } },
          tooltip: { backgroundColor: '#1a1d28', borderColor: '#252836', borderWidth: 1, titleColor: '#4f8ef7', bodyColor: '#e8eaf0', bodyFont: { size: 10 }, usePointStyle: true }
        },
        scales: {
          x: { grid: { color: '#151820' }, ticks: { color: '#3d4460', maxTicksLimit: 8, font: { size: 8 } }, title: { display: true, text: 'Time (09:00 - 09:30)', color: '#3d4460', font: { size: 8, family: 'Inter' } } },
          y: { grid: { color: '#151820' }, ticks: { color: '#3d4460', font: { size: 8 } }, title: { display: true, text: 'Price', color: '#3d4460', font: { size: 8, family: 'Inter' } } }
        }
      }
    };

    // Qty Chart
    this.qtyConfig = {
      type: 'bar',
      data: {
        labels: ticks.map(t => t.ts),
        datasets: [
          { label: 'B3', data: ticks.map(t => t.bq3), backgroundColor: 'rgba(34,211,160,0.2)', borderWidth: 0, stack: 'b' },
          { label: 'B2', data: ticks.map(t => t.bq2), backgroundColor: 'rgba(34,211,160,0.4)', borderWidth: 0, stack: 'b' },
          { label: 'B1', data: ticks.map(t => t.bq1), backgroundColor: 'rgba(34,211,160,0.7)', borderWidth: 0, stack: 'b' },
          { label: 'A1', data: ticks.map(t => -t.aq1), backgroundColor: 'rgba(244,92,92,0.7)', borderWidth: 0, stack: 'a' },
          { label: 'A2', data: ticks.map(t => -t.aq2), backgroundColor: 'rgba(244,92,92,0.4)', borderWidth: 0, stack: 'a' },
          { label: 'A3', data: ticks.map(t => -t.aq3), backgroundColor: 'rgba(244,92,92,0.2)', borderWidth: 0, stack: 'a' }
        ]
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: {
          legend: { display: true, labels: { color: '#6b7280', font: { size: 8, family: 'Roboto Mono' }, boxWidth: 10, padding: 4 } },
          tooltip: { backgroundColor: '#1a1d28', borderColor: '#252836', borderWidth: 1, callbacks: { label: (ctx) => `${ctx.dataset.label}: ${Math.abs(ctx.parsed.y as number)}` } }
        },
        scales: {
          x: { display: false, stacked: true },
          y: { grid: { color: '#151820' }, ticks: { color: '#3d4460', font: { size: 8 }, callback: (v: any) => Math.abs(v) }, stacked: true }
        }
      }
    };

    // Roll Accuracy
    const accData = this.dataService.getRollAcc();
    this.rollConfig = {
      type: 'line',
      data: {
        labels: this.dataService.getRollWindows(),
        datasets: Object.entries(accData).map(([name, data], i) => ({
          label: name, data: data as any,
          borderColor: ['#4f8ef7', '#22d3a0', '#f5a623', '#f45c5c', '#a78bfa'][i],
          borderWidth: 1.5, pointRadius: 0, fill: false, tension: 0.3
        }))
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: {
          legend: { display: true, labels: { color: '#6b7280', boxWidth: 10, font: { size: 8, family: 'Roboto Mono' } } },
          tooltip: { backgroundColor: '#1a1d28', borderColor: '#252836', borderWidth: 1 }
        },
        scales: {
          x: { grid: { color: '#151820' }, ticks: { color: '#3d4460', maxTicksLimit: 10, font: { size: 8 } }, title: { display: true, text: 'Rolling Window', color: '#6b7280', font: { size: 8 } } },
          y: { min: 0.5, max: 0.75, grid: { color: '#151820' }, ticks: { color: '#3d4460', font: { size: 8 }, callback: (v: any) => (v * 100).toFixed(0) + '%' }, title: { display: true, text: 'Accuracy', color: '#6b7280', font: { size: 8 } } }
        }
      }
    };

    // Predicted vs Actual
    const predData = this.dataService.getDsPredActual();
    const actual = predData.labels;
    this.predConfig = {
      type: 'line',
      data: {
        labels: Array.from({ length: 80 }, (_, i) => `W${i + 1}`),
        datasets: [
          {
            label: 'Actual (UP=1 / DOWN=0)',
            data: actual as any,
            borderColor: '#22d3a0', borderWidth: 1.5,
            pointRadius: actual.map(a => a ? 2.5 : 2),
            pointBackgroundColor: actual.map(a => a ? '#22d3a0' : '#f45c5c'),
            stepped: 'before', fill: false, tension: 0
          },
          {
            label: 'RF Predicted Prob',
            data: predData.probs as any,
            borderColor: '#4f8ef7', borderWidth: 1.5,
            pointRadius: 0, tension: 0.4, fill: false
          },
          {
            label: 'Threshold (0.5)',
            data: Array(80).fill(0.5),
            borderColor: 'rgba(245,166,35,0.5)', borderWidth: 1,
            borderDash: [4, 4], pointRadius: 0, fill: false
          }
        ]
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: {
          legend: { display: true, labels: { color: '#6b7280', boxWidth: 10, font: { size: 8, family: 'Roboto Mono' } } }
        },
        scales: {
          x: { grid: { color: '#151820' }, ticks: { color: '#3d4460', maxTicksLimit: 12, font: { size: 8 } } },
          y: { min: 0, max: 1.1, grid: { color: '#151820' }, ticks: { color: '#3d4460', font: { size: 8 } } }
        }
      }
    };
  }
}
