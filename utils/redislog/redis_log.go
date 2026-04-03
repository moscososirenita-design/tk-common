package redislog

import (
	"strings"

	commonlogx "github.com/moscososirenita-design/tk-common/utils/logx"
)

// WarnOp 输出 Redis 操作告警日志。
func WarnOp(op string, key string, err error) {
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 无错误时无需记录。
		return
	}
	// 记录 Redis 告警日志。
	commonlogx.GetLogger().Warn("redis op=%s key=%s err=%v", strings.TrimSpace(op), strings.TrimSpace(key), err)
}

// ErrorOp 输出 Redis 操作错误日志。
func ErrorOp(op string, key string, err error) {
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 无错误时无需记录。
		return
	}
	// 记录 Redis 错误日志。
	commonlogx.GetLogger().Error("redis op=%s key=%s err=%v", strings.TrimSpace(op), strings.TrimSpace(key), err)
}
