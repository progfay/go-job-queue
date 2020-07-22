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

func TestJobQueue_Enqueue(t *testing.T) {
	t.Run("Ignore Queue.Enqueue for stopped queue", func(t *testing.T) {
		q := jobqueue.NewQueue(5)
		q.Stop()

		var doneMap sync.Map

		q.Enqueue(&SleepJob{
			id:      0,
			sleep:   0,
			doneMap: &doneMap,
		})
		q.Wait()

		if _, ok := doneMap.Load(0); ok {
			t.Error("Job has been enqueued to stopped Queue")
		}
	})
}

func TestJobQueue_Wait(t *testing.T) {
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

	t.Run("No lock twice call of Queue.Wait", func(t *testing.T) {
		timer := time.NewTimer(time.Second)
		done := make(chan bool)

		go func() {
			q := jobqueue.NewQueue(1)
			defer q.Stop()

			q.Wait()
			q.Wait()
			done <- true
		}()

		select {
		case <-done:
			timer.Stop()
		case <-timer.C:
			t.Error("Timeout Queue.Wait")
		}
	})
}

func TestJobQueue_Stop(t *testing.T) {
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
			t.Error("Job enqueued to queue that is stopped is running")
		}
	})

	t.Run("No lock twice call of Queue.Stop", func(t *testing.T) {
		timer := time.NewTimer(time.Second)
		done := make(chan bool)

		go func() {
			q := jobqueue.NewQueue(1)

			q.Stop()
			q.Stop()
			done <- true
		}()

		select {
		case <-done:
			timer.Stop()
		case <-timer.C:
			t.Error("Timeout Queue.Stop")
		}
	})
}

func TestJobQueue_WithContext(t *testing.T) {
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
			t.Error("Job enqueued to queue whose context is cancelled is running")
		}
	})
}
