package msg

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_standardQueue_ListAndClear(t *testing.T) {
	q := NewQueue(3)
	q.Write(&Metadata{
		TraceID: "1",
	})
	q.Write(&Metadata{
		TraceID: "2",
	})
	q.Write(&Metadata{
		TraceID: "3",
	})
	t.Log(q.ListAndClear())
}

func TestPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := NewQueuePool(ctx, 5*time.Second, 8*time.Second, func() Queue {
		return NewQueue(1024 * 10)
	})
	p.SendOn(func(metadata []*Metadata) {
		fmt.Println(metadata)
	})
	p.Get("898989").Write(&Metadata{
		TraceID: "898989",
		Summary: "some",
	})
	time.Sleep(20 * time.Second)
}
