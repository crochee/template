package routine

import "context"

type option struct {
	recoverFunc func(ctx context.Context, r interface{})
	copyContext func(dst context.Context, src context.Context) context.Context
}

type Option func(*option)

// Recover register to Pool
func Recover(f func(context.Context, interface{})) Option {
	return func(o *option) { o.recoverFunc = f }
}

func CopyContext(f func(dst context.Context, src context.Context) context.Context) Option {
	return func(o *option) { o.copyContext = f }
}
