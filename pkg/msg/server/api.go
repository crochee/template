package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"template/pkg/logger/gormx"
	"template/pkg/msg"
	"template/pkg/resp"
)

func RegisterAPI(router *gin.Engine, ctx context.Context, getTraceID func(context.Context) string) {
	router.GET("/v1/produce/config", getProduceConfig)
	router.PATCH("/v1/produce/config", updateProduceConfig)
	router.GET("/v1/produce/processors", getProcessor)
	router.PUT("/v1/produce/processors", putProcessor(ctx, getTraceID))
}

// swagger:route GET /v1/produce/config 故障定位服务-内部运维使用 SwagNullRequest
// 查询生产者配置.
//
// This will get producer config info.
//
//	Responses:
//	  200: SGetProduceConfigRes
//	  default: ResponseCode
func getProduceConfig(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := GetProduce(ctx)
	if err != nil {
		Error(ctx, err)
		resp.ErrorParam(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// swagger:route PATCH /v1/produce/config 故障定位服务-内部运维使用 SUpdateProduceConfigRequest
// 更新生产者配置.
//
// This will update producer config info.
//
//	Responses:
//	  200: ResponseCode
//	  default: ResponseCode
func updateProduceConfig(c *gin.Context) {
	ctx := c.Request.Context()
	param := ProduceConfig{}
	if err := c.BindJSON(&param); err != nil {
		Error(ctx, err)
		resp.ErrorParam(c, err)
		return
	}
	if err := UpdateProduce(ctx, &param); err != nil {
		Error(ctx, err)
		resp.ErrorParam(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type processorContent struct {
	MaxQueueSize       int           `json:"max_queue_size" binding:"required,min=1"`
	BatchTimeout       time.Duration `json:"batch_timeout" binding:"required,min=1s"`
	MaxExportBatchSize int           `json:"max_export_batch_size" binding:"required,min=1"`
}

func getProcessor(c *gin.Context) {
	processorRWMutex.RLock()

	if processor == nil {
		processorRWMutex.RUnlock()
		c.Status(http.StatusNoContent)
		return
	}
	result := struct {
		processorContent
		CurrentLenght int `json:"current_lenght"`
	}{
		processorContent: processorContent{
			MaxQueueSize:       processor.GetOptions().MaxQueueSize,
			BatchTimeout:       processor.GetOptions().BatchTimeout,
			MaxExportBatchSize: processor.GetOptions().MaxExportBatchSize,
		},
		CurrentLenght: processor.Len(),
	}
	processorRWMutex.RUnlock()

	c.JSON(http.StatusOK, result)
}

func putProcessor(ctx context.Context, getTraceID func(context.Context) string) func(c *gin.Context) {
	return func(c *gin.Context) {
		if exp == nil {
			c.Status(http.StatusNoContent)
			return
		}
		var param processorContent
		if err := c.BindJSON(&param); err != nil {
			Error(ctx, err)
			resp.ErrorParam(c, err)
			return
		}
		processorRWMutex.Lock()

		if processor != nil {
			processor.Shutdown(ctx)
		}
		processor = msg.NewBatchSpanProcessor(
			exp,
			msg.WithMaxQueueSize(param.MaxQueueSize),
			msg.WithBatchTimeout(param.BatchTimeout),
			msg.WithMaxExportBatchSize(param.MaxExportBatchSize),
			msg.WithLoggerFrom(func(context.Context) gormx.Logger {
				return gormx.NewZapGormWriterFrom(ctx)
			}),
		)
		tpOpts := []sdktrace.TracerProviderOption{
			sdktrace.WithSpanProcessor(processor),
			sdktrace.WithIDGenerator(msg.DefaultIDGenerator(getTraceID)),
		}
		tp := sdktrace.NewTracerProvider(
			tpOpts...,
		)
		otel.SetTracerProvider(tp)

		processorRWMutex.Unlock()

		c.Status(http.StatusNoContent)
	}
}
