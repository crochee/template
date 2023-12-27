package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"template/pkg/async"
	"template/pkg/logger"
	"template/pkg/msg"
	"template/pkg/utils"
)

func TestRegisterAPI(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	p := async.NewMockProducer(ctl)
	p.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ async.Channel, exchange, routingKey string, param interface{}) error {
			logger.From(ctx).Sugar().Infof("ex:%s key:%s ,%s\n", exchange, routingKey, param.([]amqp.Publishing)[0].Body)
			return nil
		}).
		AnyTimes()

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	ctx := logger.With(context.Background(), logger.New())
	getTraceID := func(context.Context) string {
		return uuid.NewV4().String()
	}
	RegisterAPI(g, ctx, getTraceID)
	New(ctx, getTraceID, func(o *msg.WriterOption) {
		o.Publisher = p
	})
	t.Run("getProcessor", func(t *testing.T) {
		resp := utils.PerformRequest(g, "GET", "/v1/produce/processors", nil, nil)
		t.Log(resp.Code, resp.Body.String())
		assert.Equal(t, http.StatusOK, resp.Code)
	})
	t.Run("putProcessor", func(t *testing.T) {
		var w bytes.Buffer
		err := json.NewEncoder(&w).Encode(processorContent{
			MaxQueueSize:       1,
			BatchTimeout:       5 * time.Second,
			MaxExportBatchSize: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		resp := utils.PerformRequest(g, "PUT", "/v1/produce/processors", &w, nil)
		t.Log(resp.Code, resp.Body.String())
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})
	t.Run("putProcessorError", func(t *testing.T) {
		var w bytes.Buffer
		err := json.NewEncoder(&w).Encode(processorContent{
			MaxQueueSize:       1,
			BatchTimeout:       500 * time.Microsecond,
			MaxExportBatchSize: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		resp := utils.PerformRequest(g, "PUT", "/v1/produce/processors", &w, nil)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
