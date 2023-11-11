package models

import "time"

type Users struct {
	ID        uint64    `gorm:"primaryKey"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"createdAt"`
	UpdatedAt time.Time `db:"updatedAt"`
}

type Posts struct {
	Post_id      uint64    `db:"post_id"`
	Author_id    uint64    `db:"author_id"`
	Post_title   string    `db:"post_title"`
	Post_content string    `db:"post_content"`
	Post_likes   uint32    `db:"post_likes"`
	CreatedAt    time.Time `db:"createdat"`
	UpdatedAt    time.Time `db:"updatedat"`
}

type Comment struct {
	Comment_id      uint64    `db:"comment_id"`
	Post_id         uint64    `db:"post_id"`
	User_id         uint64    `db:"user_id"`
	Username        string    `db:"username"`
	Comment_content string    `db:"comment_content"`
	Comment_likes   uint64    `db:"comment_likes"`
	CreatedAt       time.Time `db:"createdat"`
	UpdatedAt       time.Time `db:"updatedat"`
}

type UserAndPost struct {
	User  Users
	Posts []Posts
}
type UsernameAndPost struct {
	Username string
	Post     Posts
}

type PostUsernameComments_WithUserPreference struct {
	PostAndUserPreferences PostAndUserPreferences
	Comments               []UsernameAndComment
}
type PostUsernameComments struct {
	Post     Posts
	Username string
	Comments []UsernameAndComment
}
type UsernameAndComment struct {
	User_id         uint64
	Username        string
	CommentID       uint64
	Comment_content string
	Comment_likes   uint64
}
type Passanduserid struct {
	Password string `db:"password"`
	User_id  uint64 `db:"user_id"`
}
type PostAndUserPreferences struct {
	Post               Posts
	Username           string
	PostLikedByUser    bool
	PostDislikedByUser bool
}
