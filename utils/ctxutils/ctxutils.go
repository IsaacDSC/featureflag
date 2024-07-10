package ctxutils

import (
	"context"
)

func SetContext(ctx context.Context, key string, value any) context.Context {
	ctx = context.WithValue(ctx, key, value)
	return ctx
}

func GetValueCtx(ctx context.Context, key string) any {
	value := ctx.Value(key)
	return value
}
