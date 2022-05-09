// Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
// Description:
// Author: l30002214
// Create: 2020/12/8

// Package middleware
package middleware

import (
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/crochee/lirity/e"
	"github.com/crochee/lirity/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
			e.Code(c, e.ErrInternalServerError.WithResult(extra))
		}
	}()
	c.Next()
}
