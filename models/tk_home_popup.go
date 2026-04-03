package models

import "time"

// WHomePopup 首页首屏弹窗配置。
type WHomePopup struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// Title 弹窗标题。
	Title string `gorm:"size:120;not null" json:"title"`
	// Content 弹窗正文（支持富文本）。
	Content string `gorm:"type:text" json:"content"`
	// ImageURL 弹窗配图地址。
	ImageURL string `gorm:"size:255;default:''" json:"image_url"`
	// ButtonText 按钮文案。
	ButtonText string `gorm:"size:40;default:''" json:"button_text"`
	// ButtonLink 按钮跳转地址。
	ButtonLink string `gorm:"size:255;default:''" json:"button_link"`
	// Position 展示位置，默认 home。
	Position string `gorm:"size:32;not null;default:'home'" json:"position"`
	// ShowOnce 是否每设备只显示一次：1是，0否。
	ShowOnce int8 `gorm:"not null;default:1" json:"show_once"`
	// Status 状态：1启用，0停用。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// Sort 排序值，越小越靠前。
	Sort int `gorm:"not null;default:0" json:"sort"`
	// StartAt 生效开始时间。
	StartAt *time.Time `json:"start_at"`
	// EndAt 生效结束时间。
	EndAt *time.Time `json:"end_at"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WHomePopup) TableName() string { return "tk_home_popup" }
