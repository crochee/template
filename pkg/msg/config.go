// Package msg
package msg

import (
	"sync/atomic"
	"time"
)

type CfgHandler interface {
	LogicalCfg
	NetCfg
}

type LogicalCfg interface {
	SetLimit(uint64)
	Limit() uint64
	SetTimeout(time.Duration)
	Timeout() time.Duration
	SetEnable(bool)
	Enable() bool
	OUtEnable() *switchStatus
}

type NetCfg interface {
	Queue() string
	SetQueue(string)
	Exchange() string
	SetExchange(string)
	RoutingKey() string
	SetRoutingKey(string)
	URI() string
	SetURI(string)
}

func NewCfgHandler() CfgHandler {
	s := struct {
		LogicalCfg
		NetCfg
	}{
		&standardCfg{enableStatus: NewSwitchStatus()},
		&netCfg{},
	}
	s.SetLimit(10)
	s.SetTimeout(5 * time.Second)
	return s
}

type standardCfg struct {
	enableStatus *switchStatus
	limit        uint64
	timeout      atomic.Value
}

func (s *standardCfg) SetLimit(limit uint64) {
	atomic.StoreUint64(&s.limit, limit)
}

func (s *standardCfg) Limit() uint64 {
	return atomic.LoadUint64(&s.limit)
}

func (s *standardCfg) SetTimeout(timeout time.Duration) {
	s.timeout.Store(timeout)
}

func (s *standardCfg) Timeout() time.Duration {
	v, ok := s.timeout.Load().(time.Duration)
	if !ok {
		return 0
	}
	return v
}

func (s *standardCfg) SetEnable(enable bool) {
	s.enableStatus.Input(enable)
}

func (s *standardCfg) Enable() bool {
	return s.enableStatus.Output()
}

func (s *standardCfg) OUtEnable() *switchStatus {
	return s.enableStatus
}

type netCfg struct {
	queueName  atomic.Value
	exchange   atomic.Value
	routingKey atomic.Value
	uri        atomic.Value
}

func (n *netCfg) Queue() string {
	v, ok := n.queueName.Load().(string)
	if !ok {
		return ""
	}
	return v
}

func (n *netCfg) SetQueue(topic string) {
	n.queueName.Store(topic)
}

func (n *netCfg) Exchange() string {
	v, ok := n.exchange.Load().(string)
	if !ok {
		return ""
	}
	return v
}

func (n *netCfg) SetExchange(s string) {
	n.exchange.Store(s)
}

func (n *netCfg) RoutingKey() string {
	v, ok := n.routingKey.Load().(string)
	if !ok {
		return ""
	}
	return v
}

func (n *netCfg) SetRoutingKey(s string) {
	n.routingKey.Store(s)
}

func (n *netCfg) URI() string {
	v, ok := n.uri.Load().(string)
	if !ok {
		return ""
	}
	return v
}

func (n *netCfg) SetURI(s string) {
	n.uri.Store(s)
}
