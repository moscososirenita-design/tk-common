package utils

import "strings"

// SafeTrim 对字符串做空白清理；nil/空值场景调用方可直接复用。
func SafeTrim(s string) string {
	// 返回当前处理结果。
	return strings.TrimSpace(s)
}

// IsBlank 判断字符串是否为空白。
func IsBlank(s string) bool {
	// 返回当前处理结果。
	return SafeTrim(s) == ""
}
