package tools

import (
	"context"

	"github.com/google/uuid"
)

// 上下文设置

type contextKey string

const TraceIDKey contextKey = "trace_id"

// WithTraceID 设置上下文的 TraceID
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, TraceIDKey, id)
}

// GetTraceID 从上下文获取 TraceID
func GetTraceID(ctx context.Context) string {
	if ctx == nil || ctx == context.TODO() { // 上下文为空时不设置trace_id
		return ""
	}
	if val := ctx.Value(TraceIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return uuid.NewString()
}