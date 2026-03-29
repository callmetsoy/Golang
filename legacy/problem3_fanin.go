package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
				out <- fmt.Sprintf("[%s] metric: %d", name, rand.Intn(100))
			}
		}
	}()
	return out
}

func FanIn(ctx context.Context, channels ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup

	forward := func(ch <-chan string) {
		defer wg.Done()
		for val := range ch {
			select {
			case <-ctx.Done():
				return
			case merged <- val:
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch1 := startServer(ctx, "Alpha")
	ch2 := startServer(ctx, "Beta")
	ch3 := startServer(ctx, "Gamma")

	ch4 := FanIn(ctx, ch1, ch2, ch3)

	for val := range ch4 {
		fmt.Println(val)
	}
}
