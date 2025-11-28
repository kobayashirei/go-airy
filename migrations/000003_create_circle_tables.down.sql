-- Remove foreign key from user_roles
ALTER TABLE `user_roles` DROP FOREIGN KEY `fk_user_roles_circle`;

-- Drop circle_members table
DROP TABLE IF EXISTS `circle_members`;

-- Drop circles table
DROP TABLE IF EXISTS `circles`;
