package rpc

import (
	"context"
	"time"

	amqprpc "github.com/0x4b53/amqp-rpc/v3"
	amqp "github.com/rabbitmq/amqp091-go"

	"template/pkg/logger"
	"template/pkg/utils/v"
)

type Option func(*option)

func WithGetAccountID(getAccountID func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getAccountID = getAccountID
	}
}

func WithGetAccountName(getAccountName func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getAccountName = getAccountName
	}
}

func WithGetUserID(getUserID func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getUserID = getUserID
	}
}

func WithGetTraceID(getTraceID func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getTraceID = getTraceID
	}
}

func WithGetIP(getIP func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getIP = getIP
	}
}

func WithGetOperatorID(getOperatorID func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getOperatorID = getOperatorID
	}
}

func WithGetOperatorName(getOperatorName func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getOperatorName = getOperatorName
	}
}

func WithGetOperatorType(getOperatorType func(ctx context.Context) string) Option {
	return func(o *option) {
		o.getOperatorType = getOperatorType
	}
}

func WithPushTimeout(pushTimeout time.Duration) Option {
	return func(o *option) {
		o.pushTimeout = pushTimeout
	}
}

type option struct {
	getAccountID    func(ctx context.Context) string
	getAccountName  func(ctx context.Context) string
	getUserID       func(ctx context.Context) string
	getTraceID      func(ctx context.Context) string
	getIP           func(ctx context.Context) string
	getOperatorID   func(ctx context.Context) string
	getOperatorName func(ctx context.Context) string
	getOperatorType func(ctx context.Context) string
	pushTimeout     time.Duration
}

type RpcClient struct {
	client *amqprpc.Client
	option
}

func NewRpcClient(ctx context.Context, uri string, opts ...Option) *RpcClient {
	opt := option{pushTimeout: 5 * time.Second}
	for _, f := range opts {
		f(&opt)
	}
	return &RpcClient{
		client: amqprpc.NewClient(uri).
			WithErrorLogger(logger.From(ctx).Sugar().Errorf).
			WithDebugLogger(logger.From(ctx).Sugar().Debugf).
			WithTimeout(opt.pushTimeout).
			WithPublishSettings(amqprpc.PublishSettings{
				Mandatory:   true,
				Immediate:   false,
				ConfirmMode: false,
			}),
		option: opt,
	}
}

func (c *RpcClient) send(ctx context.Context, req *amqprpc.Request) (*amqp.Delivery, error) {
	// Set the common message header
	req.WriteHeader(v.HeaderAccountID, c.getAccountID(ctx))
	req.WriteHeader(v.HeaderAccountName, c.getAccountName(ctx))
	req.WriteHeader(v.HeaderUserID, c.getUserID(ctx))
	req.WriteHeader(v.HeaderTraceID, c.getTraceID(ctx))
	req.WriteHeader(v.HeaderRealIP, c.getIP(ctx))
	req.WriteHeader(v.HeaderOperatorID, c.getOperatorID(ctx))
	req.WriteHeader(v.HeaderOperatorName, c.getOperatorName(ctx))
	req.WriteHeader(v.HeaderOperatorType, c.getOperatorType(ctx))

	return c.client.Send(req)
}

// Call send a message to rpc server and wait for the response from the server side.
func (c *RpcClient) Call(ctx context.Context, req *amqprpc.Request) (*amqp.Delivery, error) {
	req.WithResponse(true)
	return c.send(ctx, req)
}

// Cast send a message to rpc server and don't care about the response.
func (c *RpcClient) Cast(ctx context.Context, req *amqprpc.Request) error {
	req.WithResponse(false)
	_, err := c.send(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
