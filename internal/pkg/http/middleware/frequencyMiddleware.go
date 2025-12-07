package middleware

// HTTP请求频率限制中间件

import (
	"time"

	_ "github.com/gin-gonic/gin"
)


type FrequencyMiddleware struct {
	// 最大请求次数
	MaxRequests int
	// 时间窗口
	WindowTime time.Duration
}
