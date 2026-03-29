package main

import (
	"fmt"
	"sync"
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		m: make(map[string]int),
	}
}

func (s *SafeMap) Write(key string, value int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *SafeMap) Read(key string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.m[key]
	return val, ok
}

func main() {
	safeMap := NewSafeMap()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			safeMap.Write("key", key)
		}(i)
	}

	wg.Wait()

	value, ok := safeMap.Read("key")
	if ok {
		fmt.Printf("Value: %d\n", value)
	}
}
