package models

import "time"

// WDrawResultCombo 组合玩法结果表。
// 说明：
// 1. 一条开奖记录对应一条组合玩法结果；
// 2. 该表承载连码、过关、二中特、特串、不、中多选中一、特平中等玩法的结算基础数据；
// 3. payload_json 保存每类玩法所需的命中基准集合，避免在不同接口中重复推导。
type WDrawResultCombo struct {
	// ID 为主键。
	ID uint `gorm:"primaryKey"`
	// DrawRecordID 关联开奖记录主表。
	DrawRecordID uint `gorm:"not null;uniqueIndex:uk_tk_draw_result_combo_record"`
	// SpecialLotteryID 关联彩种。
	SpecialLotteryID uint `gorm:"not null;index:idx_tk_draw_result_combo_lottery_issue,priority:1"`
	// Issue 为开奖期号。
	Issue string `gorm:"size:32;not null;index:idx_tk_draw_result_combo_lottery_issue,priority:2"`
	// Year 方便按年份检索。
	Year int `gorm:"not null"`
	// DrawAt 为开奖时间。
	DrawAt time.Time `gorm:"type:datetime(3);not null"`
	// NormalNumbers 为前6个正码集合。
	NormalNumbers string `gorm:"size:64;not null;default:''"`
	// AllNumbers 为全部7个开奖号码集合。
	AllNumbers string `gorm:"size:80;not null;default:''"`
	// SpecialNumber 为特码号码。
	SpecialNumber int `gorm:"not null;default:0"`
	// PayloadJSON 保存组合玩法结算基准。
	PayloadJSON string `gorm:"type:longtext;not null"`
	// CreatedAt 为创建时间。
	CreatedAt time.Time `gorm:"type:datetime(3);autoCreateTime"`
	// UpdatedAt 为更新时间。
	UpdatedAt time.Time `gorm:"type:datetime(3);autoUpdateTime"`
}

// TableName 指定组合玩法结果表名。
func (WDrawResultCombo) TableName() string {
	return "tk_draw_result_combo"
}
