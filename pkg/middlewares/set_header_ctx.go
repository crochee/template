package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"template/pkg/utils/v"
)

// SetHeaderContext set some public parameters to header and context
func SetHeaderContext(setCtx func(*gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 请求头X-Trace-ID不能为空，为空时需要自动生成
		traceID := c.GetHeader(v.HeaderTraceID)
		if traceID == "" {
			traceID = uuid.NewV4().String()
		} else {
			// 检测trace_id是否时uuid,若不是则根据规则生成新的traceID
			_, err := uuid.FromString(traceID)
			if err != nil {
				traceID = GenerateNewTraceID(traceID)
			}
		}
		c.Request.Header.Set(v.HeaderTraceID, traceID) // 请求头

		// 设置响应头
		if c.Writer.Header().Get(v.HeaderTraceID) == "" {
			c.Writer.Header().Set(v.HeaderTraceID, traceID) // 响应头
		}
		setCtx(c)
		c.Next()
	}
}

// GenerateNewTraceID 生成合法的trace_id, 有以下规则
// 1，去除第一个-前面的内容（包括-），如果剩余部分为合法UUID，则使用; 如果不包含-，则重新生成一个新的UUID
// 2，如果都不合法，则重新生成一个新的UUID
func GenerateNewTraceID(illegalTraceID string) string {
	// 获取第一个横杠的位置
	index := strings.Index(illegalTraceID, "-")
	if index == -1 {
		return uuid.NewV4().String()
	}
	newTraceID := illegalTraceID[index+1:]
	// 判断是否合法, 不合法则重新生成一个新的UUID
	_, err := uuid.FromString(newTraceID)
	if err != nil {
		return uuid.NewV4().String()
	}
	return newTraceID
}
