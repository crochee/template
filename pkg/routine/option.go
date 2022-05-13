package routine

import "context"

type option struct {
	recoverFunc func(ctx context.Context, r interface{})
}

// Recover register to Pool
func Recover(f func(context.Context, interface{})) func(*option) {
	return func(o *option) { o.recoverFunc = f }
}
