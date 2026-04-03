package models

import "time"

// WLotteryVoteRecord 竞猜投票记录表。
// 通过 (lottery_info_id + voter_hash) 唯一约束防止同一设备/指纹重复投票。
type WLotteryVoteRecord struct {
	// 处理当前语句逻辑。
	ID uint `gorm:"primaryKey" json:"id"`
	// 处理当前语句逻辑。
	LotteryInfoID uint `gorm:"uniqueIndex:uk_lottery_voter;not null" json:"lottery_info_id"`
	// 处理当前语句逻辑。
	OptionID uint `gorm:"not null" json:"option_id"`
	// 处理当前语句逻辑。
	VoterHash string `gorm:"size:64;uniqueIndex:uk_lottery_voter;not null" json:"voter_hash"`
	// 处理当前语句逻辑。
	DeviceID string `gorm:"size:120" json:"device_id"`
	// 处理当前语句逻辑。
	ClientIP string `gorm:"size:64" json:"client_ip"`
	// 处理当前语句逻辑。
	UserAgent string `gorm:"size:255" json:"user_agent"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回模型对应的数据表名。
func (WLotteryVoteRecord) TableName() string { return "tk_lottery_vote_record" }
