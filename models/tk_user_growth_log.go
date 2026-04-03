package models

import "time"

// WUserGrowthLog 用户成长值变更日志。
type WUserGrowthLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	ChangeValue int64     `gorm:"not null" json:"change_value"`
	BizType     string    `gorm:"size:32;not null;default:'';index" json:"biz_type"`
	BizID       uint      `gorm:"not null;default:0" json:"biz_id"`
	Remark      string    `gorm:"size:255;default:''" json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 返回模型对应的数据表名。
func (WUserGrowthLog) TableName() string { return "tk_user_growth_log" }
