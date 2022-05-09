package code

import "github.com/crochee/lirity/e"

var (
	// 000~099 系统类

	ErrNoAccount = e.Froze("4001100000", "用户不存在")
	ErrNoUpdate  = e.Froze("3041100001", "数据无更新")
)

func Loading() error {
	return e.AddCode(map[e.ErrorCode]struct{}{
		ErrNoAccount: {},
		ErrNoUpdate:  {},
	})
}
