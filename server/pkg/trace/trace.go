package trace

import (
	"context"
)

const Key = "trace_id"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, Key, traceID)
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(Key).(string); ok {
		return traceID
	}
	return ""
}
