-- Add composite indexes for common query patterns

-- Posts: Composite index for feed queries (status + created_at for timeline)
CREATE INDEX IF NOT EXISTS `idx_posts_status_created` ON `posts` (`status`, `created_at` DESC);

-- Posts: Composite index for circle feed queries
CREATE INDEX IF NOT EXISTS `idx_posts_circle_status_created` ON `posts` (`circle_id`, `status`, `created_at` DESC);

-- Posts: Composite index for author posts queries
CREATE INDEX IF NOT EXISTS `idx_posts_author_status_created` ON `posts` (`author_id`, `status`, `created_at` DESC);

-- Posts: Composite index for hotness sorting
CREATE INDEX IF NOT EXISTS `idx_posts_status_hotness` ON `posts` (`status`, `hotness_score` DESC);

-- Posts: Index for published_at for scheduled posts
CREATE INDEX IF NOT EXISTS `idx_posts_published_at` ON `posts` (`published_at`);

-- Comments: Composite index for post comments with status
CREATE INDEX IF NOT EXISTS `idx_comments_post_status_created` ON `comments` (`post_id`, `status`, `created_at` DESC);

-- Comments: Composite index for author comments
CREATE INDEX IF NOT EXISTS `idx_comments_author_status_created` ON `comments` (`author_id`, `status`, `created_at` DESC);

-- Votes: Index for entity lookups (for counting votes)
CREATE INDEX IF NOT EXISTS `idx_votes_entity` ON `votes` (`entity_type`, `entity_id`);

-- Users: Index for last login (for active users queries)
CREATE INDEX IF NOT EXISTS `idx_users_last_login` ON `users` (`last_login_at` DESC);

-- Users: Composite index for status and created_at
CREATE INDEX IF NOT EXISTS `idx_users_status_created` ON `users` (`status`, `created_at` DESC);

-- Circles: Index for member count (for popular circles)
CREATE INDEX IF NOT EXISTS `idx_circles_member_count` ON `circles` (`member_count` DESC);

-- Circles: Composite index for status and member count
CREATE INDEX IF NOT EXISTS `idx_circles_status_members` ON `circles` (`status`, `member_count` DESC);

-- Notifications: Composite index for user notifications
CREATE INDEX IF NOT EXISTS `idx_notifications_receiver_read_created` ON `notifications` (`receiver_id`, `is_read`, `created_at` DESC);

-- Messages: Composite index for conversation messages
CREATE INDEX IF NOT EXISTS `idx_messages_conversation_created` ON `messages` (`conversation_id`, `created_at` DESC);

-- Conversations: Index for user conversations
CREATE INDEX IF NOT EXISTS `idx_conversations_user1_last_message` ON `conversations` (`user1_id`, `last_message_at` DESC);
CREATE INDEX IF NOT EXISTS `idx_conversations_user2_last_message` ON `conversations` (`user2_id`, `last_message_at` DESC);
