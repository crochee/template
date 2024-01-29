package tasklog

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	e "template/pkg/code"
	log "template/pkg/logger"
	"template/pkg/resp"
	"template/pkg/tasklog/model"
	"template/pkg/tasklog/request"
	"template/pkg/tasklog/response"
	"template/pkg/tasklog/store"
)

type ClientFunc func(context.Context) *gorm.DB

var (
	progressCache = cache.New(time.Hour, time.Minute)
	taskLogStore  store.TaskLog
	getTraceID    func(ctx context.Context) string
)

func RegisterDB(ctx context.Context, f ClientFunc, getTraceIDFunc func(ctx context.Context) string) error {
	getTraceID = getTraceIDFunc
	db := f(ctx)
	taskLogStore = store.NewTaskLogStore(db)
	return db.Set("gorm:table_options",
		"ENGINE=InnoDB COMMENT='任务执行日志记录数据表' DEFAULT CHARSET='utf8mb4'").
		AutoMigrate(&model.TaskLog{})
}

func RegisterAPI(router *gin.Engine) {
	router.GET("/task-logs/:id", GetTaskLog)
	router.GET("/task-logs", ListTaskLog)
}

func NewTaskLog() *TaskLogClient {
	return &TaskLogClient{
		taskLog:     model.TaskLog{StartTime: time.Now()},
		resourceNum: 0,
		completeNum: 0,
	}
}

type TaskLogClient struct {
	taskLog     model.TaskLog
	resourceNum int
	completeNum int32
}

func (t *TaskLogClient) ResourceNum(num int) *TaskLogClient {
	t.resourceNum = num
	return t
}

func (t *TaskLogClient) ResourceId(resId uint64) *TaskLogClient {
	t.taskLog.ResId = resId
	return t
}

func (t *TaskLogClient) InputParam(input interface{}) *TaskLogClient {
	if input == nil {
		return t
	}
	t.taskLog.Input, _ = json.Marshal(input)
	return t
}

func (t *TaskLogClient) Create(ctx context.Context, name, taskType string) error {
	var err error
	t.taskLog.TaskName = name
	t.taskLog.TaskType = taskType
	t.taskLog.TraceId = getTraceID(ctx)
	t.taskLog.Status = model.StatusRunning

	_, err = taskLogStore.Create(ctx, &t.taskLog)
	if err != nil {
		log.FromContext(ctx).Err(err).Interface("task_log", t.taskLog).Msg("fail to create task log")
	}
	return err
}

func (t *TaskLogClient) Success(ctx context.Context, result interface{}) error {
	var resultByte []byte
	var err error
	if result != nil {
		resultByte, err = json.Marshal(result)
		if err != nil {
			return err
		}
	}
	err = taskLogStore.Update(ctx, t.taskLog.PK(), map[string]interface{}{
		"result":   resultByte,
		"status":   model.StatusSuccess,
		"progress": 100,
		"end_time": time.Now(),
	})
	if err != nil {
		log.FromContext(ctx).Err(err).Interface("result", result).
			Str("id", t.taskLog.PK()).
			Msg("fail to update task log fail")
	}
	return err
}

func (t *TaskLogClient) Failure(ctx context.Context, reason string) error {
	err := taskLogStore.Update(ctx, t.taskLog.PK(), map[string]interface{}{
		"reason":   reason,
		"status":   model.StatusFail,
		"end_time": time.Now(),
		"progress": getProgress(t.taskLog.PK()),
	})
	if err != nil {
		log.FromContext(ctx).Err(err).Str("reason", reason).
			Str("id", t.taskLog.PK()).
			Msg("fail to update task log fail")
	}
	return err
}

func (t *TaskLogClient) Finish(ctx context.Context, result interface{}, resErr error) error {
	var resultByte []byte
	var reason string
	var err error
	if result != nil {
		resultByte, err = json.Marshal(result)
		if err != nil {
			return err
		}
	}
	status := model.StatusSuccess
	if resErr != nil {
		reason = resErr.Error()
		status = model.StatusFail
	}
	param := map[string]interface{}{
		"result":   resultByte,
		"reason":   reason,
		"status":   status,
		"end_time": time.Now(),
	}
	if resErr == nil {
		param["progress"] = 100
	} else {
		param["progress"] = getProgress(t.taskLog.PK())
	}

	err = taskLogStore.Update(ctx, t.taskLog.PK(), param)
	if err != nil {
		log.FromContext(ctx).Err(err).Str("reason", reason).
			Str("id", t.taskLog.PK()).
			Msg("fail to update task log fail")
	}
	return err
}

func (t *TaskLogClient) RefreshProgress(ctx context.Context, complete int) error {
	if t.resourceNum == 0 {
		return nil
	}
	// TODO 是否有必要加锁
	completeNum := int(atomic.AddInt32(&t.completeNum, int32(complete)))
	if completeNum > t.resourceNum {
		completeNum = t.resourceNum
		atomic.StoreInt32(&t.completeNum, int32(completeNum))
	}
	progress := (completeNum * 100) / t.resourceNum
	progressCache.SetDefault(t.taskLog.PK(), progress)
	return nil
}

func getProgress(id string) int {
	var progress int
	progressObj, ok := progressCache.Get(id)
	if ok {
		progress, _ = progressObj.(int)
	}
	return progress
}

func GetTaskLog(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	if id == "" {
		resp.Error(c, e.ErrCodeInvalidParam)
		return
	}

	taskLog, err := taskLogStore.Get(ctx, id)
	if err != nil {
		resp.Error(c, err)
		return
	}
	var cost int64
	if taskLog.Status == model.StatusRunning {
		cost = time.Since(taskLog.StartTime).Milliseconds()
		taskLog.Progress = getProgress(taskLog.PK())
	} else {
		cost = taskLog.EndTime.Sub(taskLog.StartTime).Milliseconds()
	}

	c.JSON(http.StatusOK, &response.TaskLogRes{
		TaskLog: taskLog,
		Cost:    cost,
	})
}

func ListTaskLog(c *gin.Context) {
	ctx := c.Request.Context()
	var req request.QueryTaskLogReq

	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, err)
		return
	}
	list, err := taskLogStore.List(ctx, &req)
	if err != nil {
		resp.Error(c, err)
		return
	}
	var cost int64
	result := make([]response.TaskLogRes, len(list))
	for i, log := range list {
		if log.Status == model.StatusRunning {
			cost = time.Since(log.StartTime).Milliseconds()
			list[i].Progress = getProgress(log.PK())
		} else {
			cost = log.EndTime.Sub(log.StartTime).Milliseconds()
		}
		result[i] = response.TaskLogRes{
			TaskLog: list[i],
			Cost:    cost,
		}
	}
	c.JSON(http.StatusOK, result)
}
