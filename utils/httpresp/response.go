package httpresp

import (
	"encoding/json"
	"net/http"

	"github.com/moscososirenita-design/tk-common/utils/codes"
)

// Envelope 统一响应结构。
type Envelope struct {
	// 处理当前语句逻辑。
	Code int `json:"code"`
	// 处理当前语句逻辑。
	Msg string `json:"msg"`
	// 处理当前语句逻辑。
	Data interface{} `json:"data,omitempty"`
}

// Write 按传入 HTTP 状态码输出统一响应结构。
func Write(w http.ResponseWriter, statusCode int, code int, msg string, data interface{}) {
	// 明确 JSON 响应头，避免浏览器/代理误判。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 调用w.WriteHeader完成当前处理。
	w.WriteHeader(statusCode)
	// 更新当前变量或字段值。
	_ = json.NewEncoder(w).Encode(Envelope{
		// 处理当前语句逻辑。
		Code: code,
		// 处理当前语句逻辑。
		Msg: msg,
		// 处理当前语句逻辑。
		Data: data,
	})
}

// OK 输出成功响应（HTTP 200）。
func OK(w http.ResponseWriter, data interface{}) {
	// 调用Write完成当前处理。
	Write(w, http.StatusOK, codes.OK, "ok", data)
}

// Fail 输出失败响应（支持自定义 HTTP 状态码）。
func Fail(w http.ResponseWriter, statusCode int, code int, msg string) {
	// 调用Write完成当前处理。
	Write(w, statusCode, code, msg, nil)
}

// BizFail 输出业务失败响应（HTTP 固定 200）。
func BizFail(w http.ResponseWriter, code int, msg string) {
	// 调用Write完成当前处理。
	Write(w, http.StatusOK, code, msg, nil)
}
