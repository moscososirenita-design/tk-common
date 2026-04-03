package models

import "time"

// WLotteryInfo 图库图纸与详情内容表。
type WLotteryInfo struct {
	// ID 主键。
	ID uint `gorm:"primaryKey" json:"id"`
	// SpecialLotteryID 所属彩种 ID（关联 tk_special_lottery.id，0 表示不绑定彩种）。
	SpecialLotteryID uint `gorm:"index;not null" json:"special_lottery_id"`
	// CategoryID 图库分类 ID（关联 tk_lottery_category.id）。
	CategoryID uint `gorm:"index;not null;default:0" json:"category_id"`
	// CategoryTag 兼容字段：分类标识（通常等于 category_key）。
	CategoryTag string `gorm:"size:32;index;not null;default:''" json:"category_tag"`
	// Issue 开奖期号（如：2026-024）。
	Issue string `gorm:"size:32;index;not null" json:"issue"`
	// Year 年份（如：2026）。
	Year int `gorm:"index;not null" json:"year"`
	// Title 图纸标题。
	Title string `gorm:"size:120;not null" json:"title"`
	// CoverImageURL 列表封面图地址。
	CoverImageURL string `gorm:"size:255;not null" json:"cover_image_url"`
	// DetailImageURL 详情主图地址。
	DetailImageURL string `gorm:"size:255;not null" json:"detail_image_url"`
	// DrawCode 暗码。
	DrawCode string `gorm:"size:120" json:"draw_code"`
	// NormalDrawResult 普通号码（6 个，逗号分隔）。
	NormalDrawResult string `gorm:"size:64;not null;default:''" json:"normal_draw_result"`
	// SpecialDrawResult 特别号码（1 个）。
	SpecialDrawResult string `gorm:"size:16;not null;default:''" json:"special_draw_result"`
	// DrawResult 兼容字段：完整开奖串（普通 6 个 + 特别号）。
	DrawResult string `gorm:"size:120" json:"draw_result"`
	// DrawAt 开奖时间。
	DrawAt time.Time `json:"draw_at"`
	// PlaybackURL 开奖回放地址（直播结束后录入）。
	PlaybackURL string `gorm:"size:255;not null;default:''" json:"playback_url"`
	// IsCurrent 是否当前期（1 是，0 否）。
	IsCurrent int8 `gorm:"not null;default:0" json:"is_current"`
	// Status 状态（1 启用，0 停用）。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// Sort 排序值（越小越靠前）。
	Sort int `gorm:"not null;default:0" json:"sort"`
	// LikesCount 点赞数。
	LikesCount int64 `gorm:"not null;default:0" json:"likes_count"`
	// CommentCount 评论数。
	CommentCount int64 `gorm:"not null;default:0" json:"comment_count"`
	// FavoriteCount 收藏数。
	FavoriteCount int64 `gorm:"not null;default:0" json:"favorite_count"`
	// ReadCount 阅读数。
	ReadCount int64 `gorm:"not null;default:0" json:"read_count"`
	// PollEnabled 投票区开关。
	PollEnabled int8 `gorm:"not null;default:1" json:"poll_enabled"`
	// PollDefaultExpand 投票区默认展开状态。
	PollDefaultExpand int8 `gorm:"not null;default:0" json:"poll_default_expand"`
	// RecommendInfoIDs 兼容字段：推荐图纸 ID 串。
	RecommendInfoIDs string `gorm:"size:255;not null;default:''" json:"recommend_info_ids"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WLotteryInfo) TableName() string { return "tk_lottery_info" }
