package part2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func IdempotencyMiddleware(store IdempotencyStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" && r.Method != "PUT" && r.Method != "PATCH" {
			fmt.Printf("[Middleware] %s request (no idempotency needed)\n", r.Method)
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			fmt.Printf("[Middleware] Missing Idempotency-Key header\n")
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		fmt.Printf("[Middleware] Processing request with Idempotency-Key: %s\n", key)

		if cached, exists := store.Get(r.Context(), key); exists {
			if cached.Completed {
				fmt.Printf("[Middleware] Key already completed, returning cached response\n")
				w.Header().Set("X-Idempotent-Cached", "true")
				w.Header().Set("X-Idempotent-Date", cached.CreatedAt.Format(time.RFC3339))
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			} else {
				fmt.Printf("[Middleware] Duplicate request in progress, returning 409 Conflict\n")
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
				return
			}
		}

		if !store.StartProcessing(r.Context(), key) {
			fmt.Printf("[Middleware] Failed to reserve key, checking if now completed\n")
			if cached, exists := store.Get(r.Context(), key); exists && cached.Completed {
				fmt.Printf("[Middleware] Request just completed, returning cached response\n")
				w.Header().Set("X-Idempotent-Cached", "true")
				w.Header().Set("X-Idempotent-Date", cached.CreatedAt.Format(time.RFC3339))
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				fmt.Printf("[Middleware] Request still processing, returning 409 Conflict\n")
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}

		fmt.Printf("[Middleware] First request for this key, executing handler\n")

		recorder := httptest.NewRecorder()

		next.ServeHTTP(recorder, r)

		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}

		statusCode := recorder.Code
		body := recorder.Body.Bytes()

		if err := store.Finish(r.Context(), key, statusCode, body); err != nil {
			fmt.Printf("[Middleware] Error saving result: %v\n", err)
		}

		w.Header().Set("X-Idempotent-Executed", "true")
		w.WriteHeader(statusCode)
		w.Write(body)

		fmt.Printf("[Middleware] Response sent: %d\n", statusCode)
	})
}

func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[Handler] Processing payment request\n")

	fmt.Printf("[Handler] Simulating 2 second payment processing...\n")
	time.Sleep(2 * time.Second)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	transactionID := fmt.Sprintf("txn-%d", time.Now().UnixNano())

	response := fmt.Sprintf(`{"status":"paid","amount":1000,"transaction_id":"%s"}`, transactionID)
	fmt.Fprintf(w, response)

	fmt.Printf("[Handler] Payment completed: %s\n", transactionID)
}

func CreateIdempotentPaymentServer() (*http.Server, IdempotencyStore) {
	store := NewMemoryStore()

	paymentHandlerFunc := http.HandlerFunc(PaymentHandler)

	handler := IdempotencyMiddleware(store, paymentHandlerFunc)

	testServer := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	return testServer, store
}
