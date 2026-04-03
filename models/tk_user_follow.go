package models

import "time"

// WUserFollow 用户关注关系表。
type WUserFollow struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	FollowerUserID  uint      `gorm:"not null;index:idx_tk_user_follow_follower,priority:1;index:uk_tk_user_follow_pair,unique,priority:1" json:"follower_user_id"`
	FollowingUserID uint      `gorm:"not null;index:idx_tk_user_follow_following,priority:1;index:uk_tk_user_follow_pair,unique,priority:2" json:"following_user_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WUserFollow) TableName() string { return "tk_user_follow" }
