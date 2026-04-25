package part2

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CachedResponse represents a cached response from a previously processed request
type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
	CreatedAt  time.Time
}

// IdempotencyStore defines the interface for storing idempotency keys and their results
type IdempotencyStore interface {
	// Get retrieves a cached response by key
	Get(ctx context.Context, key string) (*CachedResponse, bool)

	// StartProcessing attempts to mark a key as "processing"
	// Returns true if this is the first request with this key
	StartProcessing(ctx context.Context, key string) bool

	// Finish marks a key as completed and stores the result
	Finish(ctx context.Context, key string, status int, body []byte) error

	// Clean removes expired entries (if applicable)
	Clean(ctx context.Context) error
}

// MemoryStore is an in-memory implementation of IdempotencyStore
// Suitable for single-server deployments and testing
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]*CachedResponse
}

// NewMemoryStore creates a new in-memory idempotency store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*CachedResponse),
	}
}

// Get retrieves a cached response by key
func (m *MemoryStore) Get(ctx context.Context, key string) (*CachedResponse, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resp, exists := m.data[key]
	if exists {
		fmt.Printf("[Store] Get: Found cached response for key %s (completed=%v)\n", key, resp.Completed)
	}
	return resp, exists
}

// StartProcessing attempts to mark a key as "processing"
// Returns true if this is the first request with this key
func (m *MemoryStore) StartProcessing(ctx context.Context, key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; exists {
		fmt.Printf("[Store] StartProcessing: Key %s already exists\n", key)
		return false
	}

	// Insert an "empty" record, marking that the request is in progress
	m.data[key] = &CachedResponse{
		Completed: false,
		CreatedAt: time.Now(),
	}
	fmt.Printf("[Store] StartProcessing: Marked key %s as processing\n", key)
	return true
}

// Finish marks a key as completed and stores the result
func (m *MemoryStore) Finish(ctx context.Context, key string, status int, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if resp, exists := m.data[key]; exists {
		resp.StatusCode = status
		resp.Body = make([]byte, len(body))
		copy(resp.Body, body)
		resp.Completed = true
		fmt.Printf("[Store] Finish: Marked key %s as completed with status %d\n", key, status)
	} else {
		// This shouldn't happen, but handle it gracefully
		m.data[key] = &CachedResponse{
			StatusCode: status,
			Body:       make([]byte, len(body)),
			Completed:  true,
			CreatedAt:  time.Now(),
		}
		copy(m.data[key].Body, body)
	}
	return nil
}

// Clean removes expired entries
func (m *MemoryStore) Clean(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For now, we keep all entries. In production, you might want to
	// clean up entries older than a certain duration
	fmt.Printf("[Store] Clean: No cleanup needed for MemoryStore\n")
	return nil
}

// WaitForCompletion waits for a key to be completed (up to a timeout)
func (m *MemoryStore) WaitForCompletion(ctx context.Context, key string, timeout time.Duration) (*CachedResponse, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[Store] WaitForCompletion: Timeout waiting for key %s\n", key)
			return nil, false
		case <-ticker.C:
			resp, exists := m.Get(ctx, key)
			if exists && resp.Completed {
				fmt.Printf("[Store] WaitForCompletion: Key %s completed successfully\n", key)
				return resp, true
			}
		}
	}
}
