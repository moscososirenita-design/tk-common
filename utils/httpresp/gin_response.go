package httpresp

import "github.com/gin-gonic/gin"

// GinOK 输出 Gin 成功响应（HTTP 200）。
func GinOK(c *gin.Context, data interface{}) {
	// 调用c.JSON完成当前处理。
	c.JSON(200, Envelope{
		// 处理当前语句逻辑。
		Code: 0,
		// 处理当前语句逻辑。
		Msg: "ok",
		// 处理当前语句逻辑。
		Data: data,
	})
}

// GinError 输出 Gin 失败响应（HTTP 200，业务码透传）。
func GinError(c *gin.Context, code int, msg string) {
	// 调用c.JSON完成当前处理。
	c.JSON(200, Envelope{
		// 处理当前语句逻辑。
		Code: code,
		// 处理当前语句逻辑。
		Msg: msg,
	})
}

// GinFailWithStatus 输出 Gin 失败响应（可指定 HTTP 状态码）。
func GinFailWithStatus(c *gin.Context, statusCode int, code int, msg string) {
	// 调用c.JSON完成当前处理。
	c.JSON(statusCode, Envelope{
		// 处理当前语句逻辑。
		Code: code,
		// 处理当前语句逻辑。
		Msg: msg,
	})
}
