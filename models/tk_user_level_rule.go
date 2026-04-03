package models

import "time"

// WUserLevelRule 用户等级规则表。
type WUserLevelRule struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	LevelNo        int       `gorm:"not null;uniqueIndex" json:"level_no"`
	LevelName      string    `gorm:"size:32;not null" json:"level_name"`
	MinGrowthValue int64     `gorm:"not null;default:0;index" json:"min_growth_value"`
	IconURL        string    `gorm:"size:255;default:''" json:"icon_url"`
	PrivilegesJSON string    `gorm:"type:text" json:"privileges_json"`
	Status         int8      `gorm:"not null;default:1;index" json:"status"`
	Sort           int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WUserLevelRule) TableName() string { return "tk_user_level_rule" }
