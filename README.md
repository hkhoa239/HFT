# QuantAlpha Lab — HFT Research Platform

> **Course:** [055134] Hệ Thống Thông Minh (Intelligent Systems)  
> **Assignment:** 1 — High-Frequency Trading with Machine Learning  
> **Version:** v2.4.1

---

## Overview

**QuantAlpha Lab** is a high-performance High-Frequency Trading (HFT) Research and Simulation Platform. It consists of a decoupled client-server architecture including an Angular 18 Frontend, a Go Backend API service, and an Asynchronous Python Worker daemon backed by Redis Streams and PostgreSQL.

The platform is organised around **three professional roles** that collaborate in a sequential workflow:

```
DS  (Data Scientist)
 │
 │  analyzes data, trains models, publishes datasets
 ▼
 QR  (Quant Researcher)
 │
 │  uses published data to write & backtest alpha signals
 ▼
 PM  (Portfolio Manager)
     monitors all strategies and manages capital allocation
```

---

## How the Project Works

The project simulates how a real HFT research team operates:

1. The **Data Scientist** works with raw order-book data, computes trading signals (factors), trains machine learning classifiers, and publishes validated datasets to a shared registry.
2. The **Quant Researcher** picks up those published datasets and uses them as inputs to write Python-based alpha expressions. Each expression is backtested against the historical VN30F2112 order-book data, and its performance (PnL, Sharpe, win rate) is recorded as a strategy.
3. The **Portfolio Manager** has a read-only view of all submitted strategies. The PM monitors their performance, checks how correlated they are with each other, and decides how much capital to allocate to each active strategy.

The underlying dataset is the **VN30F2112** — VN30 Index Futures (December 2021 contract, Ho Chi Minh Stock Exchange), covering the period 2021-04-19 → 2021-12-16. Level-3 order-book snapshots (bid/ask prices at 3 depth levels) are used as the raw input.

---

## Role 1 — DS: Data Scientist

The Data Scientist is responsible for turning raw market data into clean, validated datasets and trained models that the Quant Researcher can use.

### Tab 1 — Analyze Data

This tab is for **exploring the raw order-book data** before doing any modelling.

- **Order Book Explorer** — select a data period and view a summary statistics table for the Level-3 order-book. Key metrics shown include spread, Order Book Imbalance (OBI), depth ratio, rise ratio, and the fraction of ticks where the price moved up (UP label rate).
- **Bid/Ask Price Chart** — a time-series chart of the best bid and ask prices over a sample session, so the DS can visually inspect the price dynamics.
- **Bid/Ask Quantity Chart** — a bar chart showing average quantities at order-book depth levels 1, 2, and 3, giving a sense of market liquidity.
- **Analysis Script Editor** — a Python code editor where the DS writes factor computation scripts. Running the script outputs statistics to a console. Once the DS is satisfied with the results, clicking **Publish to QR** makes those factor variables available to the Quant Researcher's variable library.

### Tab 2 — Train Model

This tab is for **training machine learning classifiers** on the processed order-book features.

- **ML Trainer Editor** — a Python code editor showing the rolling-window training pipeline. The training protocol is: use the past 30 minutes of data to train, then predict the next 10 seconds.
- **Rolling Accuracy Chart** — shows how the classifier's accuracy evolves across successive rolling windows, helping the DS assess model stability over time.
- **Confusion Matrix** — breaks down the classifier's predictions into true positives, false positives, true negatives, and false negatives, for a clearer picture of prediction quality beyond accuracy alone.
- **Predicted vs. Actual Chart** — plots the model's predicted probability against the true label across multiple windows, revealing calibration and consistency.
- **Console** — displays training progress and per-window metrics (accuracy, F1 score).

### Tab 3 — Published Data

This tab is the **data registry** — a shared store of datasets and factors that the DS team has generated.

- **Data Generator Editor** — a Python code editor where the DS writes scripts to produce new datasets (e.g. a new OBI variant or a new lookback window). Running the script registers the output in the registry.
- **Data Registry Table** — lists all generated datasets with their name, type (OBI / Rise Ratio / Depth), number of rows, author, creation date, and publication status (`Published`, `Under Review`, `Draft`, or `Deprecated`). Clicking a row opens a detail modal showing summary statistics and a preview of the first 10 rows.

> Only datasets with `Status = Published` are visible to the Quant Researcher.

---

## Role 2 — QR: Quant Researcher

The Quant Researcher designs **alpha signals** — rule-based expressions that decide, for each moment in time, whether to place a BUY order or do nothing — and evaluates their profitability via backtesting.

### Panel — HFT Factor Expression Editor

A Python code editor (with syntax highlighting) where the QR writes the alpha signal function. The function reads order-book variables and returns either `1` (BUY) or `0` (NO TRADE).

The **OB Variable Library** panel alongside the editor lists all available variables, grouped by category (price levels, quantities, derived OB factors). Variables published by the DS are marked with a `DS` badge. Clicking a variable name inserts it directly into the code at the cursor position.

### Panel — Simulation Control

Before running a backtest, the QR configures:
- **Instrument** — the futures contract to simulate on
- **Lookback window** — how far back the signal looks (e.g. 10 s, 60 s, 1800 s)
- **Prediction window** — the horizon being predicted (e.g. 10 s)
- **OBI Weights** — how much emphasis to place on each depth level of the order book

Clicking **▶ RUN BACKTEST** runs the alpha expression across all historical ticks and computes the PnL.

