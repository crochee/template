package syncx

type Locker interface {
	Lock() error
	TryLock() error
	Unlock() error
}

type RWMutexLocker interface {
	Locker
	RLock() error
	RUnlock() error
}

type NoopLocker struct {
}

func (NoopLocker) Lock() error {
	return nil
}

func (NoopLocker) Unlock() error {
	return nil
}
