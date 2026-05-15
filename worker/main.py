import os
import json
import time
import redis
from pathlib import Path
from db import mark_running, mark_completed, mark_failed
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
    
    if task_type != "backtest":
        print(f"Unknown task type: {task_type}")
        return

    print(f"Processing Backtest Job: {job_id}")
    mark_running(job_id)
    
    try:
        # 1. Prepare data
        instrument = job_data.get("params", {}).get("instrument", "VN30F2112")
        csv_path = DATA_DIR / f"{instrument}.csv"
        df = load_data(csv_path)
        
        # 2. Compile script
        script = job_data.get("script", "")
        signal_fn = compile_signal(script)
        
        # 3. Run backtest
        metrics, pnl_curve = run_backtest(df, signal_fn)
        
        # 4. Success
        metrics["pnl_curve"] = pnl_curve # Attach curve to metrics for DB
        mark_completed(job_id, metrics)
        print(f"Job {job_id} completed successfully")
        
    except Exception as e:
        error_msg = str(e)
        print(f"Job {job_id} failed: {error_msg}")
        mark_failed(job_id, error_msg)

def main():
    print("Python Worker started...")
    init_stream()
    
    while True:
        try:
            # Read from stream
            # Block for 5 seconds
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
            print(f"Worker Loop Error: {e}")
            time.sleep(1)

if __name__ == "__main__":
    main()
