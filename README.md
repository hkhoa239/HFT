# Quantitative Alpha Research & Development Lab — Dashboard Platform

A single-page, multi-workspace SaaS dashboard modeled on WorldQuant Brain, built in pure HTML/CSS/JS with no external dependencies except CDN-hosted libraries.

---

## Overview

A **strictly dark-mode**, data-dense, IDE-flavored dashboard partitioned into three role-specific workspaces:

| Workspace | Audience | Core Value |
|---|---|---|
| **Quant Researcher (QR)** | Researchers | Write & backtest alpha expressions |
| **Data Scientist (DS)** | Data Scientists | Train ML models, publish indicators |
| **Portfolio Manager (PM)** | PMs | Monitor, rank, and allocate across alphas |

---

## Tech Stack

| Concern | Choice | Rationale |
|---|---|---|
| Structure | **HTML5** (single file `index.html`) | Zero build step, deployable |
| Styling | **Vanilla CSS** (embedded `<style>`) | Full control, theme tokens |
| Logic | **Vanilla JS** (embedded `<script>`) | No bundler needed |
| Code Editor | **CodeMirror 5** (CDN) | Lightweight, syntax highlighting out-of-the-box |
| Charts | **Chart.js 4** (CDN) | Line, bar, scatter, heatmap |
| Fonts | **Roboto Mono + Space Mono** (Google Fonts) | Monospace alignment |
| Sparklines | **Custom SVG sparklines** (vanilla JS) | Zero overhead |

> [!NOTE]
> All data is **mocked/synthetic** — realistic-looking historical figures generated in JS to demonstrate a live feel without a backend.

---

## Design System

### Color Tokens
```
--bg-base:      #0d0f14   /* deepest background */
--bg-surface:   #13161e   /* panels */
--bg-elevated:  #1a1d28   /* cards, editors */
--bg-hover:     #20243a
--border:       #252836
--accent-blue:  #4f8ef7   /* primary CTA */
--accent-green: #22d3a0   /* positive PnL */
--accent-red:   #f45c5c   /* negative / drawdown */
--accent-amber: #f5a623   /* warnings / neutral */
--text-primary: #e8eaf0
--text-muted:   #6b7280
--text-code:    #a8d8ea
```

### Typography
- **UI labels / headings**: `'Roboto Mono', monospace`
- **Code editors**: `'Space Mono', monospace`
- **Numbers**: `'Roboto Mono'` with tabular-nums

### Spacing & Layout
- CSS Grid for the outer shell (sidebar + main area)
- CSS Grid 12-column within each workspace for panels
- Panels have a dark border, subtle inner shadow, rounded corners (4px)

---

## Proposed File Structure

```
assignment_1/
└── index.html          ← entire application (HTML + CSS + JS)
```

Single self-contained file for zero-friction delivery.

---

## Workspace Breakdown & Components

### 1. Global Shell
- **Top Navigation Bar**: Logo | Workspace Switcher tabs (QR / DS / PM) | User avatar | Clock
- **Left Sidebar** (collapsible): Workspace-scoped navigation icons
- **Status Bar** (bottom): Last backtest run time | Data freshness | System health dot

---

### 2. Quant Researcher (QR) Workspace

```
┌──────────────────────────┬─────────────────────┐
│  Fast Expression Editor  │  Data Variable Lib  │
│  (CodeMirror, ~60% wide) │  (searchable list)  │
├──────────────────────────┴─────────────────────┤
│  Simulation Control Panel (full width strip)   │
├──────────────────────────┬─────────────────────┤
│  Cumulative PnL Chart    │  Metrics Grid       │
└──────────────────────────┴─────────────────────┘
```

**Widget 1 – Fast Expression Editor**
- CodeMirror instance with Python syntax highlighting
- Pre-loaded example alpha expression
- Toolbar: Run | Save | Format | Share
- Line numbers, bracket matching, auto-close brackets
- Gutter with error indicators

**Widget 2 – Data Variable Library**
- Search box at top
- Categorized list: `PRICE`, `VOLUME`, `ORDER BOOK`, `ML OUTPUTS`
- Clicking a variable inserts it into the editor
- Badge showing "Published by DS" on model outputs

**Widget 3 – Simulation Control**
- Inline form: Universe selector (US Equity / US ETF), Date Range picker (1Y / 3Y / 5Y), Cost Model (Low / Mid / High)
- **"▶ RUN SIMULATION"** button (accent blue, animated pulse while running)
- Progress bar + ETA during mock execution

