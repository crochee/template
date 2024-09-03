package prometheus

import (
	"context"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"template/pkg/logger"
	"template/pkg/redis"
	"template/pkg/rpc/event"
)

// Service Status
const (
	StatusOK          = 0 // service is healthy
	StatusUnhealth    = 1 // self status is unhealthy
	StatusReliesError = 2 // the relies service is unhealthy
)

// Redis cluster status
const (
	RedisClusterStatusOK = "cluster_state:ok"
)

// Rabbitmq ping result
const (
	RabbitmqPingResult = "pong"
)

var redisStatusGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "redis_service_status",
		Help: "the status of redis cluster",
	},
)

var rabbitmqStatusGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "mq_service_status",
		Help: "the status of rabbitmq",
	},
)

func NewStatusMetricHandler() *StatusMetricHandler {
	register := prometheus.NewRegistry()
	register.MustRegister(redisStatusGauge)
	register.MustRegister(rabbitmqStatusGauge)
	return &StatusMetricHandler{
		promhttpHandler: promhttp.HandlerFor(
			register,
			promhttp.HandlerOpts{EnableOpenMetrics: true},
		),
		redis: redis.NewClient(),
	}
}

type StatusMetricHandler struct {
	promhttpHandler http.Handler
	redis           *redis.Client
}

func (s *StatusMetricHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.Handle(req.Context())
	s.promhttpHandler.ServeHTTP(rw, req)
}

func (s *StatusMetricHandler) Handle(ctx context.Context) {
	// Redis Status
	redisStatus, err := s.checkRedisStatus(ctx)
	if err != nil {
		logger.From(ctx).Error("failed to check redis status", zap.Error(err))
	}
	redisStatusGauge.Set(float64(redisStatus))

	// RabbitMQ status
	mqStatus, err := s.checkRabbitmqStatus(ctx)
	if err != nil {
		logger.From(ctx).Error("failed to check rabbitmq status", zap.Error(err))
	}
	rabbitmqStatusGauge.Set(float64(mqStatus))
}

func (s *StatusMetricHandler) checkRedisStatus(ctx context.Context) (int, error) {
	// Check redis connection
	if _, err := s.redis.Ping(); err != nil {
		return StatusUnhealth, err
	}

	// Check redis cluster status
	info, err := s.redis.ClusterInfo()
	if err != nil {
		return StatusUnhealth, err
	}
	if !strings.HasPrefix(info, RedisClusterStatusOK) {
		return StatusUnhealth, nil
	}

	return StatusOK, nil
}

func (s *StatusMetricHandler) checkRabbitmqStatus(ctx context.Context) (int, error) {
	delivery, err := event.Ping(ctx)
	if err != nil {
		return StatusUnhealth, err
	}
	body := string(delivery.Body)
	if body != RabbitmqPingResult {
		return StatusUnhealth, nil
	}
	return StatusOK, nil
}
