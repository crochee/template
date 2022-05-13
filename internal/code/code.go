package code

import "github.com/crochee/devt/pkg/code"

var (
	// 000~099 系统类

	ErrNoAccount = code.Froze("4001100000", "用户不存在")
	ErrNoUpdate  = code.Froze("3041100001", "数据无更新")
)

func Loading() error {
	return code.AddCode(map[code.ErrorCode]struct{}{
		ErrNoAccount: {},
		ErrNoUpdate:  {},
	})
}
