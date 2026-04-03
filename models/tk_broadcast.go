package models

import "time"

// WBroadcast 系统广播表。
type WBroadcast struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	Title string `gorm:"size:120;not null" json:"title"`
	// 处理当前语句逻辑。
	Content string `gorm:"size:500;not null" json:"content"`
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
func (WBroadcast) TableName() string { return "tk_broadcast" }
