package msg

import (
	"sync/atomic"
)

type switchStatus struct {
	old uint32
	now uint32
}

func NewSwitchStatus() *switchStatus {
	return &switchStatus{
		old: 2, // 1激活  2关闭
		now: 2,
	}
}

func (s *switchStatus) Output() bool {
	return s.now == 1
}

func (s *switchStatus) Input(enable bool) {
	if enable {
		atomic.StoreUint32(&s.now, 1)
		return
	}
	atomic.StoreUint32(&s.now, 2)
}

func (s *switchStatus) CheckChangeClose(f func()) {
	if atomic.LoadUint32(&s.now) == 2 && atomic.LoadUint32(&s.old) == 1 { // 1->2  由开到关
		f()
	}
	atomic.StoreUint32(&s.old, atomic.LoadUint32(&s.now))
}

func (s *switchStatus) CheckChangeOpen(f func()) {
	if atomic.LoadUint32(&s.now) == 1 && atomic.LoadUint32(&s.old) == 2 { // 2->1  由关到开,唤醒
		f()
	}
	atomic.StoreUint32(&s.old, atomic.LoadUint32(&s.now))
}
