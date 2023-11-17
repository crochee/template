package async

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"template/pkg/conc/pool"
)

type testError struct {
}

func (t testError) Name() string {
	return "testError"
}

func (t testError) Run(ctx context.Context, param *Param) error {
	fmt.Println("run testError")
	return errors.New("testError failed")
}

type test struct {
}

func (t test) Name() string {
	return "test"
}

func (t test) Run(ctx context.Context, param *Param) error {
	fmt.Println("90")
	return nil
}

type test1 struct {
	i uint
}

func (t test1) Name() string {
	return "test1"
}

func (t *test1) Run(ctx context.Context, param *Param) error {
	t.i++
	fmt.Printf("91\t %#v\n", t)
	return nil
}

type multiTest struct {
	list []TaskHandler
}

func (t multiTest) Name() string {
	return "multiTest"
}

func (m *multiTest) Run(ctx context.Context, param *Param) error {
	fmt.Println("mt", len(m.list))
	g := pool.New().WithContext(ctx).WithCancelOnError()
	for _, e := range m.list {
		tmp := e
		g.Go(func(ctx context.Context) error {
			return tmp.Run(ctx, nil)
		})
	}
	return g.Wait()
}

func TestRetry(t *testing.T) {
	m := NewManager()
	m.Register(test{})
	m.Register(&test1{})
	m.Register(&multiTest{list: []TaskHandler{test{}, &test1{}}})

	t.Log(m.Run(context.Background(), &Param{
		TaskType: "test",
		Data:     nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		TaskType: "test1",
		Data:     nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		TaskType: "multiTest",
		Data:     nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		TaskType: "async.multiTest",
		Data:     nil,
	}))
}
