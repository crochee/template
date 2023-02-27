package middlewares

import (
	"errors"
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"

	"template/pkg/code"
	"template/pkg/logger"
	"template/pkg/msg/server"
	"template/pkg/resp"
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
				server.Error(ctx, err)
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
			server.Error(ctx, code.ErrInternalServerError.WithResult(extra))
			resp.Error(c, code.ErrInternalServerError.WithResult(extra))
		}
	}()
	c.Next()
}
