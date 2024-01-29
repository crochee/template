package store

import (
	"context"

	"gorm.io/gorm"

	"template/pkg/tasklog/model"
	"template/pkg/tasklog/request"
	"template/pkg/tasklog/store/mysql"
)

func NewTaskLogStore(db *gorm.DB) TaskLog {
	return mysql.NewTaskLog(db)
}

type TaskLog interface {
	Create(ctx context.Context, data *model.TaskLog) (string, error)
	Update(ctx context.Context, id string, values interface{}) error
	Delete(ctx context.Context, id string) error
	DeleteByIds(ctx context.Context, ids []string) error
	Get(ctx context.Context, id string, selectQuery ...string) (*model.TaskLog, error)
	List(ctx context.Context, req *request.QueryTaskLogReq) ([]*model.TaskLog, error)
}
