# Handler Feature Summary / 接口处理器功能汇总

This document summarizes the features implemented in the `internal/handler` package. It serves as a reference for understanding the available API capabilities at the code level.

本文档汇总了 `internal/handler` 包中实现的功能。它作为从代码层面了解可用 API 能力的参考。

## 1. AuthHandler (Authentication / 认证)
Handles user registration, login, and token management.
处理用户注册、登录和令牌管理。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `Register` | Registers a new user account. | 注册新用户账户。 |
| `Login` | Authenticates a user via password. | 通过密码认证用户。 |
| `LoginWithCode` | Authenticates a user via verification code. | 通过验证码认证用户。 |
| `Activate` | Activates a user account via token. | 通过令牌激活用户账户。 |
| `RefreshToken` | Refreshes the JWT access token. | 刷新 JWT 访问令牌。 |

## 2. UserProfileHandler (User Profile / 用户资料)
Manages user profiles and retrieval of user-related content.
管理用户资料及用户相关内容的检索。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `GetProfile` | Retrieves public profile information of a user. | 获取用户的公开资料信息。 |
| `UpdateProfile` | Updates the authenticated user's profile. | 更新已认证用户的资料。 |
| `GetUserPosts` | Lists posts created by a specific user. | 列出特定用户创建的帖子。 |

## 3. PostHandler (Posts / 帖子)
Handles the lifecycle of community posts.
处理社区帖子的生命周期。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `CreatePost` | Creates a new post. | 创建新帖子。 |
| `GetPost` | Retrieves a single post by ID. | 通过 ID 获取单个帖子。 |
| `UpdatePost` | Updates an existing post (author only). | 更新现有帖子（仅限作者）。 |
| `DeletePost` | Deletes a post (author only). | 删除帖子（仅限作者）。 |
| `ListPosts` | Lists posts with pagination and filtering. | 分页和筛选列出帖子。 |

## 4. CommentHandler (Comments / 评论)
Manages comments on posts, supporting a nested structure.
管理帖子下的评论，支持嵌套结构。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `CreateComment` | Adds a comment to a post. | 向帖子添加评论。 |
| `GetCommentTree` | Retrieves comments for a post as a hierarchical tree. | 获取帖子的层级树状评论。 |
| `DeleteComment` | Deletes a comment. | 删除评论。 |

## 5. VoteHandler (Votes / 投票)
Handles upvotes and downvotes on entities (posts, comments).
处理对实体（帖子、评论）的点赞和踩。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `Vote` | Casts a vote (up/down) on an entity. | 对实体进行投票（顶/踩）。 |
| `CancelVote` | Removes a previously cast vote. | 取消之前的投票。 |
| `GetVote` | Retrieves the user's current vote on an entity. | 获取用户当前对实体的投票状态。 |

## 6. CircleHandler (Circles / 圈子)
Manages community circles (groups) and membership.
管理社区圈子（小组）及其成员资格。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `CreateCircle` | Creates a new circle. | 创建新圈子。 |
| `GetCircle` | Retrieves circle details. | 获取圈子详情。 |
| `JoinCircle` | Adds the user to a circle. | 用户加入圈子。 |
| `GetCircleMembers` | Lists members of a circle. | 列出圈子成员。 |
| `ApproveMember` | Approves a pending member request. | 批准待处理的成员申请。 |
| `AssignModerator` | Promotes a member to moderator. | 提升成员为管理员。 |

## 7. FeedHandler (News Feed / 动态流)
Generates content feeds for users.
为用户生成内容动态流。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `GetFeed` | Retrieves the personalized activity feed for the user. | 获取用户的个性化动态流。 |
| `GetCircleFeed` | Retrieves the feed for a specific circle. | 获取特定圈子的动态流。 |

## 8. SearchHandler (Search / 搜索)
Provides search capabilities across the platform.
提供全平台的搜索能力。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `SearchPosts` | Searches for posts by keyword, tag, etc. | 按关键字、标签等搜索帖子。 |
| `SearchUsers` | Searches for users by name or keyword. | 按名称或关键字搜索用户。 |

## 9. NotificationHandler (Notifications / 通知)
Manages user notifications.
管理用户通知。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `GetNotifications` | Lists user notifications. | 列出用户通知。 |
| `GetUnreadCount` | Gets the count of unread notifications. | 获取未读通知数量。 |
| `MarkAsRead` | Marks a specific notification as read. | 将特定通知标记为已读。 |
| `MarkAllAsRead` | Marks all notifications as read. | 将所有通知标记为已读。 |

## 10. MessageHandler (Direct Messages / 私信)
Handles private messaging between users.
处理用户间的私信。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `GetConversations` | Lists user's conversation threads. | 列出用户的会话列表。 |
| `CreateConversation`| Starts a new conversation with another user. | 与其他用户开始新会话。 |
| `GetMessages` | Retrieves messages within a conversation. | 获取会话中的消息。 |
| `SendMessage` | Sends a new message in a conversation. | 在会话中发送新消息。 |

## 11. AdminHandler (Administration / 管理)
Provides administrative functions for system management.
提供系统管理的后台功能。

| Method | Functionality | 功能 |
| :--- | :--- | :--- |
| `GetDashboard` | Retrieves system statistics for the dashboard. | 获取仪表盘的系统统计数据。 |
| `ListUsers` | Lists all users for management. | 列出所有用户以供管理。 |
| `BanUser` | Bans a user from the platform. | 封禁平台用户。 |
| `UnbanUser` | Unbans a previously banned user. | 解封先前被封禁的用户。 |
| `ListPosts` | Lists all posts for content moderation. | 列出所有帖子以供内容审核。 |
| `BatchReviewPosts` | Reviews multiple posts in a batch operation. | 批量审核多个帖子。 |
| `ListLogs` | Retrieves administrative action logs. | 获取管理操作日志。 |

---
*Generated by Trae AI for Project Airy.*
