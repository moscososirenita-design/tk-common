package models

import "time"

// WDrawResultRegular 正码玩法结果表。
// 说明：
// 1. 一条开奖记录对应一条正码玩法结果；
// 2. 该表承载总分大小单双、正码1-6 的位置结果；
// 3. payload_json 保留全部正码位置的结构化结果。
type WDrawResultRegular struct {
	// ID 为主键。
	ID uint `gorm:"primaryKey"`
	// DrawRecordID 关联开奖记录主表。
	DrawRecordID uint `gorm:"not null;uniqueIndex:uk_tk_draw_result_regular_record"`
	// SpecialLotteryID 关联彩种。
	SpecialLotteryID uint `gorm:"not null;index:idx_tk_draw_result_regular_lottery_issue,priority:1"`
	// Issue 为开奖期号。
	Issue string `gorm:"size:32;not null;index:idx_tk_draw_result_regular_lottery_issue,priority:2"`
	// Year 方便按年份检索。
	Year int `gorm:"not null"`
	// DrawAt 为开奖时间。
	DrawAt time.Time `gorm:"type:datetime(3);not null"`
	// NormalNumbers 为前6个正码。
	NormalNumbers string `gorm:"size:64;not null;default:''"`
	// TotalSum 为7个开奖号码总和。
	TotalSum int `gorm:"not null;default:0"`
	// TotalBigSmall 为总分大小。
	TotalBigSmall string `gorm:"size:16;not null;default:''"`
	// TotalSingleDouble 为总分单双。
	TotalSingleDouble string `gorm:"size:16;not null;default:''"`
	// Zheng1JSON 为正1结构化结果。
	Zheng1JSON string `gorm:"type:longtext;not null"`
	// Zheng2JSON 为正2结构化结果。
	Zheng2JSON string `gorm:"type:longtext;not null"`
	// Zheng3JSON 为正3结构化结果。
	Zheng3JSON string `gorm:"type:longtext;not null"`
	// Zheng4JSON 为正4结构化结果。
	Zheng4JSON string `gorm:"type:longtext;not null"`
	// Zheng5JSON 为正5结构化结果。
	Zheng5JSON string `gorm:"type:longtext;not null"`
	// Zheng6JSON 为正6结构化结果。
	Zheng6JSON string `gorm:"type:longtext;not null"`
	// PayloadJSON 保存完整正码结果。
	PayloadJSON string `gorm:"type:longtext;not null"`
	// CreatedAt 为创建时间。
	CreatedAt time.Time `gorm:"type:datetime(3);autoCreateTime"`
	// UpdatedAt 为更新时间。
	UpdatedAt time.Time `gorm:"type:datetime(3);autoUpdateTime"`
}

// TableName 指定正码玩法结果表名。
func (WDrawResultRegular) TableName() string {
	return "tk_draw_result_regular"
}