**Widget 4 – Performance Analysis**
- Cumulative PnL Chart (Chart.js line, green/red fill)
- Metrics grid: Sharpe | Max DD | Turnover | IR | Win Rate | Calmar
- Drawdown sub-chart overlay

---

### 3. Data Scientist (DS) Workspace

```
┌──────────────────┬───────────────────────────────┐
│ Dataset Explorer │  Model Trainer IDE             │
│ (stats + dist.)  │  (CodeMirror Python)           │
├──────────────────┼───────────────────────────────┤
│ Model Evaluation │  Feature & Signal Publisher   │
│ (ROC, Confusion) │  (published model library)    │
└──────────────────┴───────────────────────────────┘
```

**Widget 1 – Dataset Explorer**
- Dataset selector dropdown
- Descriptive stats table (mean, std, skew, min/max)
- Distribution histogram (Chart.js bar)
- Correlation mini-matrix (CSS grid colored cells)

**Widget 2 – Model Trainer IDE**
- CodeMirror with Python mode
- Pre-loaded `RandomForestClassifier` training template
- Buttons: Train | Evaluate | Export
- Console output panel below editor (dark terminal)

**Widget 3 – Model Evaluation Console**
- ROC Curve (Chart.js line)
- Confusion Matrix (4-cell grid with color intensity)
- Scalar metrics: Accuracy | Precision | Recall | F1 | AUC

**Widget 4 – Feature & Signal Publisher**
- Table of published indicators: name, type, author, date, target
- **"↑ Publish to QR"** button per row with toast confirmation
- Status badges: Draft | Under Review | Published | Deprecated

---

### 4. Portfolio Manager (PM) Workspace

```
┌────────────────────────────┬──────────────────────┐
│  Alpha Leaderboard         │  Correlation Heatmap │
│  (searchable, sortable)    │  (N×N color grid)    │
├────────────────────────────┼──────────────────────┤
│  Researcher Performance    │  Portfolio Alloc.    │
│  Grid (ranked table)       │  (donut / bar chart) │
└────────────────────────────┴──────────────────────┘
```

**Widget 1 – Alpha Leaderboard**
- Sortable columns: Alpha ID | Author | Sharpe | Max DD | Turnover | Status
- Sparkline PnL column (SVG sparklines)
- Search / filter bar
- Status chip: Active | Paused | Archived

**Widget 2 – Correlation Heatmap**
- N×N grid (10×10 mock alphas)
- HSL color mapping: red (−1) → gray (0) → green (+1)
- Tooltip on hover with exact correlation value

**Widget 3 – Researcher Performance Grid**
- Ranked table: Rank | Researcher | # Alphas | Best Sharpe | Avg Sharpe | Total PnL
- Avatar initials + color badge per researcher
- Mini bar chart of Sharpe scores

**Widget 4 – Portfolio Allocation**
- Donut chart (Chart.js) — capital weight per alpha bucket
- Table underneath: Alpha | Weight | Allocation $ | Contribution

---

## Animations & Micro-interactions

- **Workspace switch**: slide-in transition on panels
- **Run Simulation**: button pulse → progress bar fill → results fade-in
- **Chart.js**: default animation on data load
- **Hover on table rows**: subtle `--bg-hover` highlight + left border accent
- **Panel load**: staggered `opacity 0 → 1` with small `translateY`
- **Tooltip**: smooth fade on heatmap cells

---

## Verification Plan

### Automated
- Open `index.html` in browser via browser subagent
- Verify all three workspace tabs render without console errors
- Verify charts display populated with data
- Verify CodeMirror editors are interactive

### Manual
- Switch between QR / DS / PM workspaces
- Click "Run Simulation" and observe progress + results
- Search variables in the Data Library
- Click a row in Alpha Leaderboard and verify row highlight
- Hover heatmap cells for tooltip

---

## Open Questions

> [!IMPORTANT]
> **Q1**: Should the three workspaces be separate HTML pages (routes) or a single-page tab-switch? *(Default: single-page tab-switch for zero-backend simplicity)*

> [!IMPORTANT]
> **Q2**: Should the code editors be read-only demos or fully interactive? *(Default: fully interactive — CodeMirror allows editing)*

> [!NOTE]
> **Q3**: Any specific alpha expression language syntax beyond Python? *(WorldQuant Brain uses a domain-specific expression language but Python is more familiar)*
