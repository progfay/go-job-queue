package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/progfay/go-job-queue/queue"
)

type SleepJob struct {
	id    int
	sleep time.Duration
}

func (sj *SleepJob) Run(ctx context.Context) {
	fmt.Println("start:", sj.id)
	time.Sleep(sj.sleep)
	fmt.Println("end  :", sj.id)
}

func main() {
	q := queue.NewQueue(10)
	defer q.Stop()

	for i := 0; i < 100; i++ {
		r := rand.Intn(1000)
		j := &SleepJob{
			id:    i,
			sleep: time.Duration(r) * time.Millisecond,
		}
		q.Add(j)
	}

	q.Wait()
}
