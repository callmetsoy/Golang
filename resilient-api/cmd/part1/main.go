package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"resilient-api/part1"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Amount float64 `json:"amount"`
	Card   string  `json:"card"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id"`
	Timestamp     string  `json:"timestamp"`
}

// SimulateUnstableGateway creates a test server that simulates an unstable payment gateway
func SimulateUnstableGateway() *httptest.Server {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		log.Printf("")
		log.Printf("[Gateway] Received request #%d", requestCount)

		// Fail first 3 requests with 503 Service Unavailable
		if requestCount <= 3 {
			log.Printf("[Gateway] Request #%d: Returning 503 Service Unavailable", requestCount)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "service temporarily unavailable",
			})
			return
		}

		// Succeed on 4th request
		log.Printf("[Gateway] Request #%d: Returning 200 OK", requestCount)

		// Simulate some processing time
		time.Sleep(500 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := PaymentResponse{
			Status:        "paid",
			Amount:        1000,
			TransactionID: "txn-" + fmt.Sprintf("%d", requestCount),
			Timestamp:     time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))

	return server
}

func main() {
	log.Println("========================================")
	log.Println("PART 1: RESILIENT HTTP CLIENT")
	log.Println("Demonstrating Retry with Exponential Backoff + Jitter")
	log.Println("========================================")
	log.Println("")

	// Start the simulated unstable gateway
	gateway := SimulateUnstableGateway()
	defer gateway.Close()

	// Create a payment client with retry configuration
	config := part1.RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}

	client := part1.NewPaymentClient(config)

	log.Println("[Client] Executing payment with retry logic...")
	log.Printf("[Client] Configuration: MaxRetries=%d, BaseDelay=%v, MaxDelay=%v",
		config.MaxRetries, config.BaseDelay, config.MaxDelay)
	log.Println("")

	// Create a payment request
	paymentRequest := PaymentRequest{
		Amount: 1000,
		Card:   "****-****-****-1234",
	}

	body, _ := json.Marshal(paymentRequest)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute payment
	startTime := time.Now()
	resp, err := client.ExecutePayment(ctx, gateway.URL, bytes.NewReader(body))
	totalTime := time.Since(startTime)

	log.Printf("")
	log.Println("========================================")
	if err != nil {
		log.Printf("[Client] ✗ Payment failed: %v", err)
		log.Printf("[Client] Total time: %v", totalTime)
	} else {
		log.Printf("[Client] ✓ Payment succeeded!")
		log.Printf("[Client] Status: %d", resp.StatusCode)

		// Read and display response
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var paymentResp PaymentResponse
		json.Unmarshal(respBody, &paymentResp)

		log.Printf("[Client] Response:")
		log.Printf("  - Status: %s", paymentResp.Status)
		log.Printf("  - Amount: $%.2f", paymentResp.Amount)
		log.Printf("  - Transaction ID: %s", paymentResp.TransactionID)
		log.Printf("  - Timestamp: %s", paymentResp.Timestamp)
		log.Printf("[Client] Total time: %v", totalTime)
	}
	log.Println("========================================")
	log.Println("")

	// Demo 2: Show context timeout behavior
	log.Println("========================================")
	log.Println("DEMO 2: Context Timeout Behavior")
	log.Println("========================================")
	log.Println("")

	gateway2 := SimulateUnstableGateway()
	defer gateway2.Close()

	// Create a short timeout context
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()

	log.Println("[Client] Executing payment with 1 second timeout...")
	startTime2 := time.Now()
	resp2, err2 := client.ExecutePayment(ctx2, gateway2.URL, bytes.NewReader(body))
	totalTime2 := time.Since(startTime2)

	log.Printf("")
	log.Println("========================================")
	if err2 != nil {
		log.Printf("[Client] ✓ Request cancelled as expected: %v", err2)
		log.Printf("[Client] Total time: %v", totalTime2)
	} else if resp2 != nil {
		log.Printf("[Client] Received response: %d", resp2.StatusCode)
	}
	log.Println("========================================")
	log.Println("")
}
