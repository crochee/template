// Package msg
package msg

import (
	"sync"
	"time"
)

type MetadataPool interface {
	Get() *Metadata
	Put(*Metadata)
}

type Metadata struct {
	TraceID     string    `json:"trace_id" binding:"required"`
	ServiceName string    `json:"service_name" binding:"required"`
	Locate      string    `json:"locate" binding:"required"`
	RequestID   string    `json:"request_id"`
	AccountID   string    `json:"account_id"`
	UserID      string    `json:"user_id"`
	Summary     string    `json:"summary" binding:"required"`
	Desc        string    `json:"desc"`
	ErrorTime   time.Time `json:"error_time" binding:"required"`
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
	metadata.RequestID = ""
	metadata.AccountID = ""
	metadata.UserID = ""
	metadata.Summary = ""
	metadata.Desc = ""
	metadata.ErrorTime = time.Time{}
	d.pool.Put(metadata)
}
