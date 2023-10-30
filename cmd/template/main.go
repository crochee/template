package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"template/config"
	"template/internal/code"
	"template/internal/gateway"
	"template/internal/router"
	"template/internal/store/mysql"
	"template/internal/util/v"
	"template/pkg/conc/pool"
	"template/pkg/json/extension"
	"template/pkg/logger"
	"template/pkg/validator"
)

var configFile = flag.String("f", "./config/template.yaml", "the config file")

func main() {
	flag.Parse()
	// 初始化配置
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatal(err)
	}
	if err := code.Loading(); err != nil {
		log.Fatal(err)
	}
	if mode := strings.ToLower(viper.GetString("mode")); mode != "" {
		gin.SetMode(mode)
	}
	if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func run() error {
	ctx := logger.With(context.Background(),
		logger.New(
			logger.WithServerName(viper.GetString("service.name")),
			logger.WithLevel(viper.GetString("log.level")),
			logger.WithWriter(logger.SetWriter(viper.GetBool("log.console"), ""))),
	)
	// 初始化数据库
	dataStore, err := mysql.NewMysqlClient(ctx)
	if err != nil {
		return err
	}
	defer dataStore.DB.Close()
	// 后台调用注册实现
	client := gateway.NewBaseClient()

	g := pool.New().WithContext(ctx).WithCancelOnError()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.New(dataStore, client),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	logger.From(ctx).Sugar().Debugf("listen on %s", srv.Addr)
	// 服务启动流程
	g.Go(func(ctx context.Context) error {
		return startAction(ctx, srv)
	})
	// 服务关闭流程
	g.Go(func(ctx context.Context) error {
		return shutdownAction(ctx, srv)
	})

	defer logger.From(ctx).Sync()

	if err = g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.From(ctx).Error("server run with error", zap.Error(err))
		return err
	}
	return nil
}

func startAction(ctx context.Context, srv *http.Server) error {
	extension.Register()
	var err error
	if binding.Validator, err = validator.New(); err != nil {
		return err
	}
	if err = validator.RegisterValidation(binding.Validator, "order", validator.OrderWithDBSort); err != nil {
		return err
	}
	logger.From(ctx).Sugar().Infof("%s run on %s", v.ServiceName, gin.Mode())
	return srv.ListenAndServe()
}

const DefaultStopTime = 15 * time.Second

func shutdownAction(ctx context.Context, srv *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-quit:
	}
	newCtx, cancel := context.WithTimeout(ctx, DefaultStopTime)
	defer cancel()
	logger.From(ctx).Info("shutting down server...")
	return multierr.Append(err, srv.Shutdown(newCtx))
}
