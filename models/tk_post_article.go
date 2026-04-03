package models

import "time"

// WPostArticle 帖子表（官方发帖+用户发帖）。
type WPostArticle struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	LotteryInfoID uint `gorm:"index;not null;default:0" json:"lottery_info_id"`
	// 处理当前语句逻辑。
	UserID uint `gorm:"index;not null;default:0" json:"user_id"`
	// 处理当前语句逻辑。
	Title string `gorm:"size:160;not null" json:"title"`
	// 处理当前语句逻辑。
	CoverImage string `gorm:"size:255" json:"cover_image"`
	// 处理当前语句逻辑。
	Content string `gorm:"type:text" json:"content"`
	// 处理当前语句逻辑。
	IsOfficial int8 `gorm:"not null;default:0" json:"is_official"`
	// 处理当前语句逻辑。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WPostArticle) TableName() string { return "tk_post_article" }
