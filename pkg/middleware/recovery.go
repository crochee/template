package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/crochee/devt/pkg/code"
	"github.com/crochee/devt/pkg/logger"
	"github.com/crochee/devt/pkg/resp"
)

// Recovery panic logx
func Recovery(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			var brokenPipe bool
			if ne, ok := r.(*net.OpError); ok {
				var se *os.SyscallError
				if errors.As(ne.Err, &se) {
					brokenPipe = strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
						strings.Contains(strings.ToLower(se.Error()), "connection reset by peer")
				}
			}
			ctx := c.Request.Context()
			httpRequest, err := httputil.DumpRequest(c.Request, true)
			if err != nil {
				logger.From(ctx).Error(err.Error())
			}
			headers := strings.Split(string(httpRequest), "\r\n")
			for idx, header := range headers {
				current := strings.Split(header, ":")
				if current[0] == "Authorization" { // 数据脱敏
					headers[idx] = current[0] + ": *"
				}
			}
			headersToStr := strings.Join(headers, "\r\n")
			logger.From(ctx).Sugar().Errorf("[Recovery] %s\n%v\n%s",
				headersToStr, r, debug.Stack())
			extra := fmt.Sprint(r)
			if brokenPipe {
				extra = fmt.Sprintf("broken pipe or connection reset by peer;%v", r)
			}
			resp.Error(c, code.ErrInternalServerError.WithResult(extra))
		}
	}()
	c.Next()
}
