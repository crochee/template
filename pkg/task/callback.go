package task

import "context"

type Callback interface {
	Trigger(ctx context.Context, info StoreInfo, input interface{}, err error)
}

type CallbackFunc func(ctx context.Context, info StoreInfo, input interface{}, err error)

func (d CallbackFunc) Trigger(ctx context.Context, info StoreInfo, input interface{}, err error) {
	d(ctx, info, input, err)
}
