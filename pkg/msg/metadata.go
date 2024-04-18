// Package msg
package msg

import (
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

type MetadataPool interface {
	Get() *Metadata
	Put(*Metadata)
}

type Metadata struct {
	TraceID      string    `json:"trace_id"       binding:"required"`
	ServiceName  string    `json:"service_name"   binding:"required"`
	Locate       string    `json:"locate"         binding:"required"`
	SpanID       string    `json:"span_id"`
	ParentSpanID string    `json:"parent_span_id"`
	AccountID    string    `json:"account_id"`
	UserID       string    `json:"user_id"`
	ResID        string    `json:"res_id"`
	ResType      string    `json:"res_type"`
	SubResID     string    `json:"sub_res_id"`
	SubResType   string    `json:"sub_res_type"`
	Summary      string    `json:"summary"        binding:"required"`
	Desc         string    `json:"desc"`
	ErrorTime    time.Time `json:"error_time"     binding:"required"`
}

func NewMetadataPool() MetadataPool {
	return &defaultMetadataPool{pool: sync.Pool{New: func() interface{} {
		return new(Metadata)
	}}}
}

type defaultMetadataPool struct {
	pool sync.Pool
}

func (d *defaultMetadataPool) Get() *Metadata {
	v, ok := d.pool.Get().(*Metadata)
	if !ok {
		return new(Metadata)
	}
	return v
}

func (d *defaultMetadataPool) Put(metadata *Metadata) {
	metadata.TraceID = ""
	metadata.ServiceName = ""
	metadata.Locate = ""
	metadata.SpanID = ""
	metadata.ParentSpanID = ""
	metadata.AccountID = ""
	metadata.UserID = ""
	metadata.ResID = ""
	metadata.ResType = ""
	metadata.SubResID = ""
	metadata.SubResType = ""
	metadata.Summary = ""
	metadata.Desc = ""
	metadata.ErrorTime = time.Time{}
	d.pool.Put(metadata)
}

var (
	ResIDKey      = attribute.Key("ResID")
	ResTypeKey    = attribute.Key("ResType")
	SubResIDKey   = attribute.Key("SubResID")
	SubResTypeKey = attribute.Key("SubResType")
	AccountIDKey  = attribute.Key("AccountID")
	UserIDKey     = attribute.Key("UserID")
	LocateKey     = attribute.Key("Locate")
	MsgKey        = attribute.Key("msg")
	KeepKey       = attribute.Key("keep")
)

var (
	CurlEvent    = "curl"
	SlowSQLEvent = "slow_sql"
)
