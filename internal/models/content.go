package models

import "time"

// Post represents a post/article in the system
type Post struct {
	ID              int64      `gorm:"primaryKey" json:"id"`
	Title           string     `gorm:"size:255;not null" json:"title"`
	ContentMarkdown string     `gorm:"type:text;not null" json:"content_markdown"`
	ContentHTML     string     `gorm:"type:text;not null" json:"content_html"`
	Summary         string     `gorm:"size:500" json:"summary"`
	CoverImage      string     `gorm:"size:255" json:"cover_image"`
	AuthorID        int64      `gorm:"index;not null" json:"author_id"`
	CircleID        *int64     `gorm:"index" json:"circle_id"`
	Status          string     `gorm:"size:20;index;default:'draft'" json:"status"` // draft, pending, published, hidden, deleted
	Category        string     `gorm:"size:50" json:"category"`
	Tags            string     `gorm:"type:json" json:"tags"` // JSON array
	ScheduledAt     *time.Time `json:"scheduled_at"`
	IsPinned        bool       `gorm:"default:false" json:"is_pinned"`
	IsFeatured      bool       `gorm:"default:false" json:"is_featured"`
	AllowComment    bool       `gorm:"default:true" json:"allow_comment"`
	IsAnonymous     bool       `gorm:"default:false" json:"is_anonymous"`
	ViewCount       int        `gorm:"default:0" json:"view_count"`
	HotnessScore    float64    `gorm:"index;default:0" json:"hotness_score"`
	PublishedAt     *time.Time `json:"published_at"`
	CreatedAt       time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName specifies the table name for Post model
func (Post) TableName() string {
	return "posts"
}

// Comment represents a comment on a post or another comment
type Comment struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	AuthorID  int64      `gorm:"index;not null" json:"author_id"`
	PostID    int64      `gorm:"index;not null" json:"post_id"`
	ParentID  *int64     `gorm:"index" json:"parent_id"`
	RootID    int64      `gorm:"index;not null" json:"root_id"` // Root comment ID
	Level     int        `gorm:"default:0" json:"level"`
	Path      string     `gorm:"size:255;index" json:"path"` // Path enumeration, e.g., "1.2.5"
	Status    string     `gorm:"size:20;default:'published'" json:"status"` // published, hidden, deleted
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName specifies the table name for Comment model
func (Comment) TableName() string {
	return "comments"
}

// Vote represents a vote on a post or comment
type Vote struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UserID     int64     `gorm:"uniqueIndex:idx_user_entity;not null" json:"user_id"`
	EntityType string    `gorm:"size:20;uniqueIndex:idx_user_entity;not null" json:"entity_type"` // post, comment
	EntityID   int64     `gorm:"uniqueIndex:idx_user_entity;not null" json:"entity_id"`
	VoteType   string    `gorm:"size:10;not null" json:"vote_type"` // up, down
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName specifies the table name for Vote model
func (Vote) TableName() string {
	return "votes"
}

// Favorite represents a user's favorite post
type Favorite struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"uniqueIndex:idx_user_post;not null" json:"user_id"`
	PostID    int64     `gorm:"uniqueIndex:idx_user_post;index;not null" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for Favorite model
func (Favorite) TableName() string {
	return "favorites"
}

// EntityCount represents aggregated counts for posts and comments
type EntityCount struct {
	EntityType    string    `gorm:"primaryKey;size:20;not null" json:"entity_type"` // post, comment
	EntityID      int64     `gorm:"primaryKey;not null" json:"entity_id"`
	UpvoteCount   int       `gorm:"default:0" json:"upvote_count"`
	DownvoteCount int       `gorm:"default:0" json:"downvote_count"`
	CommentCount  int       `gorm:"default:0" json:"comment_count"`
	FavoriteCount int       `gorm:"default:0" json:"favorite_count"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the table name for EntityCount model
func (EntityCount) TableName() string {
	return "entity_counts"
}
