package models

import "time"

// WDrawResultZodiacTail 生肖/尾数玩法结果表。
// 说明：
// 1. 一条开奖记录对应一条生肖/尾数结果；
// 2. 该表承载六肖、一肖、尾数、生肖连、尾数连相关的基础结果集；
// 3. payload_json 保存命中集合、不中集合、家畜/野兽集合等完整明细。
type WDrawResultZodiacTail struct {
	// ID 为主键。
	ID uint `gorm:"primaryKey"`
	// DrawRecordID 关联开奖记录主表。
	DrawRecordID uint `gorm:"not null;uniqueIndex:uk_tk_draw_result_zodiac_tail_record"`
	// SpecialLotteryID 关联彩种。
	SpecialLotteryID uint `gorm:"not null;index:idx_tk_draw_result_zodiac_tail_lottery_issue,priority:1"`
	// Issue 为开奖期号。
	Issue string `gorm:"size:32;not null;index:idx_tk_draw_result_zodiac_tail_lottery_issue,priority:2"`
	// Year 方便按年份检索。
	Year int `gorm:"not null"`
	// DrawAt 为开奖时间。
	DrawAt time.Time `gorm:"type:datetime(3);not null"`
	// SpecialZodiac 为特码生肖。
	SpecialZodiac string `gorm:"size:16;not null;default:''"`
	// SpecialHomeBeast 为特码家畜/野兽。
	SpecialHomeBeast string `gorm:"size:16;not null;default:''"`
	// SpecialWuxing 为特码五行。
	SpecialWuxing string `gorm:"size:16;not null;default:''"`
	// HitZodiacs 为当期开出过的生肖集合。
	HitZodiacs string `gorm:"size:255;not null;default:''"`
	// MissZodiacs 为当期未开出的生肖集合。
	MissZodiacs string `gorm:"size:255;not null;default:''"`
	// HitTails 为当期开出过的尾数集合。
	HitTails string `gorm:"size:255;not null;default:''"`
	// MissTails 为当期未开出的尾数集合。
	MissTails string `gorm:"size:255;not null;default:''"`
	// HomeBeastZodiacs 为当期开出的家畜生肖集合。
	HomeBeastZodiacs string `gorm:"size:255;not null;default:''"`
	// WildBeastZodiacs 为当期开出的野兽生肖集合。
	WildBeastZodiacs string `gorm:"size:255;not null;default:''"`
	// PayloadJSON 保存完整生肖/尾数结果。
	PayloadJSON string `gorm:"type:longtext;not null"`
	// CreatedAt 为创建时间。
	CreatedAt time.Time `gorm:"type:datetime(3);autoCreateTime"`
	// UpdatedAt 为更新时间。
	UpdatedAt time.Time `gorm:"type:datetime(3);autoUpdateTime"`
}

// TableName 指定生肖/尾数玩法结果表名。
func (WDrawResultZodiacTail) TableName() string {
	return "tk_draw_result_zodiac_tail"
}
