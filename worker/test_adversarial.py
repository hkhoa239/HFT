import pytest
import numpy as np
import pandas as pd
from engine import compile_signal, run_backtest

def test_sandbox_escapes():
    # 1. Direct Import Attempt
    with pytest.raises(ValueError, match="Security violation"):
        compile_signal("import os\ndef signal(row): return 1")

    # 2. Attribute Traversal (__class__)
    with pytest.raises(ValueError, match="Security violation"):
        compile_signal("def signal(row): return row.__class__")

    # 3. MRO Traversal
    with pytest.raises(ValueError, match="Security violation"):
        compile_signal("def signal(row): return row.__mro__")

    # 4. Globals/Locals access
    with pytest.raises(ValueError, match="Security violation"):
        compile_signal("def signal(row): return globals()")

def test_recursion_bomb():
    script = """
def signal(row):
    def bomb(n):
        return bomb(n+1)
    return bomb(1)
"""
    signal_fn = compile_signal(script)
    df = pd.DataFrame({
        "time_sec": [1.0, 2.0, 3.0, 4.0, 5.0, 6.0],
        "bid1": [100, 101, 102, 103, 104, 105],
        "ask1": [101, 102, 103, 104, 105, 106],
        "label": [1, 0, 1, 0, 1, 0],
        "spread_bps": [0.5, 0.5, 0.5, 0.5, 0.5, 0.5]
    })
    
    # The engine catches exceptions in signal_fn and returns sig=0
    # We verify it completes without hanging or crashing the worker
    metrics, curve = run_backtest(df, signal_fn, timeout_sec=2)
    assert metrics["trade_count"] == 0
    assert len(curve) > 0

def test_memory_abuse():
    # This is harder to catch without OS-level limits, but we can verify it doesn't crash the worker instantly
    script = """
def signal(row):
    x = [i for i in range(100000)] # Smaller allocation for test speed
    return 1
"""
    signal_fn = compile_signal(script)
    df = pd.DataFrame({
        "time_sec": [1.0, 2.0, 3.0, 4.0, 5.0, 6.0],
        "bid1": [100, 101, 102, 103, 104, 105],
        "ask1": [101, 102, 103, 104, 105, 106],
        "label": [1, 0, 1, 0, 1, 0],
        "spread_bps": [0.5, 0.5, 0.5, 0.5, 0.5, 0.5]
    })
    
    # Verify it runs but doesn't cause a system crash (within limits of this test)
    metrics, curve = run_backtest(df, signal_fn, timeout_sec=5)
    assert metrics["trade_count"] >= 0

def test_indirect_eval():
    # Attempting to reconstruct eval
    script = """
def signal(row):
    e = getattr(__builtins__, 'ev' + 'al')
    return e('1+1')
"""
    # This should be caught by getattr being in DENY_LIST
    with pytest.raises(ValueError, match="Security violation"):
        compile_signal(script)
