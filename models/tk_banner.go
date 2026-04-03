package models

import "time"

// WBanner Banner 配置表：
// - type: ad / official
// - position: home / lottery_detail / post_detail
// 用一张表承载多个展示区与类型配置。
type WBanner struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	Title string `gorm:"size:120;not null" json:"title"`
	// 处理当前语句逻辑。
	ImageURL string `gorm:"size:255;not null" json:"image_url"`
	// 处理当前语句逻辑。
	LinkURL string `gorm:"size:255" json:"link_url"`
	// 处理当前语句逻辑。
	Type string `gorm:"size:32;index;not null" json:"type"`
	// 处理当前语句逻辑。
	Position string `gorm:"size:32;index;not null" json:"position"`
	// 处理当前语句逻辑。
	Positions string `gorm:"size:255;not null;default:''" json:"positions"`
	// 处理当前语句逻辑。
	JumpType string `gorm:"size:20;index;not null;default:'none'" json:"jump_type"`
	// 处理当前语句逻辑。
	JumpPostID uint `gorm:"not null;default:0" json:"jump_post_id"`
	// 处理当前语句逻辑。
	JumpURL string `gorm:"size:255;not null;default:''" json:"jump_url"`
	// 处理当前语句逻辑。
	ContentHTML string `gorm:"type:longtext" json:"content_html"`
	// 处理当前语句逻辑。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// 处理当前语句逻辑。
	Sort int `gorm:"not null;default:0" json:"sort"`
	// 处理当前语句逻辑。
	StartAt *time.Time `json:"start_at"`
	// 处理当前语句逻辑。
	EndAt *time.Time `json:"end_at"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WBanner) TableName() string { return "tk_banner" }
