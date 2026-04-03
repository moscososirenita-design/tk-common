package models

import "time"

// WUserSystemMessageRead 系统消息已读记录表。
type WUserSystemMessageRead struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index:uk_tk_user_system_message_read_pair,unique,priority:1;index" json:"message_id"`
	UserID    uint      `gorm:"not null;index:uk_tk_user_system_message_read_pair,unique,priority:2;index" json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WUserSystemMessageRead) TableName() string { return "tk_user_system_message_read" }
