-- Create circles table
CREATE TABLE IF NOT EXISTS `circles` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL UNIQUE,
    `description` TEXT,
    `avatar` VARCHAR(255),
    `background` VARCHAR(255),
    `creator_id` BIGINT NOT NULL,
    `status` VARCHAR(20) DEFAULT 'public',
    `join_rule` VARCHAR(20) DEFAULT 'free',
    `member_count` INT DEFAULT 0,
    `post_count` INT DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_creator_id` (`creator_id`),
    FOREIGN KEY (`creator_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create circle_members table
CREATE TABLE IF NOT EXISTS `circle_members` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `circle_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `role` VARCHAR(20) DEFAULT 'member',
    `joined_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_circle_user` (`circle_id`, `user_id`),
    INDEX `idx_circle_id` (`circle_id`),
    INDEX `idx_user_id` (`user_id`),
    FOREIGN KEY (`circle_id`) REFERENCES `circles`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key to user_roles for circle_id
ALTER TABLE `user_roles` ADD CONSTRAINT `fk_user_roles_circle` 
    FOREIGN KEY (`circle_id`) REFERENCES `circles`(`id`) ON DELETE CASCADE;
