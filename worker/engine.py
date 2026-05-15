import json
import math
import time
from pathlib import Path
from typing import Any, Callable, Tuple
import numpy as np
import pandas as pd

def load_data(csv_path: Path) -> pd.DataFrame:
    """Load and preprocess VN30F data."""
    if not csv_path.exists():
        raise FileNotFoundError(f"Data file not found: {csv_path}")
    
    df = pd.read_csv(csv_path)
    df.columns = [c.strip().lower() for c in df.columns]

    # Time processing
    if "time" in df.columns:
        df["time_sec"] = pd.to_datetime(df["time"], format="%H:%M:%S.%f", errors="coerce")
        df["time_sec"] = (
            df["time_sec"].dt.hour * 3600
            + df["time_sec"].dt.minute * 60
            + df["time_sec"].dt.second
            + df["time_sec"].dt.microsecond / 1e6
        )
    else:
        df["time_sec"] = pd.Series(range(len(df)), dtype=float)
    
    df["time_sec"] = df["time_sec"].ffill().fillna(0)

    # Derived features (Minimal set for MVP)
    bid1 = pd.to_numeric(df.get("bid1", 0), errors="coerce").fillna(0)
    ask1 = pd.to_numeric(df.get("ask1", 0), errors="coerce").fillna(0)
    df["spread_bps"] = np.where(bid1 > 0, (ask1 - bid1) / bid1 * 10000, 0.5)
    
    # Simple label: price movement in next 10 ticks
    df["label"] = (ask1.shift(-10) > ask1).astype(int).fillna(0)
    
    return df.dropna(subset=["time_sec"]).reset_index(drop=True)

def compile_signal(script: str) -> Callable:
    """Safely compile the signal function."""
    ns = {
        "__builtins__": {
            "abs": abs, "min": min, "max": max, "round": round,
            "int": int, "float": float, "bool": bool, "str": str,
            "len": len, "range": range, "enumerate": enumerate,
            "zip": zip, "sum": sum, "sorted": sorted,
            "True": True, "False": False, "None": None,
        },
        "math": math,
        "np": np,
    }
    exec(compile(script, "<signal>", "exec"), ns)
    fn = ns.get("signal")
    if not callable(fn):
        raise ValueError("Script must define `signal(row) -> int`")
    return fn

def run_backtest(
    df: pd.DataFrame,
    signal_fn: Callable,
    lookback_sec: int = 60,
    prediction_sec: int = 10,
    timeout_sec: int = 300
) -> Tuple[dict, list]:
    """Execute rolling window backtest."""
    start_time = time.perf_counter()
    times = df["time_sec"].values
    cumulative_pnl = 0.0
    pnl_curve = []
    pnl_deltas = []
    trade_count = 0
    win_count = 0
    
    i = 0
    while i < len(df):
        if time.perf_counter() - start_time > timeout_sec:
            raise TimeoutError("Backtest timeout")

        t_end = times[i]
        t_start = t_end - lookback_sec
        
        # Simple windowing (using searchsorted for performance)
        idx_start = np.searchsorted(times, t_start)
        window = df.iloc[idx_start:i+1]
        
        if len(window) < 5:
            i += 1
            continue

        last_row = window.iloc[-1].to_dict()
        
        try:
            sig = int(signal_fn(last_row))
            sig = 1 if sig >= 1 else 0
        except:
            sig = 0

        # Outcome
        idx_future_end = np.searchsorted(times, t_end + prediction_sec, side='right')
        future = df.iloc[i+1:idx_future_end]
        
        if len(future) == 0: break
        
        label = int(future["label"].iloc[0])
        spread_bps = float(last_row.get("spread_bps", 0.5))

        pnl_delta = 0.0
        if sig == 1:
            trade_count += 1
            pnl_delta = spread_bps if label == 1 else -spread_bps
            if label == 1: win_count += 1
        
        cumulative_pnl += pnl_delta
        pnl_curve.append(round(float(cumulative_pnl), 4))
        pnl_deltas.append(float(pnl_delta))
        
        # Advance
        i = idx_future_end

    # Compute Final Metrics
    metrics = {
        "total_pnl": round(float(cumulative_pnl), 4),
        "win_rate": round(float(win_count / trade_count), 4) if trade_count > 0 else 0.0,
        "trade_count": int(trade_count),
        "sharpe_ratio": 0.0
    }
    
    if pnl_deltas:
        arr = np.array(pnl_deltas)
        std = arr.std()
        if std > 0:
            val = (arr.mean() / std * math.sqrt(252*8*60))
            # Clean NaN/Inf
            if not np.isnan(val) and not np.isinf(val):
                metrics["sharpe_ratio"] = round(float(val), 4)

    return metrics, pnl_curve
