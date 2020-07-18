package jobqueue

import (
	"context"
	"sync"
)

// Job is interface of instance to stored and executed by Queue
type Job interface {
	Run(ctx context.Context)
}

// Queue is scheduler that store and execute jobs
type Queue struct {
	jobs   chan *Job
	queue  []*Job
	mu     *sync.Mutex
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel func()
	end    chan int
}

// NewQueue create queue instance and start job scheduling
func NewQueue(workerCount int) *Queue {
	return WithContext(context.Background(), workerCount)
}

// WithContext create queue instance controlled with ctx and start job scheduling
func WithContext(ctx context.Context, workerCount int) *Queue {
	childCtx, cancel := context.WithCancel(ctx)

	q := &Queue{
		jobs:   make(chan *Job, workerCount),
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
				defer q.wg.Done()
				(*job).Run(q.ctx)
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

// Enqueue add job to queue's pool
func (q *Queue) Enqueue(job Job) {
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

// Wait wait for all jobs in the queue to finish executing
func (q *Queue) Wait() {
	q.wg.Wait()
}

// Stop stop queue running gracefully
func (q *Queue) Stop() {
	q.cancel()
	<-q.end
}
