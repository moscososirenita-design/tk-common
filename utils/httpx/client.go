package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// defaultTransport 全局共享的 HTTP Transport，复用 TCP 连接池，
// 避免每次创建 Client 时新建连接导致 TIME_WAIT 堆积。
var defaultTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,
}

// NewTimeoutClient 创建带超时的 HTTP 客户端。
func NewTimeoutClient(timeout time.Duration) *http.Client {
	// 判断条件并进入对应分支逻辑。
	if timeout <= 0 {
		// 更新当前变量或字段值。
		timeout = 3 * time.Second
	}
	// 返回当前处理结果。
	return &http.Client{Timeout: timeout, Transport: defaultTransport}
}

// GetRange 发起带 Range 头的 GET 请求并读取有限字节。
func GetRange(ctx context.Context, client *http.Client, url string, rangeHeader string, maxRead int64) (statusCode int, contentType string, body []byte, err error) {
	// 判断条件并进入对应分支逻辑。
	if client == nil {
		// 更新当前变量或字段值。
		client = NewTimeoutClient(3 * time.Second)
	}
	// 定义并初始化当前变量。
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return 0, "", nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if rangeHeader != "" {
		// 调用req.Header.Set完成当前处理。
		req.Header.Set("Range", rangeHeader)
	}
	// 定义并初始化当前变量。
	resp, err := client.Do(req)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return 0, "", nil, err
	}
	// 注册延迟执行逻辑。
	defer resp.Body.Close()

	// 判断条件并进入对应分支逻辑。
	if maxRead <= 0 {
		// 更新当前变量或字段值。
		maxRead = 8192
	}
	// 更新当前变量或字段值。
	body, err = io.ReadAll(io.LimitReader(resp.Body, maxRead))
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return resp.StatusCode, resp.Header.Get("Content-Type"), nil, fmt.Errorf("read body failed: %w", err)
	}
	// 返回当前处理结果。
	return resp.StatusCode, resp.Header.Get("Content-Type"), body, nil
}
