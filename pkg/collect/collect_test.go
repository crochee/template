package collect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"template/pkg/conc/pool"
	"template/pkg/job"
	"template/pkg/logger"
)

type mockStore struct {
	rw    sync.RWMutex
	inner map[uint64]string
}

func (mo *mockStore) Get(ctx context.Context, id uint64, out *MetricWithUpdate) error {
	mo.rw.RLock()
	defer mo.rw.RUnlock()
	v, ok := mo.inner[id]
	if ok {
		return json.Unmarshal([]byte(v), out)
	}
	return errors.New("not found")
}

func (mo *mockStore) Put(ctx context.Context, id uint64, value *MetricWithUpdate) error {
	mo.rw.Lock()
	defer mo.rw.Unlock()
	result, err := json.Marshal(value)
	if err != nil {
		return err
	}
	mo.inner[id] = string(result)
	return nil
}

func (mo *mockStore) List(ctx context.Context, opts *Opts) ([]*MetricWithUpdate, error) {
	mo.rw.RLock()
	defer mo.rw.RUnlock()
	list := make([]*MetricWithUpdate, 0, len(mo.inner))
	for _, v := range mo.inner {
		var out MetricWithUpdate
		err := json.Unmarshal([]byte(v), &out)
		if err != nil {
			return nil, err
		}
		if opts != nil {
			if opts.Name != "" && opts.Name != out.Name {
				continue
			}

			if opts.Help != "" && opts.Help != out.Help {
				continue
			}

			if opts.UpdatedIndex != 0 && opts.UpdatedIndex <= out.UpdatedAt {
				continue
			}
			selectFlag := true
			for k, l := range opts.Label {
				for _, o := range out.Metric.Label {
					if k == *o.Name && l != *o.Value {
						selectFlag = false
						break
					}
				}
			}
			if !selectFlag {
				continue
			}
		}
		list = append(list, &out)
	}
	return list, nil
}

func NewKeyJob(
	name string,
	toleration int64,
	labels prometheus.Labels,
	store Store,
	gv *prometheus.GaugeVec,
) *keyJobSample {
	sortKeys := make([]string, 0, len(labels))
	for key := range labels {
		sortKeys = append(sortKeys, key)
	}
	sort.Strings(sortKeys)

	key := fmt.Sprintf("%s_%d:", name, toleration)
	for _, k := range sortKeys {
		key += labels[k]
		key += "_"
	}
	return &keyJobSample{
		name:       name,
		toleration: toleration,
		label:      labels,
		key:        key,
		store:      store,
		gv:         gv,
	}
}

type keyJobSample struct {
	name       string
	toleration int64
	label      prometheus.Labels
	key        string
	store      Store
	gv         *prometheus.GaugeVec
}

// Description returns the description of the Job.
func (ke *keyJobSample) Description() string {
	return ke.name
}

// Key returns the unique key for the Job.
func (ke *keyJobSample) Key() string {
	return ke.key
}

// Execute is called by a SchedulerRuntime when the Trigger associated with this job fires.
func (ke *keyJobSample) Execute(ctx context.Context) {
	// 存储端过滤指定的逻辑
	list, err := ke.store.List(ctx,
		&Opts{
			Name:         ke.name,
			UpdatedIndex: time.Now().UnixMilli() - ke.toleration,
			Label:        ke.label,
		})
	if err != nil {
		return
	}
	// 存在有数据在2000ms的则不进行处理
	if len(list) != 0 {
		return
	}
	// 获取指标得逻辑
	// code

	// 指标数据备份
	gvi := ke.gv.With(ke.label)
	gvi.Add(3)
	if err = BackupMetric(ctx, ke.store, gvi); err != nil {
		return
	}
}

func TestCollect(t *testing.T) {
	tw := job.NewTimeWheel()
	store := &mockStore{
		inner: map[uint64]string{},
	}
	// db, err := storage.New(
	// 	context.Background(),
	// 	storage.WithDatabase("template"),
	// 	storage.WithUser("root"),
	// 	storage.WithPassword("1234567"),
	// )
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// store := NewStore(db)

	gv := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_vec",
			Help: "test  vec gauge",
		},
		[]string{"a", "b"},
	)
	reg := prometheus.NewRegistry()
	if err := reg.Register(gv); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = logger.With(ctx, logger.New())
	g := pool.New().WithContext(ctx)
	g.Go(tw.Start)
	// 业务定时任务注册
	g.Go(func(ctx context.Context) error {
		return tw.ScheduleJob(ctx, job.NewFuncJob(func(ctx context.Context) {
			labels := []map[string]string{
				{
					"a": "1",
					"b": "2",
				},
				{
					"a": "3",
					"b": "4",
				},
			}
			for _, label := range labels {
				// 调度子任务
				t1, err := job.ParseStandard("0/3 * * * * ?")
				if err != nil {
					t.Log(err)
					return
				}
				tempJob := NewKeyJob("test_vec", 2000, prometheus.Labels(label), store, gv)

				if err = tw.ScheduleJob(ctx, tempJob, t1); err != nil {
					t.Log(err)
					return
				}
				// or
				// err = tw.DeleteJob(ctx, tempJob.Key())
				// if err != nil {
				// 	return
				// }
			}
		}), job.Every(3*time.Second))
	})
	// 业务驱动侧
	g.Go(func(ctx context.Context) error {
		timer := time.NewTicker(3 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-timer.C:
				// 业务逻辑
				// code

				// 指标数据备份
				gvi := gv.WithLabelValues("1", "2")
				gvi.Add(1)
				if err := BackupMetric(ctx, store, gvi); err != nil {
					t.Log(err)
				}
			}
		}
	})
	// export
	g.Go(func(ctx context.Context) error {
		timer := time.NewTicker(time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-timer.C:
				// 采集端逻辑
				f, err := reg.Gather()
				if err != nil {
					return err
				}
				t.Logf("%+v", f)
				// 存储端逻辑
				res, err := store.List(ctx, nil)
				if err != nil {
					return err
				}
				for _, v := range res {
					t.Logf("%+v", v)
				}
			}
		}
	})
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}
