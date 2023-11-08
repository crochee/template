package ctxw

import (
	"context"
)

type pluginEnable struct{}

// SetPluginEnable 设置数据库同步的插件是否开启
func SetPluginEnable(ctx context.Context, enable bool) context.Context {
	return context.WithValue(ctx, pluginEnable{}, enable)
}

// GetPluginEnable 获取数据库同步的插件是否开启,默认情况时开启
func GetPluginEnable(ctx context.Context) bool {
	flag, ok := ctx.Value(pluginEnable{}).(bool)
	if !ok {
		return true
	}
	return flag
}
