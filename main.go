package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/progfay/go-job-queue/queue"
)

func main() {
	ctx := context.Background()
	q := queue.NewQueue(ctx, 10)

	for i := 0; i < 100; i++ {
		j := queue.NewJob(func(args ...interface{}) {
			r := rand.Intn(1000)
			time.Sleep(time.Duration(r) * time.Millisecond)
			fmt.Println(r, args)
		}, i)
		q.Add(j)
	}

	q.Start()
}
