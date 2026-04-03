package models

import "time"

// WExternalLink 外链配置表。
type WExternalLink struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	Name string `gorm:"size:80;not null" json:"name"`
	// 处理当前语句逻辑。
	URL string `gorm:"size:255;not null" json:"url"`
	// 处理当前语句逻辑。
	Position string `gorm:"size:32;index;not null" json:"position"`
	// 处理当前语句逻辑。
	IconURL string `gorm:"size:255;not null;default:''" json:"icon_url"`
	// 处理当前语句逻辑。
	GroupKey string `gorm:"size:32;index;not null;default:''" json:"group_key"`
	// 处理当前语句逻辑。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// 处理当前语句逻辑。
	Sort int `gorm:"not null;default:0" json:"sort"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WExternalLink) TableName() string { return "tk_external_link" }
