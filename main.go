package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/progfay/go-job-queue/job"
	"github.com/progfay/go-job-queue/queue"
)

func main() {
	q := queue.NewQueue(10)
	defer q.Stop()

	for i := 0; i < 100; i++ {
		j := job.NewJob(func(args ...interface{}) {
			fmt.Println("start:", args[0])
			r := rand.Intn(1000)
			time.Sleep(time.Duration(r) * time.Millisecond)
			fmt.Println("end  :", args[0])
		}, i)
		q.Add(j)
	}

	// q.Wait()
	// fmt.Println("wait end")

	time.Sleep(time.Second * 1)
	q.Stop()

	time.Sleep(time.Second * 1)

	q.Add(job.NewJob(func(args ...interface{}) {
		fmt.Println("hoge")
	}))

	q.Wait()
}
