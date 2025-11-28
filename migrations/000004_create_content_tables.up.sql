-- Create posts table
CREATE TABLE IF NOT EXISTS `posts` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `title` VARCHAR(255) NOT NULL,
    `content_markdown` TEXT NOT NULL,
    `content_html` TEXT NOT NULL,
    `summary` VARCHAR(500),
    `cover_image` VARCHAR(255),
    `author_id` BIGINT NOT NULL,
    `circle_id` BIGINT,
    `status` VARCHAR(20) DEFAULT 'draft',
    `category` VARCHAR(50),
    `tags` JSON,
    `scheduled_at` DATETIME,
    `is_pinned` BOOLEAN DEFAULT FALSE,
    `is_featured` BOOLEAN DEFAULT FALSE,
    `allow_comment` BOOLEAN DEFAULT TRUE,
    `is_anonymous` BOOLEAN DEFAULT FALSE,
    `view_count` INT DEFAULT 0,
    `hotness_score` DOUBLE DEFAULT 0,
    `published_at` DATETIME,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_author_id` (`author_id`),
    INDEX `idx_circle_id` (`circle_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_hotness_score` (`hotness_score`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`author_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`circle_id`) REFERENCES `circles`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create comments table
CREATE TABLE IF NOT EXISTS `comments` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `content` TEXT NOT NULL,
    `author_id` BIGINT NOT NULL,
    `post_id` BIGINT NOT NULL,
    `parent_id` BIGINT,
    `root_id` BIGINT NOT NULL,
    `level` INT DEFAULT 0,
    `path` VARCHAR(255),
    `status` VARCHAR(20) DEFAULT 'published',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_author_id` (`author_id`),
    INDEX `idx_post_id` (`post_id`),
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_root_id` (`root_id`),
    INDEX `idx_path` (`path`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`author_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`post_id`) REFERENCES `posts`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`parent_id`) REFERENCES `comments`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create votes table
CREATE TABLE IF NOT EXISTS `votes` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `entity_type` VARCHAR(20) NOT NULL,
    `entity_id` BIGINT NOT NULL,
    `vote_type` VARCHAR(10) NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_user_entity` (`user_id`, `entity_type`, `entity_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create favorites table
CREATE TABLE IF NOT EXISTS `favorites` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL,
    `post_id` BIGINT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_user_post` (`user_id`, `post_id`),
    INDEX `idx_post_id` (`post_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`post_id`) REFERENCES `posts`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create entity_counts table
CREATE TABLE IF NOT EXISTS `entity_counts` (
    `entity_type` VARCHAR(20) NOT NULL,
    `entity_id` BIGINT NOT NULL,
    `upvote_count` INT DEFAULT 0,
    `downvote_count` INT DEFAULT 0,
    `comment_count` INT DEFAULT 0,
    `favorite_count` INT DEFAULT 0,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`entity_type`, `entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
