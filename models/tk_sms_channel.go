package models

import "time"

// WSMSChannel 短信服务通道配置。
type WSMSChannel struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// Provider 提供商标识：aliyun/tencent/twilio/custom。
	Provider string `gorm:"size:32;not null;default:'custom'" json:"provider"`
	// ChannelName 通道名称，便于后台辨识。
	ChannelName string `gorm:"size:64;not null;default:''" json:"channel_name"`
	// AccessKey 接口凭证 key。
	AccessKey string `gorm:"size:128;default:''" json:"access_key"`
	// AccessSecret 接口凭证 secret。
	AccessSecret string `gorm:"size:255;default:''" json:"access_secret"`
	// Endpoint 服务网关地址。
	Endpoint string `gorm:"size:255;default:''" json:"endpoint"`
	// SignName 短信签名。
	SignName string `gorm:"size:64;default:''" json:"sign_name"`
	// TemplateCodeLogin 登录验证码模板编码。
	TemplateCodeLogin string `gorm:"size:64;default:''" json:"template_code_login"`
	// TemplateCodeRegister 注册验证码模板编码。
	TemplateCodeRegister string `gorm:"size:64;default:''" json:"template_code_register"`
	// DailyLimit 单手机号日发送上限。
	DailyLimit int `gorm:"not null;default:20" json:"daily_limit"`
	// MinuteLimit 单手机号分钟级发送上限。
	MinuteLimit int `gorm:"not null;default:1" json:"minute_limit"`
	// CodeTTLSeconds 验证码有效时长（秒）。
	CodeTTLSeconds int `gorm:"not null;default:300" json:"code_ttl_seconds"`
	// MockMode 是否使用模拟发送：1是，0否。
	MockMode int8 `gorm:"not null;default:1" json:"mock_mode"`
	// Status 状态：1启用，0停用。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WSMSChannel) TableName() string { return "tk_sms_channel" }
