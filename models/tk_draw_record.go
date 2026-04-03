package models

import "time"

// WDrawRecord 开奖区开奖记录表。
// 说明：
// 1. 本表只服务“首页开奖区/历史开奖/开奖详情”；
// 2. 与 tk_lottery_info（图库图纸）彻底解耦；
// 3. 结构覆盖图2/图3/图4/图5所需字段（6+1号码、历史排序、开奖详情扩展信息）。
type WDrawRecord struct {
	// ID 主键。
	ID uint `gorm:"primaryKey" json:"id"`
	// SpecialLotteryID 所属彩种 ID（关联 tk_special_lottery.id）。
	SpecialLotteryID uint `gorm:"index;not null" json:"special_lottery_id"`
	// Issue 期号（如：2026-063）。
	Issue string `gorm:"size:32;index;not null" json:"issue"`
	// Year 年份（如：2026）。
	Year int `gorm:"index;not null" json:"year"`
	// DrawAt 开奖时间。
	DrawAt time.Time `gorm:"index;not null" json:"draw_at"`

	// NormalDrawResult 普通号码（6 个，逗号分隔）。
	NormalDrawResult string `gorm:"size:64;not null;default:''" json:"normal_draw_result"`
	// SpecialDrawResult 特别号码（1 个）。
	SpecialDrawResult string `gorm:"size:16;not null;default:''" json:"special_draw_result"`
	// DrawResult 兼容字段：完整开奖串（普通 6 个 + 特别号）。
	DrawResult string `gorm:"size:120;not null;default:''" json:"draw_result"`
	// DrawLabels 开奖标签（与号码一一对应，格式示例：羊/土,蛇/金...）。
	DrawLabels string `gorm:"size:255;not null;default:''" json:"draw_labels"`
	// ColorLabels 号码对应波色标签（与号码一一对应，格式示例：绿波,蓝波,红波...）。
	ColorLabels string `gorm:"size:255;not null;default:''" json:"color_labels"`
	// ZodiacLabels 号码对应属相标签（与号码一一对应，格式示例：羊,蛇,马...）。
	ZodiacLabels string `gorm:"size:255;not null;default:''" json:"zodiac_labels"`
	// WuxingLabels 号码对应五行标签（与号码一一对应，格式示例：土,金,火...）。
	WuxingLabels string `gorm:"size:255;not null;default:''" json:"wuxing_labels"`

	// PlaybackURL 开奖回放地址（直播结束后录入）。
	PlaybackURL string `gorm:"size:255;not null;default:''" json:"playback_url"`

	// SpecialSingleDouble 特码单双（如：双）。
	SpecialSingleDouble string `gorm:"size:16;not null;default:''" json:"special_single_double"`
	// SpecialBigSmall 特码大小（如：大）。
	SpecialBigSmall string `gorm:"size:16;not null;default:''" json:"special_big_small"`
	// SumSingleDouble 总和单双（如：双）。
	SumSingleDouble string `gorm:"size:16;not null;default:''" json:"sum_single_double"`
	// SumBigSmall 总和大小（如：大）。
	SumBigSmall string `gorm:"size:16;not null;default:''" json:"sum_big_small"`

	// RecommendSix 六肖推荐（空格分隔）。
	RecommendSix string `gorm:"size:120;not null;default:''" json:"recommend_six"`
	// RecommendFour 四肖推荐（空格分隔）。
	RecommendFour string `gorm:"size:120;not null;default:''" json:"recommend_four"`
	// RecommendOne 一肖推荐。
	RecommendOne string `gorm:"size:32;not null;default:''" json:"recommend_one"`
	// RecommendTen 十码推荐（空格分隔）。
	RecommendTen string `gorm:"size:255;not null;default:''" json:"recommend_ten"`

	// SpecialCode 特码（数字）。
	SpecialCode string `gorm:"size:16;not null;default:''" json:"special_code"`
	// NormalCode 正码（逗号分隔）。
	NormalCode string `gorm:"size:120;not null;default:''" json:"normal_code"`
	// Zheng1 正1特描述（如：大双,合双,尾大,蓝波）。
	Zheng1 string `gorm:"size:120;not null;default:''" json:"zheng1"`
	// Zheng2 正2特描述。
	Zheng2 string `gorm:"size:120;not null;default:''" json:"zheng2"`
	// Zheng3 正3特描述。
	Zheng3 string `gorm:"size:120;not null;default:''" json:"zheng3"`
	// Zheng4 正4特描述。
	Zheng4 string `gorm:"size:120;not null;default:''" json:"zheng4"`
	// Zheng5 正5特描述。
	Zheng5 string `gorm:"size:120;not null;default:''" json:"zheng5"`
	// Zheng6 正6特描述。
	Zheng6 string `gorm:"size:120;not null;default:''" json:"zheng6"`

	// IsCurrent 是否当前期（1 是，0 否）。
	IsCurrent int8 `gorm:"not null;default:0" json:"is_current"`
	// Status 状态（1 启用，0 停用）。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// Sort 排序值（越小越靠前）。
	Sort int `gorm:"not null;default:0" json:"sort"`

	// CreatedAt 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WDrawRecord) TableName() string { return "tk_draw_record" }
