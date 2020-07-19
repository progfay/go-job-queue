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
	ready  chan *Job
	pool   []*Job
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
		ready:   make(chan *Job, workerCount),
		pool:   []*Job{},
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
			q.wg.Add(-len(q.ready))
			q.wg.Wait()
			q.end <- 1
			return

		case j := <-q.ready:
			go func(job *Job) {
				defer q.wg.Done()
				(*job).Run(q.ctx)
				if q.ctx.Err() != nil {
					return
				}

				q.mu.Lock()
				defer q.mu.Unlock()
				if len(q.pool) == 0 {
					return
				}

				q.wg.Add(1)
				q.ready <- q.pool[0]
				q.pool = q.pool[1:]
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
	if len(q.ready) < cap(q.ready) {
		q.wg.Add(1)
		q.ready <- &job
	} else {
		q.pool = append(q.pool, &job)
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
