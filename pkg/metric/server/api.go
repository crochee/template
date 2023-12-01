package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"template/pkg/code"
	"template/pkg/logger"
	"template/pkg/metric"
	"template/pkg/metric/model"
	"template/pkg/ptf"
	"template/pkg/resp"
)

var defaultFilterDay = []int{1, 2, 3, 4, 5, 6, 7}

func RegisterAPICheck(router *gin.Engine, service_name string) {
	router.GET("/metric/query", listMetricQuery)
	router.GET("/metric/stat", listMetric)
	router.GET("/metric/details", listMetricDetail(service_name))
}

func listMetric(c *gin.Context) {
	var (
		ctx    = c.Request.Context()
		filter metric.Filter
	)

	days := c.QueryArray("day")
	labels := c.QueryArray("label")

	filter.Labels = make([]*model.Label, 0, len(labels))
	for _, label := range labels {
		tmp := strings.Split(label, ",")
		if len(tmp) != 2 {
			logger.From(ctx).Error("label format failed")
			resp.Error(c, code.ErrInvalidParam)
			return
		}
		filter.Labels = append(filter.Labels, &model.Label{Name: tmp[0], Value: tmp[1]})
	}
	filter.Days = make([]int, 0, len(days))
	for _, dayStr := range days {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			logger.From(ctx).Error("day convert failed", zap.Error(err))
			resp.ErrorParam(c, err)
			return
		}
		filter.Days = append(filter.Days, day)
	}
	if len(filter.Days) == 0 {
		filter.Days = defaultFilterDay
	}

	result := metric.Monitor.Metrics(filter)
	c.JSON(http.StatusOK, struct {
		List interface{} `json:"list"`
	}{
		List: result,
	})
}

type ListMetricParam struct {
	// 标签,可传递多个标签值,单个标签格式为：标签名 + 逗号 + 标签值
	Labels []string `json:"label" form:"label"`
	// 查询哪天的数据,可传递多个day值, day取值范围,{-1,1,2,3,4,5,6,7},其中-1表示all
	Days []string `json:"day" form:"day"`
	// 查询显示数量
	Number int `json:"number" form:"number,default=20"`
	// 排序字段, 默认以最大时延排序
	Type string `json:"type" form:"type,default=max" binding:"omitempty,oneof=max min average count"`
	// 是否将结果以表格形式返回，表格返回-true, json返回-false, 默认返回json
	Table bool `json:"table" form:"table"`
}

func listMetricDetail(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			ctx    = c.Request.Context()
			param  ListMetricParam
			filter metric.Filter
		)

		if err := c.ShouldBindQuery(&param); err != nil {
			logger.From(ctx).Error("bind query error", zap.Error(err))
			resp.ErrorParam(c, err)
			return
		}
		filter.Number = param.Number
		filter.Type = param.Type

		filter.Labels = make([]*model.Label, 0, len(param.Labels))
		for _, label := range param.Labels {
			tmp := strings.Split(label, ",")
			if len(tmp) != 2 {
				logger.From(ctx).Error("label format failed")
				resp.Error(c, code.ErrInvalidParam)
				return
			}
			filter.Labels = append(filter.Labels, &model.Label{Name: tmp[0], Value: tmp[1]})
		}
		filter.Days = make([]int, 0, len(param.Days))
		for _, dayStr := range param.Days {
			day, err := strconv.Atoi(dayStr)
			if err != nil {
				logger.From(ctx).Error("day convert failed", zap.Error(err))
				resp.ErrorParam(c, err)
				return
			}
			filter.Days = append(filter.Days, day)
		}
		if len(filter.Days) == 0 {
			filter.Days = defaultFilterDay
		}

		response := metric.MetricsSortTable{
			List: metric.Monitor.MetricsSort(filter),
		}
		if param.Table {
			tableHeader, sheet := metric.GenerateTableAttr(name, filter)
			resp.SuccessWithFile(c, response, sheet, ptf.Headers(tableHeader))
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

type ListMetricQueryConditionResult struct {
	MaxDay int            `json:"max_day"`
	Labels []*model.Label `json:"labels"`
}

func listMetricQuery(c *gin.Context) {
	var (
		result                 ListMetricQueryConditionResult
		maxLatency, minLatency uint64
		err                    error
	)

	reqLatency := c.Query("latency")
	if strings.Compare(reqLatency, "") != 0 {
		if strings.Contains(reqLatency, ">") {
			if maxLatency, err = strconv.ParseUint(reqLatency[1:], 10, 64); err != nil {
				logger.From(c.Request.Context()).Error("latency convert failed", zap.Error(err), zap.String("latency", reqLatency))
				resp.ErrorParam(c, err)
				return
			}
		} else if strings.Contains(reqLatency, "<") {
			if minLatency, err = strconv.ParseUint(reqLatency[1:], 10, 64); err != nil {
				logger.From(c.Request.Context()).Error("latency convert failed", zap.Error(err), zap.String("latency", reqLatency))
				resp.ErrorParam(c, err)
				return
			}
		} else {
			logger.From(c.Request.Context()).Error("latency format failed", zap.String("latency", reqLatency))
			resp.Error(c, code.ErrInvalidParam)
			return
		}
	}

	result.MaxDay = metric.Monitor.GetMaxDay()
	result.Labels = metric.Monitor.GetLabels(maxLatency, minLatency)

	c.JSON(http.StatusOK, result)
}
