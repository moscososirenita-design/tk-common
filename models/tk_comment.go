package models

import "time"

// WComment 评论表（支持 parent_id 追评）。
type WComment struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	PostID uint `gorm:"index;not null;default:0" json:"post_id"`
	// 处理当前语句逻辑。
	LotteryInfoID uint `gorm:"index;not null;default:0" json:"lottery_info_id"`
	// 处理当前语句逻辑。
	UserID uint `gorm:"index;not null" json:"user_id"`
	// 处理当前语句逻辑。
	ParentID uint `gorm:"index;default:0" json:"parent_id"`
	// 处理当前语句逻辑。
	Content string `gorm:"size:1000;not null" json:"content"`
	// 处理当前语句逻辑。
	Likes int64 `gorm:"not null;default:0" json:"likes"`
	// 处理当前语句逻辑。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WComment) TableName() string { return "tk_comment" }
