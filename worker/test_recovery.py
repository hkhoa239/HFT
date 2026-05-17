import json
import unittest
from unittest.mock import MagicMock, patch
import os

# Set dummy env vars for imports
os.environ["REDIS_URL"] = "redis://localhost:6379/0"
os.environ["DATABASE_URL"] = "postgresql://user:pass@localhost:5432/db"

from main import recover_pending_jobs, STREAM_NAME, GROUP_NAME, CONSUMER_NAME

class TestWorkerRecovery(unittest.TestCase):
    @patch("main.r")
    @patch("main.process_job")
    def test_recover_pending_jobs(self, mock_process, mock_redis):
        # Setup: 1 pending message in PEL
        message_id = "1620000000000-0"
        job_data = {"job_id": "test-job-123", "task_type": "backtest"}
        payload = {"payload": json.dumps(job_data)}
        
        # xreadgroup returns [(stream_name, [(message_id, payload)])]
        mock_redis.xreadgroup.return_value = [(STREAM_NAME, [(message_id, payload)])]
        
        # Execute
        recover_pending_jobs()
        
        # Verify: process_job called with correct data
        mock_process.assert_called_once_with(job_data)
        
        # Verify: message was ACKed
        mock_redis.xack.assert_called_once_with(STREAM_NAME, GROUP_NAME, message_id)

    @patch("main.r")
    @patch("main.process_job")
    def test_recover_no_pending_jobs(self, mock_process, mock_redis):
        # Setup: no pending messages
        mock_redis.xreadgroup.return_value = []
        
        # Execute
        recover_pending_jobs()
        
        # Verify: process_job not called
        mock_process.assert_not_called()
        mock_redis.xack.assert_not_called()

if __name__ == "__main__":
    unittest.main()
