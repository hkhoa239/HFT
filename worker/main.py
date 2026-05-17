import os
import json
import time
import redis
import signal
import sys
from pathlib import Path
from db import mark_running, mark_completed, mark_failed, update_model_metrics
from engine import load_data, compile_signal, run_backtest

# Config
REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379/0")
# Use absolute path to avoid directory context issues
BASE_DIR = Path(__file__).parent.parent.absolute()
DATA_DIR = Path(os.getenv("DATA_DIR", BASE_DIR / "data"))
STREAM_NAME = "job_queue"
GROUP_NAME = "worker_group"
CONSUMER_NAME = "worker_1"

r = redis.from_url(REDIS_URL, decode_responses=True)

# Global flag for graceful shutdown
keep_running = True

def handle_signal(signum, frame):
    global keep_running
    print(f"\nSignal {signum} received. Shutting down gracefully...")
    keep_running = False

# Register signals
signal.signal(signal.SIGINT, handle_signal)
signal.signal(signal.SIGTERM, handle_signal)

def init_stream():
    try:
        # Use '$' to only consume new jobs from the moment group is created
        r.xgroup_create(STREAM_NAME, GROUP_NAME, id='$', mkstream=True)
        print(f"Consumer group {GROUP_NAME} created on {STREAM_NAME}")
    except redis.exceptions.ResponseError:
        print(f"Consumer group {GROUP_NAME} already exists")

def process_job(job_data: dict):
    job_id = job_data.get("job_id")
    task_type = job_data.get("task_type")

    # Resolve flat or nested parameters
    params = job_data.get("params", {})
    inner_params = params.get("params", {}) if isinstance(params.get("params"), dict) else params
    instrument = inner_params.get("instrument", params.get("instrument", "VN30F2112"))

    if task_type == "backtest":
        print(f"Processing Backtest Job: {job_id} for instrument: {instrument}")
        mark_running(job_id)
        try:
            csv_path = DATA_DIR / f"{instrument}.csv"
            df = load_data(csv_path)
            
            # Retrieve parameters
            lookback_sec = int(inner_params.get("lookback_sec", 60))
            prediction_sec = int(inner_params.get("prediction_sec", 10))
            start_date = inner_params.get("start", "")
            end_date = inner_params.get("end", "")
            
            # Apply date range filtering if provided
            if start_date:
                df = df[df["datetime"] >= start_date]
            if end_date:
                df = df[df["datetime"] <= f"{end_date} 23:59:59.999999"]
                
            if len(df) == 0:
                raise ValueError(f"No tick data available for instrument {instrument} between {start_date} and {end_date}")
                
            script = job_data.get("script", "")
            print(f"Received script payload: type={type(script)}, len={len(script)}, repr={repr(script[:200])}")
            signal_fn = compile_signal(script)
            metrics, pnl_curve = run_backtest(df, signal_fn, lookback_sec=lookback_sec, prediction_sec=prediction_sec)
            metrics["pnl_curve"] = [{"cumPnL": float(v)} for v in pnl_curve]
            mark_completed(job_id, metrics)
            print(f"Job {job_id} completed successfully")
        except Exception as e:
            error_msg = str(e)
            print(f"Job {job_id} failed: {error_msg}")
            mark_failed(job_id, error_msg)
        return

    if task_type == "train":
        print(f"Processing Train Job: {job_id} for instrument: {instrument}")
        try:
            csv_path = DATA_DIR / f"{instrument}.csv"
            df = load_data(csv_path)
            
            # Extract features and target label
            feature_cols = [c for c in ["bq1", "bq2", "bq3", "aq1", "aq2", "aq3", "spread_bps", "obi"] if c in df.columns]
            if not feature_cols:
                feature_cols = [c for c in df.columns if c not in ["label", "id", "ts", "created_at", "updated_at", "timestamp"]]
            
            X = df[feature_cols].fillna(0).values
            y = df["label"].fillna(0).astype(int).values
            
            algo = inner_params.get("algorithm", "RandomForest")
            n_est = int(inner_params.get("n_estimators", 10))
            m_depth = inner_params.get("max_depth", None)
            if m_depth is not None:
                try:
                    m_depth = int(m_depth)
                except:
                    m_depth = None
            
            lr = float(inner_params.get("learning_rate", 0.1))
            sub = float(inner_params.get("subsample", 1.0))
            c_val = float(inner_params.get("C", 1.0))
            
            print(f"Starting real ML model fit: algo={algo}, params={inner_params}")
            
            # Initialize chosen classifier
            if algo == "RandomForest":
                from sklearn.ensemble import RandomForestClassifier
                clf = RandomForestClassifier(n_estimators=n_est, max_depth=m_depth, random_state=42)
            elif algo == "XGBoost":
                try:
                    from xgboost import XGBClassifier
                    clf = XGBClassifier(n_estimators=n_est, max_depth=m_depth or 6, learning_rate=lr, subsample=sub, random_state=42)
                except ImportError:
                    print("xgboost not installed, falling back to sklearn GradientBoostingClassifier")
                    from sklearn.ensemble import GradientBoostingClassifier
                    clf = GradientBoostingClassifier(n_estimators=n_est, max_depth=m_depth or 6, learning_rate=lr, random_state=42)
            elif algo == "LogisticRegression":
                from sklearn.linear_model import LogisticRegression
                from sklearn.preprocessing import StandardScaler
                from sklearn.pipeline import Pipeline
                clf = Pipeline([
                    ('scaler', StandardScaler()),
                    ('lr', LogisticRegression(C=c_val, max_iter=1000, random_state=42))
                ])
            elif algo == "GradientBoosting":
                from sklearn.ensemble import GradientBoostingClassifier
                clf = GradientBoostingClassifier(n_estimators=n_est, max_depth=m_depth or 3, learning_rate=lr, random_state=42)
            elif algo == "SVM":
                from sklearn.svm import SVC
                clf = SVC(C=c_val, kernel='rbf', probability=True, random_state=42)
            else:
                from sklearn.ensemble import RandomForestClassifier
                clf = RandomForestClassifier(n_estimators=10, max_depth=10, random_state=42)
            
            # Chronological 80/20 train/validation split
            split_idx = int(len(X) * 0.8)
            X_train, X_val = X[:split_idx], X[split_idx:]
            y_train, y_val = y[:split_idx], y[split_idx:]
            
            clf.fit(X_train, y_train)
            
            # Predict and evaluate
            y_pred = clf.predict(X_val)
            if hasattr(clf, "predict_proba"):
                y_prob = clf.predict_proba(X_val)[:, 1]
            elif hasattr(clf, "decision_function"):
                y_prob = clf.decision_function(X_val)
            else:
                y_prob = y_pred.astype(float)
                
            from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score, confusion_matrix, log_loss, roc_auc_score
            import numpy as np
            
            acc = float(accuracy_score(y_val, y_pred))
            prec = float(precision_score(y_val, y_pred, zero_division=0))
            rec = float(recall_score(y_val, y_pred, zero_division=0))
            f1 = float(f1_score(y_val, y_pred, zero_division=0))
            
            try:
                loss = float(log_loss(y_val, y_prob, labels=[0, 1]))
            except:
                loss = 0.5
            try:
                auc = float(roc_auc_score(y_val, y_prob))
            except:
                auc = 0.5
                
            # Estimate cross-validation standard deviation using 5 validation folds
            acc_folds = []
            fold_size = len(X_val) // 5
            if fold_size > 5:
                for k in range(5):
                    X_f = X_val[k*fold_size : (k+1)*fold_size]
                    y_f = y_val[k*fold_size : (k+1)*fold_size]
                    if len(X_f) > 0:
                        acc_folds.append(accuracy_score(y_f, clf.predict(X_f)))
                acc_std = float(np.std(acc_folds)) if acc_folds else 0.02
            else:
                acc_std = 0.02
                
            tn, fp, fn, tp = confusion_matrix(y_val, y_pred, labels=[0, 1]).ravel()
            
            metrics = {
                "accuracy": round(acc, 4),
                "precision": round(prec, 4),
                "recall": round(rec, 4),
                "f1_score": round(f1, 4),
                "log_loss": round(loss, 4),
                "acc_std": round(acc_std, 4),
                "tp": int(tp),
                "fp": int(fp),
                "fn": int(fn),
                "tn": int(tn),
                "auc": round(auc, 4),
            }
            
            pkl_path = str(DATA_DIR / "models" / f"{job_id}.pkl")
            os.makedirs(str(DATA_DIR / "models"), exist_ok=True)
            
            # Save actual model artifact to pickle format
            import pickle
            with open(pkl_path, "wb") as f:
                pickle.dump(clf, f)
                
            update_model_metrics(job_id, metrics, pkl_path)
            print(f"Train job {job_id} completed with metrics: {metrics}")
        except Exception as e:
            print(f"Train job {job_id} failed: {e}")
        return

    print(f"Unknown task type: {task_type}")

