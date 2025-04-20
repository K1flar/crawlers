package worker

import (
	"context"
	"sync"
	"time"
)

type job func(ctx context.Context)

type Worker struct {
	wg  *sync.WaitGroup
	job job
}

type Cron struct {
	Worker
	period time.Duration
}

func New(job job) *Worker {
	return &Worker{
		wg:  &sync.WaitGroup{},
		job: job,
	}
}

func NewWithPeriod(job job, period time.Duration) *Cron {
	return &Cron{
		Worker: Worker{
			wg:  &sync.WaitGroup{},
			job: job,
		},
		period: period,
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		for {
			w.job(ctx)
		}
	}()
}

func (w *Worker) Wait() {
	w.wg.Wait()
}

func (c *Cron) Run(ctx context.Context) {
	ticker := time.NewTicker(c.period)

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		for {
			<-ticker.C
			c.job(ctx)
		}
	}()
}

func (c *Cron) Wait() {
	c.wg.Wait()
}
