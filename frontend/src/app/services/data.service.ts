import { Injectable } from '@angular/core';
import { 
  VN30F_RAW, MODELS_PERF, DEPLOYMENT, VARIABLES, 
  FACTORS, GEN_DATASETS, QR_METRICS, DS_SCALARS, 
  DS_CONFUSION, DS_STATS_DATA, QR_CODE, ROLL_WINDOWS 
} from './data.mocks';

// PRNG
function lcg(seed: number) {
  let s = seed >>> 0;
  return () => { s = (Math.imul(1664525, s) + 1013904223) >>> 0; return s / 4294967296; };
}

export interface Tick {
  ts: string; tsDate: string;
  bid1: number; bid2: number; bid3: number;
  ask1: number; ask2: number; ask3: number;
  bq1: number; bq2: number; bq3: number;
  aq1: number; aq2: number; aq3: number;
  obi: number; depth: number; rr: number; label: number;
  mid: number; spread: number;
}

function buildVN30Ticks(raw: any[], seed: number): Tick[] {
  const rng = lcg(seed);
  return raw.map((r, i) => {
    const { ts, b1, b2, b3, a1, a2, a3 } = r;
    const bq1 = 1 + Math.floor(rng() * 50);
    const bq2 = 1 + Math.floor(rng() * 30);
    const bq3 = 1 + Math.floor(rng() * 20);
    const aq1 = 1 + Math.floor(rng() * 50);
    const aq2 = 1 + Math.floor(rng() * 30);
    const aq3 = 1 + Math.floor(rng() * 20);
    const WB = bq1 + bq2 + bq3, WA = aq1 + aq2 + aq3;
    const obi = (WB - WA) / (WB + WA);
    const depth = WA / WB;
    const rr = i >= 10 ? (a1 - raw[i - 10].a1) / raw[i - 10].a1 * 100 : 0;
    const label = i < raw.length - 10
      ? (raw[i + 10].a1 > a1 ? 1 : 0)
      : (rng() < 0.5 ? 1 : 0);
    return {
      ts: ts.slice(11),
      tsDate: ts.slice(0, 10),
      bid1: b1, bid2: b2, bid3: b3,
      ask1: a1, ask2: a2, ask3: a3,
      bq1, bq2, bq3, aq1, aq2, aq3,
      obi: +obi.toFixed(4), depth: +depth.toFixed(4),
      rr: +rr.toFixed(4), label,
      mid: +((b1 + a1) / 2).toFixed(2),
      spread: +(a1 - b1).toFixed(2),
    };
  });
}

const TICKS = buildVN30Ticks(VN30F_RAW, 42);
const TICKS_SAMPLE = TICKS.slice(0, 80);

export interface PredActualData {
  probs: number[];
  labels: number[];
}

export interface PnLData {
  perTrade: number[];
  equity: number[];
  tradeCount: number;
  wins: number;
}

function genHFTPnL(seed: number, n = 300, winRate = 0.56, spreadBps = 10): PnLData {
  const rng = lcg(seed);
  const perTrade: number[] = [];
  const equity: number[] = [1.0];
  let tradeCount = 0, wins = 0;

  for (let i = 0; i < n; i++) {
    const spread = spreadBps * (0.8 + rng() * 0.4);
    const submission = rng() < 0.60 ? 1 : 0;
    const trueValue = rng() < 0.52 ? 1 : 0;

    let pnl = 0;
    if (submission === trueValue) {
      if (submission === 1) {
        pnl = +spread.toFixed(2);
        wins++; tradeCount++;
      }
    } else {
      if (submission === 1) {
        pnl = -spread.toFixed(2);
        tradeCount++;
      }
    }
    perTrade.push(+pnl.toFixed(2));
    equity.push(+(equity[equity.length - 1] * (1 + pnl / 10000)).toFixed(6));
  }
  return { perTrade, equity, tradeCount, wins };
}

function genRollingAcc(seed: number, n = 60, base = 0.60): number[] {
  const rng = lcg(seed);
  return Array.from({ length: n }, (_, i) => +(base + (rng() - 0.5) * 0.12 + Math.sin(i / 8) * 0.02).toFixed(4));
}

const ROLL_ACC = {
  'RandomForest': genRollingAcc(11, 60, 0.635),
  'ExtraTrees': genRollingAcc(22, 60, 0.628),
  'GradBoost': genRollingAcc(33, 60, 0.619),
  'AdaBoost': genRollingAcc(44, 60, 0.601),
  'SVC': genRollingAcc(55, 60, 0.584),
};

export interface Strategy {
  id: string; author: string; model: string; weights: string; lb: string; status: string;
  totalPnL: number; sharpe: number; winRate: number; maxDD: number; trades: number; rollPnL: number[];
}

function stratPnLSummary(seed: number, n = 200, winRate = 0.56, spreadBps = 10) {
  const pnl = genHFTPnL(seed, n, winRate, spreadBps);
  const pt = pnl.perTrade;
  let running = 0;
  const cum = pt.map(v => (running += v, +running.toFixed(2)));
  const total = cum[cum.length - 1];
  let peak = -Infinity, maxDD = 0;
  for (const v of cum) {
    if (v > peak) peak = v;
    const dd = peak - v;
    if (dd > maxDD) maxDD = dd;
  }
  const traded = pt.filter(v => v !== 0);
  const avg = traded.length > 0 ? traded.reduce((s, v) => s + v, 0) / traded.length : 0;
  const std = Math.sqrt(traded.reduce((s, v) => s + (v - avg) ** 2, 0) / (traded.length || 1));
  const sharpe = std > 0 ? +(avg / std * Math.sqrt(2340)).toFixed(2) : 0;
  const step = Math.max(1, Math.floor(cum.length / 30));
  const rollPnL = cum.filter((_, i) => i % step === 0);
  return {
    totalPnL: +total.toFixed(1),
    sharpe,
    winRate: pnl.tradeCount > 0 ? +(pnl.wins / pnl.tradeCount).toFixed(3) : 0,
    maxDD: +maxDD.toFixed(1),
    trades: pnl.tradeCount,
    rollPnL,
  };
}

