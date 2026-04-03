package models

import "time"

// WUserPrivateMessage 用户私信表。
type WUserPrivateMessage struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ConversationKey string     `gorm:"size:64;not null;index" json:"conversation_key"`
	SenderUserID    uint       `gorm:"not null;index" json:"sender_user_id"`
	ReceiverUserID  uint       `gorm:"not null;index" json:"receiver_user_id"`
	Content         string     `gorm:"size:2000;not null" json:"content"`
	MessageType     string     `gorm:"size:20;not null;default:'text'" json:"message_type"`
	Status          int8       `gorm:"not null;default:1;index" json:"status"`
	IsRead          int8       `gorm:"not null;default:0;index" json:"is_read"`
	ReadAt          *time.Time `json:"read_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WUserPrivateMessage) TableName() string { return "tk_user_private_message" }
