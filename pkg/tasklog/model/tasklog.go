package model

import (
	"strconv"
	"time"

	"gorm.io/datatypes"

	database "template/pkg/storage"
)

// TaskLog  任务执行日志记录数据表
type TaskLog struct {
	database.SnowID
	ID        uint64          `json:"id" gorm:"primary_key:id;comment:主键id"`
	CreatedAt time.Time       `json:"created_at" gorm:"column:created_at;not null;index:idx_idx_task_type_created,priority:2;comment:创建时间"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"column:updated_at;not null;comment:更新时间"`
	StartTime time.Time       `json:"start_time" gorm:"column:start_time;not null;index:idx_idx_task_type_start_time,priority:2;comment:开始时间"`
	EndTime   time.Time       `json:"end_time" gorm:"column:end_time;not null;comment:结束时间"`
	TaskName  string          `json:"task_name" gorm:"column:task_name"`                                               //任务名称
	TaskType  string          `json:"task_type" gorm:"column:task_type;index:idx_idx_task_type_start_time,priority:1"` //任务类型
	Progress  int             `json:"progress" gorm:"column:progress"`                                                 //任务的执行进度百分比值
	ResId     uint64          `json:"res_id" gorm:"column:res_id;index:idx_res_id"`                                    //任务操作资源ID
	Input     datatypes.JSON  `json:"input" gorm:"column:input"`                                                       //任务开始输入
	Result    datatypes.JSON  `json:"result" gorm:"column:result"`                                                     //任务结果输出
	TraceId   string          `json:"trace_id" gorm:"column:trace_id;index:idx_trace_id"`                              //追踪trace_id
	Status    database.Status `json:"status" gorm:"column:status"`                                                     //任务状态
	Reason    string          `json:"reason" gorm:"column:reason"`                                                     //失败原因
}

func (TaskLog) TableName() string {
	return "task_log"
}

func (t *TaskLog) PK() string {
	return strconv.FormatUint(t.ID, 10)
}

const (
	StatusSuccess = "success"
	StatusFail    = "fail"
	StatusRunning = "running"
)
