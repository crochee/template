package async

import (
	"context"
	"fmt"
)

type ManagerTaskHandler interface {
	Register(handlers ...TaskHandler)
	Unregister(names ...string)
	Run(ctx context.Context, param *Param) error
}

// TaskHandler your business should implement it
type TaskHandler interface {
	Name() string
	Run(ctx context.Context, param *Param) error
}

func NewManager() ManagerTaskHandler {
	return &manager{inner: make(map[string]TaskHandler, 40)}
}

type manager struct {
	inner map[string]TaskHandler
}

// Register unsafe
func (m *manager) Register(handlers ...TaskHandler) {
	for _, handler := range handlers {
		m.inner[handler.Name()] = handler
	}
}

// Unregister unsafe
func (m *manager) Unregister(names ...string) {
	for _, name := range names {
		delete(m.inner, name)
	}
}

func (m *manager) Run(ctx context.Context, param *Param) error {
	handler, found := m.inner[param.TaskType]
	if !found {
		return fmt.Errorf("must register %s", param.TaskType)
	}
	return handler.Run(ctx, param)
}
