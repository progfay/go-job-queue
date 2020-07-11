package queue

import (
	"context"
	"fmt"
	"sync"
)

type Job struct {
	handler func()
}

func NewJob(handler func()) Job {
	return Job{handler: handler}
}

type Queue struct {
	IsRunning bool
	queue     []Job
	mu        sync.Mutex
	waiting   chan int
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    func()
}

func NewQueue(ctx context.Context, maxJobCount int) Queue {
	waiting := make(chan int, maxJobCount)
	for i := 0; i < maxJobCount; i++ {
		waiting <- i
	}

	childCtx, cancel := context.WithCancel(ctx)

	return Queue{
		IsRunning: false,
		queue:     make([]Job, 0),
		mu:        sync.Mutex{},
		waiting:   waiting,
		wg:        sync.WaitGroup{},
		ctx:       childCtx,
		cancel:    cancel,
	}
}

func (q *Queue) Start() {
	if q.IsRunning {
		return
	}

	q.IsRunning = true

	for {
		select {
		case <-q.ctx.Done():
			q.wg.Wait()
			q.IsRunning = false
			return
		case w := <-q.waiting:
			func() {
				q.mu.Lock()
				defer q.mu.Unlock()

				if len(q.queue) == 0 {
					q.cancel()
					return
				}

				job := q.queue[0]
				q.queue = q.queue[1:]

				q.wg.Add(1)
				go func(job Job) {
					fmt.Printf("start %v\n", w)
					job.handler()
					fmt.Printf("end   %v\n", w)
					q.waiting <- w
					q.wg.Done()
				}(job)
			}()
		}
	}
}

func (q *Queue) Add(job Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, job)
}
