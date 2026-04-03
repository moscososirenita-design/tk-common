package models

import "time"

// WLotteryOption 竞猜选项表。
type WLotteryOption struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	LotteryInfoID uint `gorm:"index;not null" json:"lottery_info_id"`
	// 处理当前语句逻辑。
	OptionName string `gorm:"size:32;not null" json:"option_name"`
	// 处理当前语句逻辑。
	Votes int64 `gorm:"not null;default:0" json:"votes"`
	// 处理当前语句逻辑。
	Sort int `gorm:"not null;default:0" json:"sort"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WLotteryOption) TableName() string { return "tk_lottery_option" }
