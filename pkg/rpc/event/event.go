package event

import (
	"context"
	"errors"
	"fmt"
	"strings"

	amqprpc "github.com/0x4b53/amqp-rpc/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"template/pkg/idx"
	"template/pkg/json"
	"template/pkg/logger"
	"template/pkg/replace"
	"template/pkg/rpc"
	"template/pkg/utils"
)

// Event action name list
const (
	actionCreateEvent = "create_event"
	actionUpdateEvent = "update_event"
	ActionFinishEvent = "finish_event"
)

// Event common header used by rpc message
const (
	HeaderActionName = "X-Action-Name"
)

// Event status
const (
	statusRunning = "running"
	statusSuccess = "success"
	statusFailure = "failure"
)

func SetClient(c *rpc.RpcClient,
	getEventIDFunc func(context.Context) uint64,
	setEventIDFunc func(context.Context, uint64) context.Context) {
	client = c
	getEventID = getEventIDFunc
	setEventID = setEventIDFunc
}

var client *rpc.RpcClient

func send(ctx context.Context, actionName string, msgBody interface{}) error {
	exchange := viper.GetString("rabbitmq.producer.dcs_event.exchange")
	routingKey := viper.GetString("rabbitmq.producer.dcs_event.routing-key")
	if exchange == "" || routingKey == "" {
		logger.From(ctx).Error("can't go exchange or routingKey from config file")
		return errors.New("got no exchange or routingKey from config file")
	}

	body, err := json.Marshal(msgBody)
	if err != nil {
		logger.From(ctx).Error("failed to marshal message body")
		return err
	}

	req := amqprpc.NewRequest().
		WithExchange(exchange).
		WithRoutingKey(routingKey).
		WithBody(replace.PwdReplacerReplaceStr(string(body)))
	req.WriteHeader(HeaderActionName, actionName)

	return client.Cast(ctx, req)
}

type createEventBody struct {
	// The reason why we send the event id to the server is that we
	// want the create event request to be asynchronous, so the create
	// action will not block the normal task.
	ID uint64 `json:"id"`

	// 操作名称, 如创建，更新，删除等
	Action string `json:"action"`

	// 资源类型，如云主机，分布式网络等
	ResType string `json:"res_type"`

	// 事件状态
	Status string `json:"status"`

	// 操作的资源ID
	ResID uint64 `json:"res_id,omitempty"`

	// 操作的资源名称
	ResName string `json:"res_name,omitempty"`

	// 关联资源ID
	LinkedResID uint64 `json:"linked_res_id,omitempty"`

	// 关联资源名称
	LinkedResName string `json:"linked_res_name,omitempty"`

	// 请求参数
	Request interface{} `json:"request,omitempty"`

	// 对应rudder的task id
	TaskID uint64 `json:"task_id,omitempty"`
}

func doCreateEvent(ctx context.Context, msg *createEventBody) (uint64, error) {
	// Generate a snow id as event id
	entID, err := idx.NextID()
	if err != nil {
		logger.From(ctx).Error("failed to generate event id", zap.Error(err))
		return 0, err
	}
	msg.ID = entID

	if err = send(ctx, actionCreateEvent, msg); err != nil {
		logger.From(ctx).Error("failed to send message", zap.Error(err))
		return entID, err
	}
	return entID, nil
}

type updateEventBody struct {
	// 事件ID
	ID uint64 `json:"id"`

	// 对应rudder的TaskID
	TaskID uint64 `json:"task_id,omitempty"`

	// 操作的资源ID
	ResID uint64 `json:"res_id,omitempty"`

	// 操作的资源名称
	ResName string `json:"res_name,omitempty"`

	// 关联资源ID
	LinkedResID uint64 `json:"linked_res_id,omitempty"`

	// 关联资源名称
	LinkedResName string `json:"linked_res_name,omitempty"`

	// 请求响应，主要用于更新异步操作响应
	Response interface{} `json:"response,omitempty"`
}

func doUpdateEvent(ctx context.Context, id uint64, attrs *updateEventBody) error {
	attrs.ID = id
	if err := send(ctx, actionUpdateEvent, attrs); err != nil {
		logger.From(ctx).Error("failed to update event", zap.Error(err))
		return err
	}

	return nil
}

type finishEventBody struct {
	// 事件ID
	ID uint64 `json:"id"`

	// 请求响应，主要用于记录同步请求响应
	Response interface{} `json:"response,omitempty"`

	// 事件结果，成功或者失败
	Status string `json:"status"`

	// 可选，如果事件失败，用于记录失败原因
	Reason string `json:"reason,omitempty"`
}

func doFinishEvent(ctx context.Context, id uint64, body interface{}, status string, reason ...string) error {
	attrs := new(finishEventBody)
	attrs.ID = id
	attrs.Response = body
	attrs.Status = status
	if len(reason) > 0 {
		attrs.Reason = strings.Join(reason, ";")
	}
	if err := send(ctx, ActionFinishEvent, attrs); err != nil {
		logger.From(ctx).Error("failed to finish event", zap.Error(err))
		return err
	}

	return nil
}

// ------------------------------------------DCS EVENT SUGAR FUNCTION---------------------------------------------------

var (
	getEventID func(context.Context) uint64
	setEventID func(context.Context, uint64) context.Context
)

