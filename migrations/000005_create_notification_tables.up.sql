-- Create notifications table
CREATE TABLE IF NOT EXISTS `notifications` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `receiver_id` BIGINT NOT NULL,
    `trigger_user_id` BIGINT,
    `type` VARCHAR(20) NOT NULL,
    `entity_type` VARCHAR(20),
    `entity_id` BIGINT,
    `content` VARCHAR(500),
    `is_read` BOOLEAN DEFAULT FALSE,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_receiver_id` (`receiver_id`),
    INDEX `idx_is_read` (`is_read`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`receiver_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`trigger_user_id`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create conversations table
CREATE TABLE IF NOT EXISTS `conversations` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `user1_id` BIGINT NOT NULL,
    `user2_id` BIGINT NOT NULL,
    `last_message_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_users` (`user1_id`, `user2_id`),
    INDEX `idx_last_message_at` (`last_message_at`),
    FOREIGN KEY (`user1_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`user2_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create messages table
CREATE TABLE IF NOT EXISTS `messages` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `conversation_id` BIGINT NOT NULL,
    `sender_id` BIGINT NOT NULL,
    `content_type` VARCHAR(20) DEFAULT 'text',
    `content` TEXT NOT NULL,
    `is_read` BOOLEAN DEFAULT FALSE,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_conversation_id` (`conversation_id`),
    INDEX `idx_sender_id` (`sender_id`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`conversation_id`) REFERENCES `conversations`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`sender_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
