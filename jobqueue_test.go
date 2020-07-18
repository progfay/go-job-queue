package jobqueue_test

import (
	"context"

	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/progfay/jobqueue"
)

type SleepJob struct {
	id      int
	sleep   time.Duration
	dest    *[]int64
	doneMap *sync.Map
}

func (sj *SleepJob) Run(ctx context.Context) {
	t := time.NewTimer(sj.sleep)
	select {
	case <-t.C:
	case <-ctx.Done():
		return
	}
	sj.doneMap.Store(sj.id, true)
}

func TestJobQueue(t *testing.T) {
	t.Run("Queue.Wait should wait all enqueued jobs", func(t *testing.T) {
		jobCount := 50

		q := jobqueue.NewQueue(30)
		defer q.Stop()

		var doneMap sync.Map

		for i := 0; i < jobCount; i++ {
			q.Enqueue(&SleepJob{
				id:      i,
				sleep:   time.Duration(rand.Intn(10)) * time.Millisecond,
				doneMap: &doneMap,
			})
		}

		q.Wait()

		for i := 0; i < jobCount; i++ {
			if done, ok := doneMap.Load(i); ok != true || done.(bool) != true {
				t.Fail()
			}
		}
	})

	t.Run("Queue.Stop should stop queue immediately", func(t *testing.T) {
		q := jobqueue.NewQueue(1)

		var doneMap sync.Map

		q.Enqueue(&SleepJob{
			id:      0,
			sleep:   50 * time.Millisecond,
			doneMap: &doneMap,
		})

		q.Stop()

		q.Enqueue(&SleepJob{
			id:      1,
			sleep:   50 * time.Millisecond,
			doneMap: &doneMap,
		})

		if _, ok := doneMap.Load(1); ok == true {
			t.Errorf("Job enqueued to queue that is stopped is running")
		}
	})

	t.Run("Queue.WithContext should return contexted job queue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		q := jobqueue.WithContext(ctx, 1)
		defer q.Stop()

		var doneMap sync.Map

		q.Enqueue(&SleepJob{
			id:      0,
			sleep:   10 * time.Millisecond,
			doneMap: &doneMap,
		})

		cancel()
		q.Wait()

		q.Enqueue(&SleepJob{
			id:      1,
			sleep:   10 * time.Millisecond,
			doneMap: &doneMap,
		})

		if _, ok := doneMap.Load(1); ok == true {
			t.Errorf("Job enqueued to queue whose context is cancelled is running")
		}
	})
}
