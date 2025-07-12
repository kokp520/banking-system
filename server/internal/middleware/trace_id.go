package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kokp520/banking-system/server/pkg/trace"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// put in gin context for gin http
		c.Set(trace.Key, traceID)

		// put in context for service
		ctx := trace.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
