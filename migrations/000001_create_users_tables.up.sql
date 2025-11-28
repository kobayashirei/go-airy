-- Create users table
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `username` VARCHAR(50) NOT NULL UNIQUE,
    `email` VARCHAR(100) UNIQUE,
    `phone` VARCHAR(20) UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `avatar` VARCHAR(255),
    `gender` VARCHAR(10),
    `birthday` DATETIME,
    `bio` VARCHAR(500),
    `status` VARCHAR(20) DEFAULT 'inactive',
    `last_login_at` DATETIME,
    `last_login_ip` VARCHAR(45),
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create user_profiles table
CREATE TABLE IF NOT EXISTS `user_profiles` (
    `user_id` BIGINT PRIMARY KEY,
    `points` INT DEFAULT 0,
    `level` INT DEFAULT 1,
    `follower_count` INT DEFAULT 0,
    `following_count` INT DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create user_stats table
CREATE TABLE IF NOT EXISTS `user_stats` (
    `user_id` BIGINT PRIMARY KEY,
    `post_count` INT DEFAULT 0,
    `comment_count` INT DEFAULT 0,
    `vote_received_count` INT DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
