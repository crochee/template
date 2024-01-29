package request

import (
	database "template/pkg/storage"
)

type QueryTaskLogReq struct {
	database.Pagination
	TraceId    string `form:"trace_id"`    // traceID
	ResourceId string `form:"resource_id"` // 资源ID
	TaskType   string `form:"task_type"`   // 任务类型
	IsRelation int    `form:"is_relation"` // 当 page_size 为 1时 生效，是否通过 trace_id 查出其它数据， 1 表示 是
	ShowInfo   int    `form:"show_info"`   // 1 显示详情
}
