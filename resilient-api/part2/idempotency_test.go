package part2

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestIdempotencyBasic(t *testing.T) {
	store := NewMemoryStore()
	handler := http.HandlerFunc(PaymentHandler)
	idempotentHandler := IdempotencyMiddleware(store, handler)
	server := httptest.NewServer(idempotentHandler)
	defer server.Close()

	key := "test-key-1"

	t.Logf("Sending first request with key: %s\n", key)
	req1, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req1.Header.Set("Idempotency-Key", key)
	req1.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp1, _ := client.Do(req1)
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	t.Logf("First response status: %d\n", resp1.StatusCode)
	t.Logf("First response body: %s\n", body1)

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp1.StatusCode)
	}

	t.Logf("Sending second request with same key: %s\n", key)
	time.Sleep(100 * time.Millisecond)
	req2, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req2.Header.Set("Idempotency-Key", key)
	req2.Header.Set("Content-Type", "application/json")

	resp2, _ := client.Do(req2)
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	t.Logf("Second response status: %d\n", resp2.StatusCode)
	t.Logf("Second response body: %s\n", body2)

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp2.StatusCode)
	}

	if resp2.Header.Get("X-Idempotent-Cached") != "true" {
		t.Fatalf("Expected cached response, but X-Idempotent-Cached header not found")
	}

	if string(body1) != string(body2) {
		t.Fatalf("Expected identical responses")
	}

	t.Logf("Test passed: Idempotency works correctly\n")
}

func TestMissingIdempotencyKey(t *testing.T) {
	store := NewMemoryStore()
	handler := http.HandlerFunc(PaymentHandler)
	idempotentHandler := IdempotencyMiddleware(store, handler)
	server := httptest.NewServer(idempotentHandler)
	defer server.Close()

	t.Logf("Sending request without Idempotency-Key header\n")
	req, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	t.Logf("Response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request, got %d", resp.StatusCode)
	}

	t.Logf("Test passed: Missing header correctly returns 400\n")
}

func TestConcurrentRequests(t *testing.T) {
	store := NewMemoryStore()
	handler := http.HandlerFunc(PaymentHandler)
	idempotentHandler := IdempotencyMiddleware(store, handler)
	server := httptest.NewServer(idempotentHandler)
	defer server.Close()

	key := "test-concurrent-key"
	numRequests := 10

	type RequestResult struct {
		Status int
		Body   []byte
		Error  error
	}

	results := make([]RequestResult, numRequests)
	var wg sync.WaitGroup
	var mu sync.Mutex

	t.Logf("Launching %d concurrent requests with same key\n", numRequests)
	startTime := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			req, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"amount":1000}`)))
			req.Header.Set("Idempotency-Key", key)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)

			var body []byte
			if resp != nil {
				body, _ = io.ReadAll(resp.Body)
				resp.Body.Close()

				mu.Lock()
				results[index] = RequestResult{
					Status: resp.StatusCode,
					Body:   body,
					Error:  err,
				}
				mu.Unlock()

				fmt.Printf("[Request %d] Status: %d, Cached: %s\n",
					index+1, resp.StatusCode, resp.Header.Get("X-Idempotent-Cached"))
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	var successCount, conflictCount, errorCount int
	var firstResponse []byte

	for i, result := range results {
		if result.Error != nil {
			errorCount++
			t.Logf("Request %d: Error - %v\n", i+1, result.Error)
		} else if result.Status == http.StatusOK {
			successCount++
			if firstResponse == nil {
				firstResponse = result.Body
			}
			if string(result.Body) != string(firstResponse) {
				t.Errorf("Request %d: Response differs from first response", i+1)
			}
		} else if result.Status == http.StatusConflict {
			conflictCount++
		} else {
			t.Logf("Request %d: Unexpected status - %d\n", i+1, result.Status)
		}
	}

	t.Logf("\n=== Results ===")
	t.Logf("Total Time: %v\n", totalTime)
	t.Logf("Successful: %d\n", successCount)
	t.Logf("Conflicts (409): %d\n", conflictCount)
	t.Logf("Errors: %d\n", errorCount)

	if successCount < 1 {
		t.Fatalf("Expected at least 1 successful response")
	}

	if conflictCount < 1 && numRequests > 1 {
		t.Logf("Warning: Expected some conflict responses, but got none\n")
	}

	if errorCount > 0 {
		t.Fatalf("Unexpected errors: %d", errorCount)
	}

	t.Logf("Test passed: Concurrent requests handled correctly\n")
}

func TestDifferentKeys(t *testing.T) {
	store := NewMemoryStore()
	handler := http.HandlerFunc(PaymentHandler)
	idempotentHandler := IdempotencyMiddleware(store, handler)
	server := httptest.NewServer(idempotentHandler)
	defer server.Close()

	key1 := "test-key-1"
	key2 := "test-key-2"

	t.Logf("Sending first request with key: %s\n", key1)
	req1, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"amount":1000}`)))
	req1.Header.Set("Idempotency-Key", key1)
	req1.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp1, _ := client.Do(req1)
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	t.Logf("First request status: %d\n", resp1.StatusCode)

	t.Logf("Sending request with different key: %s\n", key2)
	req2, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"amount":2000}`)))
	req2.Header.Set("Idempotency-Key", key2)
	req2.Header.Set("Content-Type", "application/json")

	resp2, _ := client.Do(req2)
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	t.Logf("Second request status: %d\n", resp2.StatusCode)

	if resp1.StatusCode != http.StatusOK || resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected both requests to succeed")
	}

	if string(body1) == string(body2) {
		t.Fatalf("Expected different responses for different keys")
	}

	t.Logf("Test passed: Different keys processed independently\n")
}
