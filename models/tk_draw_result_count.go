package models

import "time"

// WDrawResultCount 统计玩法结果表。
// 说明：
// 1. 一条开奖记录对应一条统计结果；
// 2. 该表承载七码、一肖量、尾数量、五行覆盖等计数型结果；
// 3. payload_json 保存完整统计明细。
type WDrawResultCount struct {
	// ID 为主键。
	ID uint `gorm:"primaryKey"`
	// DrawRecordID 关联开奖记录主表。
	DrawRecordID uint `gorm:"not null;uniqueIndex:uk_tk_draw_result_count_record"`
	// SpecialLotteryID 关联彩种。
	SpecialLotteryID uint `gorm:"not null;index:idx_tk_draw_result_count_lottery_issue,priority:1"`
	// Issue 为开奖期号。
	Issue string `gorm:"size:32;not null;index:idx_tk_draw_result_count_lottery_issue,priority:2"`
	// Year 方便按年份检索。
	Year int `gorm:"not null"`
	// DrawAt 为开奖时间。
	DrawAt time.Time `gorm:"type:datetime(3);not null"`
	// TotalSum 为7码总和。
	TotalSum int `gorm:"not null;default:0"`
	// OddCount 为七码单数数量（49 按单计）。
	OddCount int `gorm:"not null;default:0"`
	// EvenCount 为七码双数数量。
	EvenCount int `gorm:"not null;default:0"`
	// BigCount 为七码大数数量（49 按大计）。
	BigCount int `gorm:"not null;default:0"`
	// SmallCount 为七码小数数量。
	SmallCount int `gorm:"not null;default:0"`
	// DistinctZodiacCount 为不同生肖总数。
	DistinctZodiacCount int `gorm:"not null;default:0"`
	// DistinctTailCount 为不同尾数总数。
	DistinctTailCount int `gorm:"not null;default:0"`
	// DistinctWuxingCount 为不同五行总数。
	DistinctWuxingCount int `gorm:"not null;default:0"`
	// AppearedZodiacs 为当期开出过的生肖集合。
	AppearedZodiacs string `gorm:"size:255;not null;default:''"`
	// MissedZodiacs 为当期未开出的生肖集合。
	MissedZodiacs string `gorm:"size:255;not null;default:''"`
	// AppearedTails 为当期开出过的尾数集合。
	AppearedTails string `gorm:"size:255;not null;default:''"`
	// MissedTails 为当期未开出的尾数集合。
	MissedTails string `gorm:"size:255;not null;default:''"`
	// AppearedWuxings 为当期开出过的五行集合。
	AppearedWuxings string `gorm:"size:255;not null;default:''"`
	// PayloadJSON 保存完整统计结果。
	PayloadJSON string `gorm:"type:longtext;not null"`
	// CreatedAt 为创建时间。
	CreatedAt time.Time `gorm:"type:datetime(3);autoCreateTime"`
	// UpdatedAt 为更新时间。
	UpdatedAt time.Time `gorm:"type:datetime(3);autoUpdateTime"`
}

// TableName 指定统计玩法结果表名。
//
//	统计玩法结果表名。
func (WDrawResultCount) TableName() string {
	return "tk_draw_result_count"
}
