package models

import "time"

// WSpecialLottery 首页顶部“澳彩/港彩”切换信息表。
type WSpecialLottery struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	Name string `gorm:"size:64;not null" json:"name"`
	// 处理当前语句逻辑。
	Code string `gorm:"size:32;uniqueIndex;not null" json:"code"`
	// 处理当前语句逻辑。
	CurrentIssue string `gorm:"size:32" json:"current_issue"`
	// 处理当前语句逻辑。
	NextDrawAt time.Time `json:"next_draw_at"`
	// 处理当前语句逻辑。
	LiveEnabled int8 `gorm:"not null;default:0" json:"live_enabled"`
	// 处理当前语句逻辑。
	LiveStatus string `gorm:"size:16;not null;default:'pending'" json:"live_status"`
	// 处理当前语句逻辑。
	LiveStreamURL string `gorm:"size:255" json:"live_stream_url"`
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
func (WSpecialLottery) TableName() string { return "tk_special_lottery" }
