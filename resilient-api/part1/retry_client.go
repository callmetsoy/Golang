package part1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}
}

type PaymentClient struct {
	httpClient *http.Client
	config     RetryConfig
}

func NewPaymentClient(config RetryConfig) *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		config:     config,
	}
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var timeoutErr interface{ Timeout() bool }
		if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
			return true
		}

		var netErr interface{ Temporary() bool }
		if errors.As(err, &netErr) && netErr.Temporary() {
			return true
		}

		return true
	}

	if resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return true
		case http.StatusInternalServerError:
			return true
		case http.StatusBadGateway:
			return true
		case http.StatusServiceUnavailable:
			return true
		case http.StatusGatewayTimeout:
			return true
		case http.StatusUnauthorized:
			return false
		case http.StatusNotFound:
			return false
		}
	}

	if resp != nil && resp.StatusCode < 400 {
		return false
	}

	if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return false
	}

	return false
}

func CalculateBackoff(attempt int, config RetryConfig) time.Duration {
	if attempt == 0 {
		return 0
	}

	backoffTime := config.BaseDelay * time.Duration(math.Pow(2, float64(attempt-1)))

	if backoffTime > config.MaxDelay {
		backoffTime = config.MaxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(backoffTime)))

	return jitter
}

func (pc *PaymentClient) ExecutePayment(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt < pc.config.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		}

		fmt.Printf("\n=== Attempt %d ===\n", attempt+1)

		req, err := http.NewRequestWithContext(ctx, "POST", url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := pc.httpClient.Do(req)
		lastResp = resp

		if !IsRetryable(resp, err) {
			if err != nil {
				lastErr = err
				return nil, fmt.Errorf("request failed (not retryable): %w", err)
			}
			fmt.Printf("Request succeeded with status %d\n", resp.StatusCode)
			return resp, nil
		}

		lastErr = err

		if attempt == pc.config.MaxRetries-1 {
			fmt.Printf("Max retries reached (%d attempts)\n", pc.config.MaxRetries)
			if lastResp != nil {
				return lastResp, fmt.Errorf("max retries exceeded, last status: %d", lastResp.StatusCode)
			}
			return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
		}

		backoff := CalculateBackoff(attempt+1, pc.config)
		fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, backoff)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			fmt.Printf("Context cancelled during backoff wait\n")
			return nil, fmt.Errorf("request cancelled during retry: %w", ctx.Err())
		}
	}

	return lastResp, lastErr
}
