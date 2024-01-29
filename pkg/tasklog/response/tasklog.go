package response

import (
	database "template/pkg/storage"
	"template/pkg/tasklog/model"
)

type ListTaskLogRes struct {
	database.Pagination
	List []*TaskLogRes `json:"list"`
}

type TaskLogRes struct {
	*model.TaskLog
	Cost int64 `json:"cost"` // 执行耗时
}
