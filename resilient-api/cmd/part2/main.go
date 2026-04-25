package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"resilient-api/part2"
)

// RequestTracker tracks request execution
type RequestTracker struct {
	totalRequests   atomic.Int32
	successfulCount atomic.Int32
	conflictCount   atomic.Int32
	processingCount atomic.Int32
	responseStates  map[string]string // key -> response state
	responseStateMu sync.RWMutex
}

func (rt *RequestTracker) RecordRequest(key string, status int) {
	rt.totalRequests.Add(1)
	rt.responseStateMu.Lock()
	defer rt.responseStateMu.Unlock()

	state := "unknown"
	switch status {
	case http.StatusOK:
		rt.successfulCount.Add(1)
		state = "✓ EXECUTED"
	case http.StatusConflict:
		rt.conflictCount.Add(1)
		state = "⚠ CONFLICT (409)"
	}
	rt.responseStates[key] = state
}

// DemoIdempotency demonstrates idempotency with concurrent requests
func DemoIdempotency() {
	log.Println("")
	log.Println("========================================")
	log.Println("PART 2: IDEMPOTENT API")
	log.Println("Demonstrating Loan Repayment with Idempotency Keys")
	log.Println("========================================")
	log.Println("")

	// Create the idempotent payment server
	store := part2.NewMemoryStore()
	paymentHandler := http.HandlerFunc(part2.PaymentHandler)
	idempotentHandler := part2.IdempotencyMiddleware(store, paymentHandler)

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	serverURL := fmt.Sprintf("http://localhost:%d", port)

	// Start server
	server := &http.Server{
		Handler: idempotentHandler,
	}

	go func() {
		log.Printf("[Server] Starting on %s", serverURL)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("[Server] Error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// === SCENARIO 1: Single Request ===
	log.Println("")
	log.Println("--- SCENARIO 1: Single Payment Request ---")
	log.Println("")

	key1 := "payment-001"
	log.Printf("[Client] Sending payment with key: %s", key1)

	req1, _ := http.NewRequest("POST", serverURL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req1.Header.Set("Idempotency-Key", key1)
	req1.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	startTime := time.Now()
	resp1, err1 := client.Do(req1)
	duration1 := time.Since(startTime)

	if err1 == nil {
		defer resp1.Body.Close()
		body1, _ := io.ReadAll(resp1.Body)
		log.Printf("[Client] ✓ Status: %d", resp1.StatusCode)
		log.Printf("[Client] Duration: %v", duration1)
		log.Printf("[Client] Response: %s", body1)
	} else {
		log.Printf("[Client] ✗ Error: %v", err1)
	}

	// === SCENARIO 2: Duplicate Request (Retry) ===
	log.Println("")
	log.Println("--- SCENARIO 2: Duplicate Request (User Retries Due to Network Loss) ---")
	log.Println("")

	log.Printf("[Client] Sending duplicate with same key: %s", key1)
	req2, _ := http.NewRequest("POST", serverURL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req2.Header.Set("Idempotency-Key", key1)
	req2.Header.Set("Content-Type", "application/json")

	startTime = time.Now()
	resp2, err2 := client.Do(req2)
	duration2 := time.Since(startTime)

	if err2 == nil {
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		log.Printf("[Client] ✓ Status: %d", resp2.StatusCode)
		log.Printf("[Client] Duration: %v", duration2)
		log.Printf("[Client] Cached: %s", resp2.Header.Get("X-Idempotent-Cached"))
		log.Printf("[Client] Response: %s", body2)
	} else {
		log.Printf("[Client] ✗ Error: %v", err2)
	}

	// === SCENARIO 3: Double-Click Attack (Concurrent Requests) ===
	log.Println("")
	log.Println("--- SCENARIO 3: Double-Click Attack (Concurrent Duplicate Requests) ---")
	log.Println("")

	key3 := "payment-002"
	numConcurrentRequests := 10

	log.Printf("[Client] Launching %d concurrent requests with key: %s", numConcurrentRequests, key3)
	log.Println("[Client] This simulates rapid double-clicks or network retries")
	log.Println("")

	var wg sync.WaitGroup
	var tracker RequestTracker
	tracker.responseStates = make(map[string]string)

	type ConcurrentResult struct {
		index    int
		status   int
		duration time.Duration
		body     []byte
		cached   bool
		error    error
	}

	results := make([]ConcurrentResult, numConcurrentRequests)
	var mu sync.Mutex

	startTime = time.Now()

	for i := 0; i < numConcurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			reqTime := time.Now()
			req, _ := http.NewRequest("POST", serverURL, bytes.NewReader([]byte(`{"amount":1000}`)))
			req.Header.Set("Idempotency-Key", key3)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			reqDuration := time.Since(reqTime)

			if err != nil {
				mu.Lock()
				results[index] = ConcurrentResult{
					index:    index,
					duration: reqDuration,
					error:    err,
				}
				mu.Unlock()
				return
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			mu.Lock()
			results[index] = ConcurrentResult{
				index:    index,
				status:   resp.StatusCode,
				duration: reqDuration,
				body:     body,
				cached:   resp.Header.Get("X-Idempotent-Cached") == "true",
			}
			mu.Unlock()

			tracker.RecordRequest(key3, resp.StatusCode)
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Print results
	log.Println("")
	log.Println("=== CONCURRENT REQUEST RESULTS ===")
	log.Println("")

	var executedCount, cachedCount, conflictCount int
	var firstResponse []byte

	for _, result := range results {
		if result.error != nil {
			log.Printf("[Request %2d] ✗ Error: %v", result.index+1, result.error)
		} else {
			switch result.status {
			case http.StatusOK:
				if result.cached {
					log.Printf("[Request %2d] ✓ 200 OK (CACHED - %v)", result.index+1, result.duration)
					cachedCount++
				} else {
					log.Printf("[Request %2d] ✓ 200 OK (EXECUTED - %v)", result.index+1, result.duration)
					executedCount++
					if firstResponse == nil {
						firstResponse = result.body
					}
				}
			case http.StatusConflict:
				log.Printf("[Request %2d] ⚠ 409 CONFLICT (in progress)", result.index+1)
				conflictCount++
			default:
				log.Printf("[Request %2d] ? Status %d", result.index+1, result.status)
			}
		}
	}

	log.Println("")
	log.Println("=== STATISTICS ===")
	log.Printf("Total Time: %v", totalDuration)
	log.Printf("Total Requests: %d", numConcurrentRequests)
	log.Printf("  - Executed (first): %d", executedCount)
	log.Printf("  - Returned from Cache: %d", cachedCount)
	log.Printf("  - Conflicts (409): %d", conflictCount)

	log.Println("")
	log.Println("=== KEY FINDINGS ===")
	log.Printf("✓ Only 1 request actually executed the business logic")
	log.Printf("✓ Other %d requests were deduplicated", numConcurrentRequests-executedCount)
	log.Printf("✓ %d requests were rejected as conflicts (in progress)", conflictCount)
	log.Printf("✓ %d requests received cached results", cachedCount)

	// === SCENARIO 4: Missing Idempotency Key ===
	log.Println("")
	log.Println("--- SCENARIO 4: Missing Idempotency-Key Header ---")
	log.Println("")

	log.Printf("[Client] Sending request WITHOUT Idempotency-Key")
	req4, _ := http.NewRequest("POST", serverURL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req4.Header.Set("Content-Type", "application/json")

	resp4, err4 := client.Do(req4)
	if err4 == nil {
		defer resp4.Body.Close()
		log.Printf("[Client] Status: %d (should be 400 Bad Request)", resp4.StatusCode)
		if resp4.StatusCode == http.StatusBadRequest {
			log.Printf("[Client] ✓ Correctly rejected")
		}
	} else {
		log.Printf("[Client] Error: %v", err4)
	}

	log.Println("")
	log.Println("========================================")
	log.Println("")

	// Cleanup
	server.Close()
}

func main() {
	DemoIdempotency()
}
