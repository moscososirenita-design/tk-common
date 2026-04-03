package models

import "time"

// WUser 客户端用户表。
type WUser struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// Username 用户名（全局唯一，兼容历史账号体系）。
	Username string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	// Phone 手机号（用于手机号注册/登录，允许为空）。
	Phone string `gorm:"size:20;uniqueIndex;default:''" json:"phone"`
	// Nickname 昵称。
	Nickname string `gorm:"size:64" json:"nickname"`
	// Avatar 头像地址。
	Avatar string `gorm:"size:255" json:"avatar"`
	// PasswordHash 密码哈希（bcrypt）。
	PasswordHash string `gorm:"size:255;default:''" json:"password_hash"`
	// RegisterSource 注册来源：password/sms/admin/import。
	RegisterSource string `gorm:"size:20;not null;default:'password'" json:"register_source"`
	// LastLoginAt 最近登录时间。
	LastLoginAt *time.Time `json:"last_login_at"`
	// UserType 用户类型：natural/official/robot。
	UserType string `gorm:"size:20;index;not null;default:'natural'" json:"user_type"`
	// FansCount 粉丝数（用户模块统计）。
	FansCount int64 `gorm:"not null;default:0" json:"fans_count"`
	// FollowingCount 关注数（用户模块统计）。
	FollowingCount int64 `gorm:"not null;default:0" json:"following_count"`
	// GrowthValue 成长值（用户模块统计）。
	GrowthValue int64 `gorm:"not null;default:0" json:"growth_value"`
	// ReadPostCount 阅读帖子数（用户模块统计）。
	ReadPostCount int64 `gorm:"not null;default:0" json:"read_post_count"`
	// Status 账号状态：1启用，0停用。
	Status int8 `gorm:"not null;default:1" json:"status"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WUser) TableName() string { return "tk_users" }
