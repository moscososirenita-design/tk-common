package reqx

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// ParseIntOrDefault 解析正整数；非法值时回退默认值。
func ParseIntOrDefault(raw string, fallback int) int {
	// 定义并初始化当前变量。
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	// 判断条件并进入对应分支逻辑。
	if err != nil || v <= 0 {
		// 返回当前处理结果。
		return fallback
	}
	// 返回当前处理结果。
	return v
}

// ParsePathID 从路径里按 prefix 读取下一个段作为正整数 ID。
func ParsePathID(path string, prefix string) (uint64, error) {
	// 定义并初始化当前变量。
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// 循环处理当前数据集合。
	for idx := range parts {
		// 判断条件并进入对应分支逻辑。
		if parts[idx] == prefix && idx+1 < len(parts) {
			// 定义并初始化当前变量。
			id, err := strconv.ParseUint(parts[idx+1], 10, 64)
			// 判断条件并进入对应分支逻辑。
			if err != nil || id == 0 {
				// 返回当前处理结果。
				return 0, fmt.Errorf("invalid id")
			}
			// 返回当前处理结果。
			return id, nil
		}
	}
	// 返回当前处理结果。
	return 0, fmt.Errorf("invalid id")
}

// DeviceID 从请求中提取设备标识：优先 Header，再回退 Query。
func DeviceID(r *http.Request) string {
	// 定义并初始化当前变量。
	deviceID := strings.TrimSpace(r.Header.Get("X-Device-ID"))
	// 判断条件并进入对应分支逻辑。
	if deviceID != "" {
		// 返回当前处理结果。
		return deviceID
	}
	// 返回当前处理结果。
	return strings.TrimSpace(r.URL.Query().Get("device_id"))
}

// ClientIP 从代理头中提取客户端 IP，最后回退 RemoteAddr。
func ClientIP(r *http.Request) string {
	// 判断条件并进入对应分支逻辑。
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		// 定义并初始化当前变量。
		parts := strings.Split(xff, ",")
		// 判断条件并进入对应分支逻辑。
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			// 返回当前处理结果。
			return strings.TrimSpace(parts[0])
		}
	}
	// 判断条件并进入对应分支逻辑。
	if rip := strings.TrimSpace(r.Header.Get("X-Real-IP")); rip != "" {
		// 返回当前处理结果。
		return rip
	}
	// 定义并初始化当前变量。
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 返回当前处理结果。
		return host
	}
	// 返回当前处理结果。
	return strings.TrimSpace(r.RemoteAddr)
}
