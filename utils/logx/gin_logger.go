package logx

import "go.uber.org/zap"

// GetGinLogger 返回专用于 gin 访问日志的 *zap.Logger。
// 该 Logger 不携带调用者信息（WithCaller(false)），避免输出 gin 内部框架路径。
// 在 gin 中间件中使用示例：
//
//	ginLogger := logx.GetGinLogger()
//	ginLogger.Info("request", zap.String("path", c.Request.URL.Path), ...)
func GetGinLogger() *zap.Logger {
	globalMu.RLock()
	l := ginZapLogger
	globalMu.RUnlock()
	if l != nil {
		return l
	}
	return GetZapLogger()
}
