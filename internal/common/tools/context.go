package tools

import (
	"context"

	"github.com/google/uuid"
)

// 上下文设置

type contextKey string

const TraceIDKey contextKey = "trace_id"

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, TraceIDKey, id)
}

func GetTraceID(ctx context.Context) string {
	if val := ctx.Value(TraceIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return uuid.NewString()
}