package models

import "time"

// WCommentLike 评论点赞关系表。
type WCommentLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CommentID uint      `gorm:"not null;index:idx_tk_comment_like_comment,priority:1;index:uk_tk_comment_like_pair,unique,priority:1" json:"comment_id"`
	UserID    uint      `gorm:"not null;index:idx_tk_comment_like_user,priority:1;index:uk_tk_comment_like_pair,unique,priority:2" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WCommentLike) TableName() string { return "tk_comment_like" }
