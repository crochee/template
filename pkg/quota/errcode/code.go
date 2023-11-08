package errcode

import (
	"template/pkg/code"
)

var (
	// 配额相关错误
	ErrCodeResourceQuotaInsufficient     = code.Froze("500-11004400", "资源配额不足")
	ErrCodeResourceQuotaInvalid          = code.Froze("500-11004401", "资源配额数据无效，请重试")
	ErrCodeResourceQuotaAbnormal         = code.Froze("500-11004402", "资源配额数据异常，需要重新刷新用户配额数据")
	ErrCodeResourceQuotaScriptResInvalid = code.Froze("500-11004403", "资源配额脚本执行结果无效")
	ErrCodeQuotaServerDisable            = code.Froze("500-11004404", "配额服务暂时不可用，请稍后再试")
	ErrCodeWaitLockTimeout               = code.Froze("500-11004405", "等待锁超时")
)