// CreateEvent will create an dcs event via rpc server.
// if action success, the event id will be returned, otherwise,
// an error will be returned to nitify the action is failed.
// `linkedRes` is a two tuple, the element is [linkedResID(uint64), linkedResName(string)]
func CreateEvent(ctx context.Context, action, resType, resID, resName string, params interface{}, linkedRes ...interface{}) (uint64, error) {
	tmpID, err := utils.ToUint64(resID)
	if err != nil {
		return 0, err
	}
	if eventID := getEventID(ctx); eventID != 0 {
		return eventID, nil
	}
	msg := createEventBody{
		Action:  action,
		ResType: resType,
		ResID:   tmpID,
		ResName: resName,
		Request: params,
		Status:  statusRunning,
	}
	// linked resource id
	if len(linkedRes) > 0 {
		linkedResID, ok := linkedRes[0].(uint64)
		if !ok {
			return 0, fmt.Errorf("invalid linked res id:%v", linkedResID)
		}
		msg.LinkedResID = linkedResID
	}
	// linked resource name
	if len(linkedRes) > 1 {
		linkedResName, ok := linkedRes[1].(string)
		if !ok {
			return 0, fmt.Errorf("invalid linked res name:%v", linkedRes[1])
		}
		msg.LinkedResName = linkedResName
	}

	return doCreateEvent(ctx, &msg)
}

// CreateEventContext 新建一个 event ,并将event id 添加到 ctx中返回
func CreateEventContext(ctx context.Context, action, resType string, resID string, resName string, params interface{}, linkedRes ...interface{}) (context.Context, uint64, error) {
	tmpID, err := utils.ToUint64(resID)
	if err != nil {
		return ctx, 0, err
	}
	msg := createEventBody{
		Action:  action,
		ResType: resType,
		ResID:   tmpID,
		ResName: resName,
		Request: params,
		Status:  statusRunning,
	}
	// linked resource id
	if len(linkedRes) > 0 {
		linkedResID, ok := linkedRes[0].(uint64)
		if !ok {
			return ctx, 0, fmt.Errorf("invalid linked res id:%v", linkedResID)
		}
		msg.LinkedResID = linkedResID
	}
	// linked resource name
	if len(linkedRes) > 1 {
		linkedResName, ok := linkedRes[1].(string)
		if !ok {
			return ctx, 0, fmt.Errorf("invalid linked res name:%v", linkedRes[1])
		}
		msg.LinkedResName = linkedResName
	}
	eventID, err := doCreateEvent(ctx, &msg)
	return setEventID(ctx, eventID), eventID, err
}

// SetEventToSuccess will mark the dcs event to success and finish the event.
func SetEventToSuccess(ctx context.Context, id uint64, body ...interface{}) error {
	if eventID := getEventID(ctx); eventID == id {
		return nil
	}
	var response interface{}
	if len(body) > 0 {
		response = body[0]
	}
	return doFinishEvent(ctx, id, response, statusSuccess)
}

// SetEventToFailed will mark the dcs event to failure and finish the event.
func SetEventToFailed(ctx context.Context, id uint64, reason ...string) error {
	if id == 0 {
		return nil
	}
	if eventID := getEventID(ctx); eventID == id {
		return nil
	}
	return doFinishEvent(ctx, id, nil, statusFailure, reason...)
}

// SetEventLinkedRes will set linked resource to the event.
func SetEventLinkedRes(ctx context.Context, id uint64, resID uint64, resName string) error {
	attrs := updateEventBody{
		LinkedResID:   resID,
		LinkedResName: resName,
	}
	return doUpdateEvent(ctx, id, &attrs)
}

// SetEventRespBody will send the api response to the event.
func SetEventRespBody(ctx context.Context, id uint64, body interface{}) error {
	if eventID := getEventID(ctx); eventID == id {
		return nil
	}

	attrs := updateEventBody{
		Response: body,
	}
	return doUpdateEvent(ctx, id, &attrs)
}

// SetEventRespBodyForce will send the api response to the event.
func SetEventRespBodyForce(ctx context.Context, id uint64, body interface{}) error {

	attrs := updateEventBody{
		Response: body,
	}
	return doUpdateEvent(ctx, id, &attrs)
}

// SetEventTaskID will add rudder task id to the dcs event.
func SetEventTaskID(ctx context.Context, id, taskID uint64) error {
	attrs := updateEventBody{
		TaskID: taskID,
	}

	return doUpdateEvent(ctx, id, &attrs)
}

// SetEventResID will add res id to the dcs event.
func SetEventResID(ctx context.Context, id uint64, resID string) error {
	tmpID, err := utils.ToUint64(resID)
	if err != nil {
		return err
	}
	attrs := updateEventBody{
		ResID: tmpID,
	}

	return doUpdateEvent(ctx, id, &attrs)
}

// SetEventStatus will set a new status to the dcs event
func SetEventStatus(ctx context.Context, id uint64, status string) error {
	if eventID := getEventID(ctx); eventID == id {
		return nil
	}
	if status == "error" {
		status = statusFailure
	}
	return doFinishEvent(ctx, id, nil, status)
}

// SetEventToSuccessForce will mark the dcs event to success and finish the event.
func SetEventToSuccessForce(ctx context.Context, id uint64, body ...interface{}) error {
	var response interface{}
	if len(body) > 0 {
		response = body[0]
	}
	return doFinishEvent(ctx, id, response, statusSuccess)
}

// SetEventToFailedForce will mark the dcs event to failure and finish the event.
func SetEventToFailedForce(ctx context.Context, id uint64, reason ...string) error {
	return doFinishEvent(ctx, id, nil, statusFailure, reason...)
}

// SetEventStatusForce will set a new status to the dcs event
func SetEventStatusForce(ctx context.Context, id uint64, status string) error {
	if status == "error" {
		status = statusFailure
	}
	return doFinishEvent(ctx, id, nil, status)
}
