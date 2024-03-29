package job

import (
	"container/list"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"

	"template/pkg/conc/pool"
)

type delayTask struct {
	delay int64
	Job
	Trigger
}

type entry struct {
	delayTask
	circle int
}

func WithInterval(interval time.Duration) Option {
	return func(o *option) {
		o.interval = interval
	}
}

func WithSlot(slotNum int) Option {
	return func(o *option) {
		o.slotNum = slotNum
	}
}

type option struct {
	interval time.Duration
	slotNum  int
	nowFunc  func() int64
}

type Option func(*option)

func NewTimeWheel(opts ...Option) SchedulerRuntime {
	o := &option{
		interval: time.Second,
		slotNum:  1024,
		nowFunc: func() int64 {
			return time.Now().UnixNano()
		},
	}
	for _, opt := range opts {
		opt(o)
	}
	t := &timeWheel{
		interval:          o.interval,
		slots:             make([]*list.List, o.slotNum),
		timerMap:          cmap.New(),
		cur:               o.slotNum - 1,
		slotSum:           o.slotNum,
		addTaskChannel:    make(chan *entry),
		removeTaskChannel: make(chan string),
		nowFunc:           o.nowFunc,
	}
	for i := 0; i < t.slotSum; i++ {
		t.slots[i] = list.New()
	}
	return t
}

type timeWheel struct {
	interval time.Duration // 指针每隔多久往前移动一格
	slots    []*list.List  // 时间轮槽
	// key: 定时器唯一标识 value: 定时器所在的槽, 主要用于删除定时器
	timerMap          cmap.ConcurrentMap
	cur               int         // 当前指针指向哪一个槽
	slotSum           int         // 槽数量
	addTaskChannel    chan *entry // 新增任务channel
	removeTaskChannel chan string // 删除任务channel

	pool    *pool.ContextPool
	nowFunc func() int64

	running uint32
}

func (t *timeWheel) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapUint32(&t.running, 0, 1) {
		return nil
	}
	ticker := time.NewTicker(t.interval)
	t.pool = pool.New().WithContext(ctx)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return t.pool.Wait()
		case <-ticker.C:
			t.handler()
		case task := <-t.addTaskChannel:
			t.addTask(task)
		case key := <-t.removeTaskChannel:
			t.removeJob(key)
		}
	}
}

// 获取定时器在槽中的位置, 时间轮需要转动的圈数
func (t *timeWheel) getPositionAndCircle(d int64) (int, int) {
	steps := d / t.interval.Nanoseconds()
	pos := (int64(t.cur) + steps) % int64(t.slotSum)
	circle := (steps - 1) / int64(t.slotSum)
	return int(pos), int(circle)
}

// 新增任务到链表中
func (t *timeWheel) addTask(task *entry) {
	pos, circle := t.getPositionAndCircle(task.delay)
	task.circle = circle
	t.timerMap.Set(task.Key(), pos)
	t.slots[pos].PushBack(task)
}

func (t *timeWheel) ScheduleJob(ctx context.Context, job Job, trigger Trigger) error {
	_, found := t.timerMap.Get(job.Key())
	if found {
		return fmt.Errorf("found key %s", job.Key())
	}
	task, err := t.createTask(job, trigger)
	if err != nil {
		return err
	}
	if atomic.LoadUint32(&t.running) == 0 {
		// 当系统为未运行，则直接将任务添加
		t.addTask(task)
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case t.addTaskChannel <- task:
		return nil
	}
}

func (t *timeWheel) GetJobKeys() []string {
	return t.timerMap.Keys()
}

func (t *timeWheel) GetScheduledJob(key string) (*ScheduledJob, error) {
	value, found := t.timerMap.Get(key)
	if !found {
		return nil, fmt.Errorf("not found key %s", key)
	}
	pos, ok := value.(int)
	if !ok {
		return nil, fmt.Errorf("invalid value %v", value)
	}
	l := t.slots[pos]
	for e := l.Front(); e != nil; {
		task, ok := e.Value.(*entry)
		if !ok {
			next := e.Next()
			l.Remove(e)
			e = next
			continue
		}
		if task.Key() == key {
			return &ScheduledJob{
				Job:     task.Job,
				Trigger: task.Trigger,
			}, nil
		}
		e = e.Next()
	}
	return nil, fmt.Errorf("not found key %s in slots", key)
}

func (t *timeWheel) DeleteJob(ctx context.Context, key string) error {
	if atomic.LoadUint32(&t.running) == 0 {
		t.removeJob(key)
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case t.removeTaskChannel <- key:
		return nil
	}
}

func (t *timeWheel) Has(key string) bool {
	return t.timerMap.Has(key)
}

// 标记job删除
func (t *timeWheel) removeJob(key string) {
	value, found := t.timerMap.Get(key)
	if !found {
		return
	}
	pos, ok := value.(int)
	if !ok {
		return
	}
	l := t.slots[pos]
	for e := l.Front(); e != nil; {
		task, ok := e.Value.(*entry)
		if !ok {
			// 清楚脏数据
			next := e.Next()
			l.Remove(e)
			e = next
			continue
		}
		if task.Key() == key {
			// 删除
			l.Remove(e)
			// 删除位置信息
			t.timerMap.Remove(key)
			return
		}
		e = e.Next()
	}
}

func (t *timeWheel) handler() {
	t.cur = (t.cur + 1) % t.slotSum
	l := t.slots[t.cur]
	t.scanAndRunTask(l)
}

// 扫描链表中过期定时器, 并执行回调函数
func (t *timeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task, ok := e.Value.(*entry)
		if !ok { // 清楚脏数据
			next := e.Next()
			l.Remove(e)
			e = next
			continue
		}
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		// 任务执行
		t.pool.Go(func(ctx context.Context) error {
			task.Execute(ctx)
			return nil
		})
		next := e.Next()
		l.Remove(e)
		// 重新调度
		t.moveTask(task)
		e = next
	}
}

func (t *timeWheel) createTask(job Job, trigger Trigger) (*entry, error) {
	now := t.nowFunc()
	nextRunTime, err := trigger.NextFireTime(now)
	if err != nil {
		return nil, err
	}
	// calculate the relative value of the current time
	delay := nextRunTime - now
	if delay < t.interval.Nanoseconds() {
		delay = t.interval.Nanoseconds()
	}
	return &entry{
		delayTask: delayTask{
			delay:   delay,
			Job:     job,
			Trigger: trigger,
		},
	}, nil
}

func (t *timeWheel) moveTask(task *entry) {
	taskValue, err := t.createTask(task.Job, task.Trigger)
	if err != nil {
		t.timerMap.Remove(task.Key())
		if errors.Is(err, ErrSkipScheduleJob) {
			return
		}
		return
	}
	t.addTask(taskValue)
}
