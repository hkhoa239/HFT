import pytest
import pandas as pd
import numpy as np
from engine import validate_script, compile_signal, run_backtest

def test_validate_script_safe():
    safe_script = "def signal(row): return 1"
    validate_script(safe_script) # Should not raise

def test_validate_script_unsafe():
    unsafe_scripts = [
        "import os; os.system('ls')",
        "eval('1+1')",
        "open('/etc/passwd')",
        "subprocess.run(['ls'])"
    ]
    for script in unsafe_scripts:
        with pytest.raises(ValueError, match="Security violation"):
            validate_script(script)

def test_compile_signal_success():
    script = """
def signal(row):
    # Test access to math and np
    val = math.sqrt(row.get('bid1', 0))
    return 1 if val > 0 else 0
"""
    fn = compile_signal(script)
    assert callable(fn)
    assert fn({'bid1': 100}) == 1

def test_compile_signal_no_builtins_leak():
    # Attempt to use a builtin that SHOULD be restricted (like __import__)
    script = """
def signal(row):
    x = __import__('os')
    return 1
"""
    fn = compile_signal(script)
    with pytest.raises(NameError): # Restricted builtins should cause NameError at runtime
        fn({})

def test_run_backtest_deterministic():
    # Create deterministic data - NEED at least 5 rows for window
    data = {
        'time': [f'09:00:{i:02d}.000' for i in range(10)],
        'bid1': [100.0 + i for i in range(10)],
        'ask1': [100.5 + i for i in range(10)],
        'label': [1] * 10
    }
    df = pd.DataFrame(data)
    # Preprocess as engine expects
    df["time_sec"] = pd.Series([float(i) for i in range(10)])
    df["spread_bps"] = 10.0
    
    # Simple signal: always buy
    def signal_fn(row): return 1
    
    metrics, pnl_curve = run_backtest(df, signal_fn, lookback_sec=60, prediction_sec=1)
    
    assert "total_pnl" in metrics
    assert "win_rate" in metrics
    assert "sharpe_ratio" in metrics
    assert len(pnl_curve) > 0
    assert metrics["trade_count"] > 0

