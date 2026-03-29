package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func counterWithMutex() {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println("[Mutex] Counter:", counter)
}

func counterWithAtomic() {
	var counter int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()
	fmt.Println("[Atomic] Counter:", counter)
}

func main() {
	counterWithMutex()
	counterWithAtomic()
}
