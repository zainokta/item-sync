-- Create items table for storing data from various APIs
CREATE TABLE IF NOT EXISTS items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    external_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    api_source VARCHAR(100) NOT NULL DEFAULT 'unknown',
    extend_info JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_external_id (external_id),
    INDEX idx_created_at (created_at),
    INDEX idx_synced_at (synced_at),
    INDEX idx_api_source (api_source),
    
    UNIQUE KEY uk_external_api_source (external_id, api_source)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;