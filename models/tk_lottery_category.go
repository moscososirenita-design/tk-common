package models

import "time"

// WLotteryCategory 图库分类配置表。
// 用途：
// 1. 首页分类条按排序展示；
// 2. “更多分类”页可搜索；
// 3. 支持后台启停与是否首页展示控制。
type WLotteryCategory struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	CategoryKey string `gorm:"column:category_key;size:32;uniqueIndex;not null" json:"category_key"`
	// 处理当前语句逻辑。
	Name string `gorm:"size:32;not null" json:"name"`
	// 处理当前语句逻辑。
	SearchKeywords string `gorm:"size:255;not null;default:''" json:"search_keywords"`
	// 处理当前语句逻辑。
	ShowOnHome int8 `gorm:"not null;default:1" json:"show_on_home"`
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
func (WLotteryCategory) TableName() string { return "tk_lottery_category" }