### Panel — Cumulative PnL Chart

Shows the running profit-and-loss in **basis points (bps)** as the backtest progresses through time. The chart makes it easy to see whether the alpha generates consistent gains or has large losses at certain periods.

### Panel — Backtest Metrics

A grid of summary statistics computed after the backtest:

| Metric | What it measures |
|---|---|
| **Total PnL** | Net profit across all windows, in basis points |
| **Sharpe Ratio** | Risk-adjusted return — higher is better |
| **Win Rate** | Fraction of BUY signals that were correct |
| **Max Drawdown** | Largest peak-to-trough loss on the cumulative PnL curve |
| **Trade Count** | How many windows generated a BUY signal |

### PnL Calculation Rules

For every rolling window, the signal's outcome is scored as follows:

| Signal | True outcome | PnL |
|---|---|---|
| BUY (`1`) | Price went UP (`1`) | **`+spread`** bps — correct trade |
| NO TRADE (`0`) | Price went DOWN (`0`) | **`0`** — correctly avoided a loss |
| BUY (`1`) | Price went DOWN (`0`) | **`−spread`** bps — wrong trade |
| NO TRADE (`0`) | Price went UP (`1`) | **`0`** — missed opportunity, but no cost |

Where `spread = ask1 − bid1` in basis points, representing the transaction cost of entering the position.

---

## Role 3 — PM: Portfolio Manager

The Portfolio Manager has **read-only visibility** over everything produced by the QR team. The PM's job is to monitor strategy performance, ensure the portfolio is diversified, and set capital weights.

### Panel — HFT Strategy Leaderboard

A sortable table listing every strategy submitted by QR researchers. Each row shows:
- Strategy ID and author
- Which ML model and OBI weight scheme it uses
- Performance metrics: Total PnL, Sharpe, Win Rate, Max Drawdown
- A small PnL sparkline showing the shape of the equity curve
- Status: `Active`, `Paused`, or `Archived`

The PM can search by strategy ID or author, and sort by any column to find the top performers.

### Panel — Alpha Signal Correlation Heatmap

A matrix showing the **Pearson correlation** between each pair of strategies' submission sequences. High correlation (close to 1.0) means two strategies are making very similar trading decisions and do not add diversification value to the portfolio.

### Panel — Model Performance Comparison

A ranked summary table of the DS-trained classifiers, showing their mean accuracy, standard deviation, F1 score, and the rolling window where they performed best. This helps the PM understand which models the best-performing QR strategies are built on.

### Panel — Live Order Book Depth & Deployment Weights

- **Depth Chart** — a real-time bar chart of the current bid/ask quantity at each price level, giving the PM a feel for current market liquidity.
- **Strategy Deployment Weights** — a table showing how much of the portfolio capital is allocated to each active strategy, the current signal it is generating, and its estimated contribution to today's portfolio PnL.

---

## HFT Factors (OB Signals)

Three core order-book signals are used as features and building blocks for alpha expressions:

### Order Book Imbalance (OBI)
Measures the balance between buying and selling pressure at the top of the order book.
```
OBI = (WQ_bid − WQ_ask) / (WQ_bid + WQ_ask)
```
Positive OBI → more buying pressure → price may rise.  
Different weight schemes put more or less emphasis on the best price level vs. deeper levels.

### Rise Ratio
Measures short-term upward price momentum.
```
rise_ratio(t, n) = (ask1[t] − ask1[t−n]) / ask1[t−n] × 100
```
Computed for various lookback windows (e.g. 10 s, 30 s, 60 s).

### Depth Ratio
Measures the relative liquidity on the ask side vs. the bid side.
```
depth_ratio = WQ_ask / WQ_bid
```

---

## Project Structure

```
HFT/
├── backend/                # Go Backend Service (Gin HTTP framework, Pgx PG Client, Redis Producer)
├── frontend/               # Angular 18 SPA Frontend (served via Nginx or pnpm)
├── worker/                 # Asynchronous Python Worker (Redis consumer, scikit-learn training, backtester)
├── data/                   # Dataset Folder
│   └── VN30F2112.csv       # Raw Level-3 order-book data, VN30 Futures (55MB)
├── docker-compose.yml      # Service Orchestration Configuration
└── README.md               # This file
```

---

## How to Run

The easiest way to spin up the entire stabilized HFT environment (PostgreSQL, Redis, Go Backend, Python Worker, and Angular Frontend) is via Docker Compose:

```bash
# 1. Clean build and start all containers in the background
docker compose up --build -d

# 2. Check service status
docker compose ps
```

Once all services are healthy:
* **Frontend**: Open `http://localhost:3000` (Nginx-served) or `http://localhost:4200` (local dev)
* **Backend API**: Accessible at `http://localhost:8080`
* **Default Seed Users**:
  * **Quant Researcher**: `quant` / `password123`
  * **Data Scientist**: `admin` / `password123`
  * **Portfolio Manager**: `pm` / `password123`

To run backend tests locally:
```bash
cd backend && go test ./...
```

To run worker sandbox tests locally:
```bash
pytest worker/
```

## References

1. VN30 Futures dataset: `data/VN30F2112.csv`
2. ML-HFT reference framework (third-party, not included in this repo): [bradleyboyuyang/ML-HFT](https://github.com/bradleyboyuyang/ML-HFT)
3. Scikit-learn documentation: https://scikit-learn.org/stable/