const STRATEGIES: Strategy[] = [
  { id: 'STRAT-001', author: 'J. Smith', model: 'RandomForest', weights: '(1,1,1)', lb: '30s', status: 'Active', ...stratPnLSummary(1001, 200, 0.580, 10) },
  { id: 'STRAT-002', author: 'M. Chen', model: 'ExtraTrees', weights: '(9,1,0)', lb: '60s', status: 'Active', ...stratPnLSummary(1002, 200, 0.570, 10) },
  { id: 'STRAT-003', author: 'A. Patel', model: 'GradBoost', weights: '(5,3,2)', lb: '30s', status: 'Active', ...stratPnLSummary(1003, 200, 0.565, 10) },
  { id: 'STRAT-004', author: 'L. Kim', model: 'RandomForest', weights: '(7,2,1)', lb: '120s', status: 'Paused', ...stratPnLSummary(1004, 200, 0.555, 10) },
  { id: 'STRAT-005', author: 'R. Torres', model: 'AdaBoost', weights: '(8,2,0)', lb: '60s', status: 'Active', ...stratPnLSummary(1005, 200, 0.548, 10) },
  { id: 'STRAT-006', author: 'J. Smith', model: 'SVC', weights: '(1,0,0)', lb: '10s', status: 'Paused', ...stratPnLSummary(1006, 200, 0.544, 10) },
  { id: 'STRAT-007', author: 'M. Chen', model: 'ExtraTrees', weights: '(0,1,0)', lb: '60s', status: 'Archived', ...stratPnLSummary(1007, 200, 0.540, 10) },
  { id: 'STRAT-008', author: 'A. Patel', model: 'GradBoost', weights: '(6,4,0)', lb: '30s', status: 'Active', ...stratPnLSummary(1008, 200, 0.561, 10) },
  { id: 'STRAT-009', author: 'L. Kim', model: 'RandomForest', weights: '(5,5,0)', lb: '120s', status: 'Active', ...stratPnLSummary(1009, 200, 0.553, 10) },
  { id: 'STRAT-010', author: 'R. Torres', model: 'AdaBoost', weights: '(2,3,5)', lb: '30s', status: 'Archived', ...stratPnLSummary(1010, 200, 0.537, 10) },
];

const STRAT_LABELS = STRATEGIES.map(s => s.id);
function buildStratCorr(strats: Strategy[]) {
  const n = strats.length;
  const rng = lcg(24680);
  const m = Array.from({ length: n }, () => new Array(n).fill(0));
  for (let i = 0; i < n; i++) {
    m[i][i] = 1;
    for (let j = i + 1; j < n; j++) {
      const si = strats[i], sj = strats[j];
      let base = 0.05 + rng() * 0.15;
      if (si.model === sj.model) base += 0.35;
      if (si.weights === sj.weights) base += 0.25;
      if (si.author === sj.author) base += 0.12;
      if (si.lb === sj.lb) base += 0.08;
      if ((si.status === 'Archived') !== (sj.status === 'Archived')) base -= 0.10;
      const c = +Math.max(-0.35, Math.min(0.98, base + (rng() - 0.5) * 0.08)).toFixed(3);
      m[i][j] = c; m[j][i] = c;
    }
  }
  return m;
}
const STRAT_CORR = buildStratCorr(STRATEGIES);

@Injectable({
  providedIn: 'root'
})
export class DataService {

  getTicks(): Tick[] {
    return TICKS;
  }

  getTicksSample(): Tick[] {
    return TICKS_SAMPLE;
  }

  getStrategies(): Strategy[] {
    return STRATEGIES;
  }

  getStratCorr() {
    return STRAT_CORR;
  }

  getStratLabels() {
    return STRAT_LABELS;
  }

  getModelsPerf() {
    return MODELS_PERF;
  }

  getDeployment() { return DEPLOYMENT; }

  getDsPredActual(): PredActualData {
    const rng = lcg(12345);
    const probs = Array.from({ length: 80 }, () => +(0.2 + rng() * 0.6).toFixed(3));
    const labels = probs.map(p => (p > 0.5 ? 1 : 0));
    // Add some "surprise" errors
    for (let i = 0; i < 10; i++) {
      const idx = Math.floor(rng() * 80);
      labels[idx] = (1 - labels[idx]) as (0 | 1);
    }
    return { probs, labels: labels as (0 | 1)[] };
  }

  getVariables() {
    return VARIABLES;
  }

  getFactors() {
    return FACTORS;
  }

  getQrMetrics() {
    return QR_METRICS;
  }

  getDsScalars() {
    return DS_SCALARS;
  }

  getDsConfusion() {
    return DS_CONFUSION;
  }

  getDsStatsData() {
    return DS_STATS_DATA;
  }

  getGenDatasets() {
    return GEN_DATASETS;
  }

  getGenDataset(name: string) {
    return GEN_DATASETS.find(d => d.name === name);
  }

  getRollWindows() {
    return ROLL_WINDOWS;
  }

  getRollAcc() {
    return ROLL_ACC;
  }

  genHFTPnL(seed: number, n: number, winRate: number, spreadBps: number): PnLData {
    return genHFTPnL(seed, n, winRate, spreadBps);
  }

  getPnlInit(): PnLData {
    return genHFTPnL(42, 300, 0.564, 5);
  }
}
export { QR_CODE };
