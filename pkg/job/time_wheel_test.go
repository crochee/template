package job

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"go.uber.org/zap"

	"template/pkg/logger"
	"template/pkg/routine"
)

type Job1 struct {
}

func (j Job1) Execute(ctx context.Context) {
	logger.From(ctx).Info("running job1")
}

func (j Job1) Description() string {
	return "job1"
}

func (j Job1) Key() string {
	return "1"
}

type Job2 struct {
}

func (j Job2) Execute(ctx context.Context) {
	logger.From(ctx).Info("running job2")
}

func (j Job2) Description() string {
	return "job2"
}

func (j Job2) Key() string {
	return "2"
}

type Job3 struct {
	t string
}

func (j Job3) Execute(ctx context.Context) {
	logger.From(ctx).Info("running", zap.String("key", j.Key()))
}

func (j Job3) Description() string {
	return j.Key()
}

func (j Job3) Key() string {
	return j.t
}

func TestJob(t *testing.T) {
	tw := NewTimeWheel(WithInterval(time.Second), WithSlot(500))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	ctx = logger.With(ctx, logger.New())
	g := routine.NewGroup(ctx)
	g.Go(tw.Start)
	g.Go(func(ctx context.Context) error {
		return tw.ScheduleJob(ctx, Job1{}, Every(5*time.Second))
	})
	g.Go(func(ctx context.Context) error {
		return tw.ScheduleJob(ctx, Job2{}, RunOnce(5*time.Second))
	})
	g.Go(func(ctx context.Context) error {
		t1, err := ParseStandard("@every 1s")
		if err != nil {
			log.Println(err)
			return err
		}
		return tw.ScheduleJob(ctx, Job3{t: "20"}, t1)
	})
	g.Go(func(ctx context.Context) error {
		t1, err := NewParser(SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor).
			Parse("0/3 * * * * ?")
		if err != nil {
			return err
		}
		return tw.ScheduleJob(ctx, Job3{t: "21"}, t1)
	})
	g.Go(func(ctx context.Context) error {
		return tw.ScheduleJob(ctx, Job3{t: "temp"}, RunOnce(0))
	})
	g.Go(func(ctx context.Context) error {
		log.Println(tw.GetJobKeys())
		return nil
	})
	g.Go(func(ctx context.Context) error {
		log.Println(tw.GetScheduledJob("12"))
		return nil
	})
	g.Go(func(ctx context.Context) error {
		log.Println(tw.DeleteJob(ctx, "89"))
		return nil
	})
	g.Go(func(ctx context.Context) error {
		return tw.ScheduleJob(ctx, Job3{t: "temp at"}, RunAt(time.Now().Add(-1*time.Minute).UnixNano()))
	})
	if err := g.Wait(); err != nil && errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