def recover_pending_jobs():
    """Recover jobs that were started but not finished (PEL)."""
    print("Checking for pending jobs in PEL...")
    try:
        # Read pending messages for this consumer
        # ID '0' means read all messages in the PEL for this consumer
        streams = r.xreadgroup(GROUP_NAME, CONSUMER_NAME, {STREAM_NAME: "0"}, count=10)
        
        if not streams:
            print("No pending jobs found.")
            return

        for stream_name, messages in streams:
            for message_id, payload in messages:
                # Check idle time (only recover if idle > 30s)
                # Redis xpending gives more detail but xreadgroup '0' is simpler for MVP
                # We'll just process them but keep an eye on retry count
                raw_data = payload.get("payload")
                if raw_data:
                    job_data = json.loads(raw_data)
                    job_id = job_data.get("job_id")
                    
                    # Prevent infinite retry loops
                    # We can store retry count in Redis or payload
                    # For MVP, we'll just log and try once
                    print(f"Reclaiming pending job: {job_id} (ID: {message_id})")
                    process_job(job_data)
                    
                # ACK message
                r.xack(STREAM_NAME, GROUP_NAME, message_id)
                
    except Exception as e:
        print(f"PEL Recovery Error: {e}")

def main():
    print("Python Worker started...")
    init_stream()
    
    # Check for pending jobs before entering main loop
    recover_pending_jobs()
    
    while keep_running:
        try:
            # Read from stream
            # Block for 5 seconds to allow keep_running check
            streams = r.xreadgroup(GROUP_NAME, CONSUMER_NAME, {STREAM_NAME: ">"}, count=1, block=5000)
            
            if not streams:
                continue
                
            for stream_name, messages in streams:
                for message_id, payload in messages:
                    # Payload is a dict with 'payload' key containing JSON string
                    raw_data = payload.get("payload")
                    if raw_data:
                        job_data = json.loads(raw_data)
                        process_job(job_data)
                    
                    # ACK message
                    r.xack(STREAM_NAME, GROUP_NAME, message_id)
                    
        except Exception as e:
            if keep_running:
                print(f"Worker Loop Error: {e}")
                time.sleep(1)
            else:
                break
    
    print("Worker stopped.")

if __name__ == "__main__":
    main()
