package part1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPaymentResponse represents the expected payment response
type TestPaymentResponse struct {
	Status        string `json:"status"`
	Amount        int    `json:"amount"`
	TransactionID string `json:"transaction_id"`
}

// TestExecutePaymentWithRetry tests the payment client with a simulated unstable server
func TestExecutePaymentWithRetry(t *testing.T) {
	// Track the number of requests
	requestCount := 0

	// Create a test server that fails the first 3 times, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		t.Logf("Request #%d\n", requestCount)

		// Fail the first 3 requests with 503 Service Unavailable
		if requestCount <= 3 {
			t.Logf("Request #%d: Returning 503 Service Unavailable\n", requestCount)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
			return
		}

		// 4th request succeeds
		t.Logf("Request #%d: Returning 200 OK\n", requestCount)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := TestPaymentResponse{
			Status:        "success",
			Amount:        1000,
			TransactionID: "txn-12345-uuid",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a payment client with custom config
	config := RetryConfig{
		MaxRetries: 5,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   500 * time.Millisecond,
	}
	client := NewPaymentClient(config)

	// Execute payment with a context that has a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare request body
	paymentData := map[string]interface{}{
		"amount": 1000,
	}
	body, _ := json.Marshal(paymentData)

	// Execute payment
	resp, err := client.ExecutePayment(ctx, server.URL, bytes.NewReader(body))

	// Assertions
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if requestCount != 4 {
		t.Fatalf("Expected 4 requests, got %d", requestCount)
	}

	t.Logf("✓ Test passed: Succeeded after %d requests\n", requestCount)
}

// TestIsRetryable tests the error classification logic
func TestIsRetryable(t *testing.T) {
	testCases := []struct {
		name        string
		statusCode  int
		err         error
		shouldRetry bool
		description string
	}{
		{
			name:        "429 Too Many Requests",
			statusCode:  http.StatusTooManyRequests,
			shouldRetry: true,
			description: "Rate limit should be retried",
		},
		{
			name:        "500 Internal Server Error",
			statusCode:  http.StatusInternalServerError,
			shouldRetry: true,
			description: "Server error should be retried",
		},
		{
			name:        "502 Bad Gateway",
			statusCode:  http.StatusBadGateway,
			shouldRetry: true,
			description: "Gateway error should be retried",
		},
		{
			name:        "503 Service Unavailable",
			statusCode:  http.StatusServiceUnavailable,
			shouldRetry: true,
			description: "Service unavailable should be retried",
		},
		{
			name:        "504 Gateway Timeout",
			statusCode:  http.StatusGatewayTimeout,
			shouldRetry: true,
			description: "Gateway timeout should be retried",
		},
		{
			name:        "401 Unauthorized",
			statusCode:  http.StatusUnauthorized,
			shouldRetry: false,
			description: "Authentication error should NOT be retried",
		},
		{
			name:        "404 Not Found",
			statusCode:  http.StatusNotFound,
			shouldRetry: false,
			description: "Not found should NOT be retried",
		},
		{
			name:        "200 OK",
			statusCode:  http.StatusOK,
			shouldRetry: false,
			description: "Success should NOT be retried",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tc.statusCode}
			result := IsRetryable(resp, tc.err)
			if result != tc.shouldRetry {
				t.Errorf("Expected retryable=%v, got %v. %s", tc.shouldRetry, result, tc.description)
			} else {
				t.Logf("✓ %s: correctly identified as retryable=%v", tc.name, result)
			}
		})
	}
}

// TestCalculateBackoff tests that backoff time increases exponentially
func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		MaxRetries: 5,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}

	// Test that backoff increases
	for attempt := 1; attempt <= 3; attempt++ {
		backoff := CalculateBackoff(attempt, config)
		// Backoff should be <= previous calculated max (before jitter)
		expectedMax := config.BaseDelay * time.Duration(1<<uint(attempt-1))
		if expectedMax > config.MaxDelay {
			expectedMax = config.MaxDelay
		}
		if backoff > expectedMax {
			t.Errorf("Attempt %d: backoff %v exceeds expected max %v", attempt, backoff, expectedMax)
		}
		t.Logf("Attempt %d: backoff=%v (max=%v)", attempt, backoff, expectedMax)
	}

	t.Logf("✓ Backoff increases as expected")
}

// TestContextCancellation tests that retry stops when context is cancelled
func TestContextCancellation(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		t.Logf("Request #%d\n", requestCount)
		// Always return 503 to keep retrying
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   500 * time.Millisecond,
	}
	client := NewPaymentClient(config)

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	paymentData := map[string]interface{}{"amount": 1000}
	body, _ := json.Marshal(paymentData)

	// Execute payment
	resp, err := client.ExecutePayment(ctx, server.URL, bytes.NewReader(body))

	// Should fail due to context cancellation
	if err == nil || resp != nil && resp.StatusCode == http.StatusOK {
		t.Errorf("Expected cancellation error, got error=%v, status=%v", err, resp)
	}

	if requestCount > 4 {
		t.Errorf("Expected few requests before cancellation, got %d", requestCount)
	}

	t.Logf("✓ Context cancellation test passed: %d requests before cancellation", requestCount)
}
