package models

import "time"

// WUserSystemMessage 用户系统消息表（支持全员广播）。
type WUserSystemMessage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ReceiverUserID uint      `gorm:"not null;default:0;index" json:"receiver_user_id"`
	Title          string    `gorm:"size:200;not null" json:"title"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	MessageType    string    `gorm:"size:20;not null;default:'notice';index" json:"message_type"`
	Level          string    `gorm:"size:20;not null;default:'info'" json:"level"`
	BizType        string    `gorm:"size:32;not null;default:'';index" json:"biz_type"`
	BizID          uint      `gorm:"not null;default:0" json:"biz_id"`
	Status         int8      `gorm:"not null;default:1;index" json:"status"`
	SentAt         time.Time `json:"sent_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WUserSystemMessage) TableName() string { return "tk_user_system_message" }
