package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/progfay/go-job-queue/queue"
)

func main() {
	q := queue.NewQueue(10)
	defer q.Stop()

	for i := 0; i < 100; i++ {
		j := queue.NewJob(func(args ...interface{}) {
			fmt.Println("start", args)
			r := rand.Intn(1000)
			time.Sleep(time.Duration(r) * time.Millisecond)
			fmt.Println(r, args)
		}, i)
		q.Add(j)
	}

	q.Wait()
	fmt.Println("wait end")

	time.Sleep(time.Second)

	q.Add(queue.NewJob(func(args ...interface{}) {
		fmt.Println("hoge")
	}))

	q.Wait()
}
