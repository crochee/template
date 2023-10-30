package async

import "sync"

var (
	Pool = NewParamPool()
	Get  = Pool.Get
	Put  = Pool.Put
)

type ParamPool interface {
	Get() *Param
	Put(*Param)
}

type Param struct {
	TaskType string                 `json:"task_type" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
	Data     []byte                 `json:"data"`
}

func NewParamPool() ParamPool {
	return &defaultParamPool{pool: sync.Pool{New: func() interface{} {
		return &Param{
			TaskType: "",
			Metadata: make(map[string]interface{}),
			Data:     make([]byte, 0),
		}
	}}}
}

type defaultParamPool struct {
	pool sync.Pool
}

func (d *defaultParamPool) Get() *Param {
	v, ok := d.pool.Get().(*Param)
	if !ok {
		return &Param{
			TaskType: "",
			Metadata: make(map[string]interface{}),
			Data:     make([]byte, 0, 10),
		}
	}
	return v
}

func (d *defaultParamPool) Put(param *Param) {
	param.TaskType = ""
	if len(param.Metadata) > 4096 {
		param.Metadata = make(map[string]interface{})
	} else {
		for key := range param.Metadata {
			delete(param.Metadata, key)
		}
	}
	if cap(param.Data) > 4096 {
		param.Data = param.Data[:0:0]
	} else {
		param.Data = param.Data[:0]
	}
	d.pool.Put(param)
}
