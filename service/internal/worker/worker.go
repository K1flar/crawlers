package worker

import (
	"context"
	"time"
)

type job func(ctx context.Context)

type Worker struct {
	job job
}

type Cron struct {
	Worker
	period time.Duration
}

func New(job job) *Worker {
	return &Worker{
		job: job,
	}
}

func NewWithPeriod(job job, period time.Duration) *Cron {
	return &Cron{
		Worker: Worker{
			job: job,
		},
		period: period,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		w.job(ctx)
	}
}

func (c *Cron) Run(ctx context.Context) {
	ticker := time.NewTicker(c.period)

	for {
		<-ticker.C
		c.job(ctx)
	}
}
