package utils

import (
	"context"
	"sync"
)

type ctxKey string

const TraceKey ctxKey = "traceId"

var once sync.Once

func NewContext() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, TraceKey, RandomString(8))
}

func AddTraceId(ctx context.Context) context.Context {
	if ctx.Value(TraceKey) != nil {
		return ctx
	}
	return WithTraceId(ctx, RandomString(8))
}

func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, TraceKey, traceId)
}

func GetTraceId(ctx context.Context) string {
	if traceId := ctx.Value(TraceKey); traceId != nil {
		return traceId.(string)
	}
	return "-"
}

var initContent context.Context

func GetInitContext() context.Context {
	once.Do(func() {
		initContent = NewContext()
	})
	return initContent
}
