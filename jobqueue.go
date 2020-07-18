package jobqueue

import (
	"context"
	"sync"
)

type Job interface {
	Run(ctx context.Context)
}

type Queue struct {
	jobs   chan *Job
	queue  []*Job
	mu     *sync.Mutex
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel func()
	end    chan int
}

func NewQueue(maxJobCount int) *Queue {
	return WithContext(context.Background(), maxJobCount)
}

func WithContext(ctx context.Context, maxJobCount int) *Queue {
	childCtx, cancel := context.WithCancel(ctx)

	q := &Queue{
		jobs:   make(chan *Job, maxJobCount),
		queue:  []*Job{},
		mu:     &sync.Mutex{},
		wg:     &sync.WaitGroup{},
		ctx:    childCtx,
		cancel: cancel,
		end:    make(chan int, 1),
	}

	go q.start()

	return q
}

func (q *Queue) runJob(job *Job) {
	defer q.wg.Done()
	(*job).Run(q.ctx)
}

func (q *Queue) start() {
	for {
		select {
		case <-q.ctx.Done():
			q.wg.Add(-len(q.jobs))
			q.wg.Wait()
			q.end <- 1
			return

		case j := <-q.jobs:
			go func(job *Job) {
				q.runJob(job)
				if q.ctx.Err() != nil {
					return
				}

				q.mu.Lock()
				defer q.mu.Unlock()
				if len(q.queue) == 0 {
					return
				}

				q.wg.Add(1)
				q.jobs <- q.queue[0]
				q.queue = q.queue[1:]
			}(j)
		}
	}
}

func (q *Queue) Add(job Job) {
	if q.ctx.Err() != nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) < cap(q.jobs) {
		q.wg.Add(1)
		q.jobs <- &job
	} else {
		q.queue = append(q.queue, &job)
	}
}

func (q *Queue) Wait() {
	q.wg.Wait()
}

func (q *Queue) Stop() {
	q.cancel()
	<-q.end
}
