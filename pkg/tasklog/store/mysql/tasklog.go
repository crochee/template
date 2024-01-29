package mysql

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	e "template/pkg/code"
	"template/pkg/tasklog/model"
	"template/pkg/tasklog/request"
)

func NewTaskLog(db *gorm.DB) *taskLog {
	return &taskLog{
		DB: db,
	}
}

type taskLog struct {
	*gorm.DB
}

// Create 创建
func (r *taskLog) Create(ctx context.Context, data *model.TaskLog) (string, error) {
	if err := r.WithContext(ctx).Model(data).Create(data).Error; err != nil {
		return "", errors.WithStack(err)
	}
	return data.PK(), nil
}

// Delete 根据ID删除
func (r *taskLog) Delete(ctx context.Context, id string) error {
	if err := r.WithContext(ctx).Model(&model.TaskLog{}).Where("id = ?", id).
		Delete(&model.TaskLog{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteByIds 根据ID批量删除
func (r *taskLog) DeleteByIds(ctx context.Context, ids []string) error {
	if err := r.WithContext(ctx).Model(&model.TaskLog{}).Where("id IN (?)", ids).
		Delete(&model.TaskLog{}).Error; err != nil {
		return errors.WithStack(e.ErrCodeNotFound.WithResult(err.Error()))
	}
	return nil
}

// Update 根据id 更新 ，排除零值
func (r *taskLog) Update(ctx context.Context, id string, values interface{}) error {
	query := r.WithContext(ctx).Model(&model.TaskLog{}).
		Where("id = ?", id).Updates(values)
	if err := query.Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Get (id int64)  根据id获取model
func (r *taskLog) Get(ctx context.Context, id string, selectQuery ...string) (*model.TaskLog, error) {
	var obj model.TaskLog
	query := r.WithContext(ctx).Model(&model.TaskLog{})
	if len(selectQuery) != 0 {
		query = query.Select(selectQuery[0])
	}
	if err := query.Where("id = ?", id).First(&obj).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, errors.WithStack(e.ErrCodeNotFound.WithResult("TaskLog"))
		}
		return nil, errors.WithStack(err)
	}
	return &obj, nil
}

// List 按条件分页查询
func (r *taskLog) List(ctx context.Context, data *request.QueryTaskLogReq) ([]*model.TaskLog, error) {
	var list []*model.TaskLog
	query := r.WithContext(ctx).Model(&model.TaskLog{})
	if data.ShowInfo != 1 {
		query = query.Select("id,created_at,updated_at,task_name,task_type,progress,res_id,trace_id,status")
	}

	if data.TraceId != "" {
		query = query.Where("trace_id = ?", data.TraceId)
	}

	if data.ResourceId != "" {
		query = query.Where("resource_id = ?", data.ResourceId)
	}

	if data.TaskType != "" {
		query = query.Where("task_type = ?", data.TaskType)
	}

	if err := data.Build(ctx, query).
		Order("start_time DESC").Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if len(list) == 1 && data.PageSize == 1 && data.IsRelation == 1 {
		query = r.WithContext(ctx).Model(&model.TaskLog{})
		if data.ShowInfo != 1 {
			query = query.Select("id,created_at,updated_at,task_name,task_type,progress,res_id,trace_id,status")
		}
		query.Where("trace_id", list[0].TraceId).Find(&list)
	}

	return list, nil
}
