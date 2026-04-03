package models

import "time"

// WDrawResultSpecial 特码玩法结果表。
// 说明：
// 1. 一条开奖记录对应一条特码玩法结果；
// 2. 该表承载特码两面、波色、合数单双、尾大尾小、半波、特码生肖、五行等结果；
// 3. payload_json 用于保留完整的结构化结果，便于后续业务扩展而不频繁改表。
type WDrawResultSpecial struct {
	// ID 为主键。
	ID uint `gorm:"primaryKey"`
	// DrawRecordID 关联开奖记录主表。
	DrawRecordID uint `gorm:"not null;uniqueIndex:uk_tk_draw_result_special_record"`
	// SpecialLotteryID 关联彩种。
	SpecialLotteryID uint `gorm:"not null;index:idx_tk_draw_result_special_lottery_issue,priority:1"`
	// Issue 为开奖期号。
	Issue string `gorm:"size:32;not null;index:idx_tk_draw_result_special_lottery_issue,priority:2"`
	// Year 方便按年份检索。
	Year int `gorm:"not null"`
	// DrawAt 为开奖时间。
	DrawAt time.Time `gorm:"type:datetime(3);not null"`
	// SpecialNumber 为特码号码。
	SpecialNumber int `gorm:"not null"`
	// SpecialColorWave 为特码波色。
	SpecialColorWave string `gorm:"size:16;not null;default:''"`
	// SpecialBigSmall 为特码大小。
	SpecialBigSmall string `gorm:"size:16;not null;default:''"`
	// SpecialSingleDouble 为特码单双。
	SpecialSingleDouble string `gorm:"size:16;not null;default:''"`
	// SpecialSumSingleDouble 为特码合数单双。
	SpecialSumSingleDouble string `gorm:"size:16;not null;default:''"`
	// SpecialTailBigSmall 为特码尾大尾小。
	SpecialTailBigSmall string `gorm:"size:16;not null;default:''"`
	// SpecialZodiac 为特码生肖。
	SpecialZodiac string `gorm:"size:16;not null;default:''"`
	// SpecialWuxing 为特码五行。
	SpecialWuxing string `gorm:"size:16;not null;default:''"`
	// SpecialHomeBeast 为特码家畜/野兽。
	SpecialHomeBeast string `gorm:"size:16;not null;default:''"`
	// HalfWaveColorSize 为特码半波（波色+大小）。
	HalfWaveColorSize string `gorm:"size:32;not null;default:''"`
	// HalfWaveColorParity 为特码半波（波色+单双）。
	HalfWaveColorParity string `gorm:"size:32;not null;default:''"`
	// PayloadJSON 保存完整结构化结果。
	PayloadJSON string `gorm:"type:longtext;not null"`
	// CreatedAt 为创建时间。
	CreatedAt time.Time `gorm:"type:datetime(3);autoCreateTime"`
	// UpdatedAt 为更新时间。
	UpdatedAt time.Time `gorm:"type:datetime(3);autoUpdateTime"`
}

// TableName 指定特码玩法结果表名。
func (WDrawResultSpecial) TableName() string {
	return "tk_draw_result_special"
}
