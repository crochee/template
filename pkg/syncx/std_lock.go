package syncx

import (
	"sync"

	"github.com/pkg/errors"
)

func NewStdMutex() *stdMutex {
	return &stdMutex{
		mutex: &sync.Mutex{},
	}
}

type stdMutex struct {
	mutex sync.Locker
}

func (st *stdMutex) Lock() error {
	st.mutex.Lock()
	return nil
}

func (st *stdMutex) TryLock() error {
	return errors.New("not support")
}

func (st *stdMutex) Unlock() error {
	st.mutex.Unlock()
	return nil
}

func NewStdRWMutex() *stdRWMutex {
	return &stdRWMutex{}
}

type stdRWMutex struct {
	mutex sync.RWMutex
}

func (st *stdRWMutex) Lock() error {
	st.mutex.Lock()
	return nil
}

func (st *stdRWMutex) Unlock() error {
	st.mutex.Unlock()
	return nil
}

func (st *stdRWMutex) RLock() error {
	st.mutex.RLock()
	return nil
}

func (st *stdRWMutex) RUnlock() error {
	st.mutex.RUnlock()
	return nil
}
