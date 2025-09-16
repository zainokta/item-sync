ALTER TABLE items CHANGE COLUMN synced_at last_synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE items 
ADD COLUMN content_hash VARCHAR(64) NOT NULL DEFAULT '',
ADD COLUMN sync_attempts INT NOT NULL DEFAULT 0,
ADD COLUMN last_sync_error TEXT;

CREATE INDEX idx_last_synced_at ON items(last_synced_at);
CREATE INDEX idx_content_hash ON items(content_hash);
CREATE INDEX idx_sync_attempts ON items(sync_attempts);

CREATE TABLE IF NOT EXISTS sync_jobs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    job_name VARCHAR(100) NOT NULL,
    api_source VARCHAR(100) NOT NULL,
    status ENUM('running', 'completed', 'failed') NOT NULL DEFAULT 'running',
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    items_processed INT NOT NULL DEFAULT 0,
    items_succeeded INT NOT NULL DEFAULT 0,
    items_failed INT NOT NULL DEFAULT 0,
    error_message TEXT,
    execution_time_ms INT NOT NULL DEFAULT 0,
    
    INDEX idx_job_name (job_name),
    INDEX idx_api_source (api_source),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;