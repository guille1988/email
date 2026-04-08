CREATE TABLE IF NOT EXISTS emails (
    id SERIAL PRIMARY KEY,
    `to` VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_emails_to (`to`),
    INDEX idx_emails_status (status),
    INDEX idx_emails_type (type)
);
