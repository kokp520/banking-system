package middleware

import (
	"github.com/kokp520/banking-system/server/pkg/trace"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/pkg/logger"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		traceID := trace.GetTraceID(c.Request.Context())

		logger.Info("request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("trace_id", traceID),
		)
	}
}
