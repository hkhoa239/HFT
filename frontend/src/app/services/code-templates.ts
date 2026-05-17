export const QR_CODE = `# ═══════════════════════════════════════════════════════════
# HFT Alpha Expression — VN30F2112 (VN30 Futures Dec 2021)
# Output: submission = 0 (NO TRADE) | 1 (BUY signal)
# ═══════════════════════════════════════════════════════════

def alpha_signal(t, ask1, bq1, bq2, bq3, aq1, aq2, aq3):
    # BUY when buy-side pressure dominates
    obi = (bq1 - aq1) / (bq1 + aq1 + 1e-8)
    if obi > 0.20:
        return 1   # BUY
    return 0       # NO TRADE
`;

export const DS_CODE = `# HFT Rolling-Window Classifier
from sklearn.ensemble import RandomForestClassifier
import numpy as np

# Load pre-processed features
data = pd.read_csv('./data/VN30F2112.csv')

model = RandomForestClassifier(n_estimators=10, max_depth=10)
# ... training logic ...
`;

export const DS_ANALY_CODE = `# HFT Data Analysis
import pandas as pd

def analyze_ob(df):
    # Compute OBI at multiple levels
    df['obi_l1'] = (df['bq1'] - df['aq1']) / (df['bq1'] + df['aq1'])
    return df.describe()
`;

export const DS_GEN_CODE = `# HFT Factor Generator
def generate_obi(bq1, aq1):
    return (bq1 - aq1) / (bq1 + aq1 + 1e-8)
`;

export const VARIABLES = {
  'PRICE LEVELS': [
    { name: 'bid_price_1', desc: 'Best bid price (level 1)' },
    { name: 'ask_price_1', desc: 'Best ask price (level 1)' },
    { name: 'mid_price', desc: '(bid1+ask1)/2' },
  ],
  'QUANTITIES': [
    { name: 'bid_qty_1', desc: 'Bid quantity level 1' },
    { name: 'ask_qty_1', desc: 'Ask quantity level 1' },
  ],
  'OB FACTORS': [
    { name: 'spread', desc: 'ask1 - bid1 (in ticks)' },
    { name: 'OBI', desc: '(WQ_b - WQ_a)/(WQ_b + WQ_a)', ds: true },
  ],
};
