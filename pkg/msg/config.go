// Package msg
package msg

import (
	"sync/atomic"
)

type CfgHandler interface {
	Status() *switchStatus
	Enable() bool
	SetEnable(bool)
	Exchange() string
	SetExchange(string)
	RoutingKey() string
	SetRoutingKey(string)
	URI() string
	SetURI(string)
	Topic() string
	SetTopic(string)
}

func NewCfgHandler() CfgHandler {
	return &netCfg{
		enableStatus: NewSwitchStatus(),
	}
}

type netCfg struct {
	enableStatus *switchStatus
	exchange     atomic.Value
	routingKey   atomic.Value
	uri          atomic.Value
	topic        atomic.Value
}

func (n *netCfg) Status() *switchStatus {
	return n.enableStatus
}

func (n *netCfg) Enable() bool {
	return n.enableStatus.Output()
}

func (n *netCfg) SetEnable(v bool) {
	n.enableStatus.Input(v)
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

func (n *netCfg) Topic() string {
	v, ok := n.topic.Load().(string)
	if !ok {
		return ""
	}
	return v
}

func (n *netCfg) SetTopic(s string) {
	n.topic.Store(s)
}
