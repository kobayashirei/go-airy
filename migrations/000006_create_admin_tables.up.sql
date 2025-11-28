-- Create admin_logs table
CREATE TABLE IF NOT EXISTS `admin_logs` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `operator_id` BIGINT NOT NULL,
    `action` VARCHAR(50) NOT NULL,
    `entity_type` VARCHAR(20),
    `entity_id` BIGINT,
    `ip` VARCHAR(45),
    `details` JSON,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_operator_id` (`operator_id`),
    INDEX `idx_action` (`action`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`operator_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
