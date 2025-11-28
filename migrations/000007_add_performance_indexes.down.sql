-- Drop performance indexes

DROP INDEX IF EXISTS `idx_posts_status_created` ON `posts`;
DROP INDEX IF EXISTS `idx_posts_circle_status_created` ON `posts`;
DROP INDEX IF EXISTS `idx_posts_author_status_created` ON `posts`;
DROP INDEX IF EXISTS `idx_posts_status_hotness` ON `posts`;
DROP INDEX IF EXISTS `idx_posts_published_at` ON `posts`;

DROP INDEX IF EXISTS `idx_comments_post_status_created` ON `comments`;
DROP INDEX IF EXISTS `idx_comments_author_status_created` ON `comments`;

DROP INDEX IF EXISTS `idx_votes_entity` ON `votes`;

DROP INDEX IF EXISTS `idx_users_last_login` ON `users`;
DROP INDEX IF EXISTS `idx_users_status_created` ON `users`;

DROP INDEX IF EXISTS `idx_circles_member_count` ON `circles`;
DROP INDEX IF EXISTS `idx_circles_status_members` ON `circles`;

DROP INDEX IF EXISTS `idx_notifications_receiver_read_created` ON `notifications`;

DROP INDEX IF EXISTS `idx_messages_conversation_created` ON `messages`;

DROP INDEX IF EXISTS `idx_conversations_user1_last_message` ON `conversations`;
DROP INDEX IF EXISTS `idx_conversations_user2_last_message` ON `conversations`;
