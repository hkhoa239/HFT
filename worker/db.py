import os
import json
from typing import Any, Optional
import psycopg
from datetime import datetime

DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://hft:hft@localhost:5432/hft")

def get_connection():
    return psycopg.connect(DATABASE_URL, autocommit=True)

def update_job_status(
    job_id: str, 
    status: str, 
    metrics: Optional[dict] = None, 
    error_log: Optional[str] = None
):
    """
    Update backtest_runs table status and results.
    Status can be: pending, running, completed, failed
    """
    with get_connection() as conn:
        with conn.cursor() as cur:
            finished_at = None
            if status in ['completed', 'failed']:
                finished_at = datetime.now()
            
            # Map 'completed' to 'completed' for consistency with models.JobStatus
            # Models use: pending, running, completed, failed
            
            cur.execute(
                """
                UPDATE backtest_runs
                SET status = %s,
                    metrics = %s,
                    error_log = %s,
                    finished_at = COALESCE(%s, finished_at)
                WHERE id = %s
                """,
                (
                    status,
                    json.dumps(metrics) if metrics else None,
                    error_log,
                    finished_at,
                    job_id
                )
            )

def mark_running(job_id: str):
    update_job_status(job_id, 'running')

def mark_completed(job_id: str, metrics: dict):
    update_job_status(job_id, 'completed', metrics=metrics)

def mark_failed(job_id: str, error_log: str):
    update_job_status(job_id, 'failed', error_log=error_log)

def update_model_metrics(job_id: str, metrics: dict, pkl_path: str = ""):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE models
                SET training_metrics = %s,
                    pkl_path = CASE WHEN %s <> '' THEN %s ELSE pkl_path END,
                    updated_at = CURRENT_TIMESTAMP
                WHERE id = %s
                """,
                (json.dumps(metrics), pkl_path, pkl_path, job_id)
            )
