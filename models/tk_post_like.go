package models

import "time"

// WPostLike 帖子点赞关系表。
type WPostLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;index:idx_tk_post_like_post,priority:1;index:uk_tk_post_like_pair,unique,priority:1" json:"post_id"`
	UserID    uint      `gorm:"not null;index:idx_tk_post_like_user,priority:1;index:uk_tk_post_like_pair,unique,priority:2" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WPostLike) TableName() string { return "tk_post_like" }
