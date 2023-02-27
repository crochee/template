package middlewares

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"template/pkg/code"
	"template/pkg/metric"
	"template/pkg/metric/model"
	"template/pkg/resp"
)

// Metric 埋点和收集器处理中间件
func Metric(metrics chan<- metric.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			labels []*model.Label
			req    ActionBody
		)
		start := time.Now()

		path := c.FullPath()
		if strings.Contains(path, "action") {
			if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
				resp.Error(c, code.ErrInvalidParam.WithMessage(err.Error()))
				return
			}
			path = path + "/" + req.Action
		}

		labels = []*model.Label{
			{
				Name:  "path",
				Value: c.Request.Method + "-" + path,
			},
		}

		c.Next()
		// 请求结束之后信息收集
		end := time.Now()
		d := end.Sub(start) / time.Millisecond

		tmpMaxGauge := metric.MaxGauge.With(labels)
		tmpMaxGauge.Set(uint64(d))
		metric.Push(tmpMaxGauge, metrics)

		tmpMinGauge := metric.MinGauge.With(labels)
		tmpMinGauge.Set(uint64(d))
		metric.Push(tmpMinGauge, metrics)

		tempCounterLatency := metric.CounterLatency.With(labels)
		tempCounterLatency.Set(uint64(d))
		metric.Push(tempCounterLatency, metrics)

		temp := metric.CounterRequestCount.With(labels)
		temp.Inc()
		metric.Push(temp, metrics)
	}
}

type ActionBody struct {
	Action string `json:"action"`
}
