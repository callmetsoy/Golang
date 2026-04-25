# Resilient API: Retries & Idempotency in Go

A comprehensive implementation of fault-tolerant API patterns in Go, featuring:
- **Part 1**: Resilient HTTP Client with exponential backoff, jitter, and context support
- **Part 2**: Idempotent API with duplicate request detection using idempotency keys

## Overview

This project demonstrates production-ready patterns for building reliable distributed systems that can withstand:
- Temporary network failures
- Service unavailability
- Rate limiting
- Concurrent duplicate requests
- Network timeouts and retries

## Project Structure

```
resilient-api/
├── part1/                          # Resilient HTTP Client
│   ├── retry_client.go            # Core retry logic with backoff + jitter
│   └── retry_client_test.go       # Comprehensive tests
├── part2/                          # Idempotent API
│   ├── idempotency_store.go       # In-memory idempotency store
│   ├── idempotency_middleware.go  # HTTP middleware for idempotency
│   └── idempotency_test.go        # Idempotency tests
├── cmd/
│   ├── part1/main.go              # Part 1 demonstration
│   └── part2/main.go              # Part 2 demonstration
├── docker-compose.yml             # Redis/Postgres for production
├── go.mod                          # Go module definition
└── README.md                       # This file
```

## Part 1: Resilient HTTP Client

### Features

- **Error Classification**: Distinguishes between temporary and permanent errors
- **Exponential Backoff**: Starts with short delays, increases exponentially
- **Full Jitter**: Adds randomness to prevent thundering herd effect
- **Context Support**: Respects context cancellation and timeouts
- **Configurable**: Customize max retries, base delay, and max delay

### Key Components

#### `IsRetryable(resp *http.Response, err error) bool`
Determines if an error should be retried:
- **Retryable** (2xx):
  - 429 Too Many Requests
  - 500 Internal Server Error
  - 502 Bad Gateway
  - 503 Service Unavailable
  - 504 Gateway Timeout
  - Network timeouts and temporary errors
  
- **Non-retryable** (permanent):
  - 401 Unauthorized
  - 404 Not Found
  - Most 4xx errors

#### `CalculateBackoff(attempt int, config RetryConfig) time.Duration`
Implements exponential backoff with full jitter:
```
backoff = min(baseDelay * 2^attempt, maxDelay)
jitter = random(0, backoff)
return jitter
```

#### `ExecutePayment(ctx context.Context, url string, body io.Reader) (*http.Response, error)`
Main retry loop that:
1. Checks context for cancellation before each attempt
2. Executes HTTP request
3. Classifies error with `IsRetryable()`
4. Calculates backoff with jitter
5. Respects context timeout during wait

### Running Part 1

```bash
# Install dependencies
go mod download

# Run the demo
go run cmd/part1/main.go

# Run tests
go test -v ./part1/...

# Run tests with race detector
go test -race -v ./part1/...
```

### Part 1 Output Example

```
========================================
PART 1: RESILIENT HTTP CLIENT
Demonstrating Retry with Exponential Backoff + Jitter
========================================

[Client] Executing payment with retry logic...
[Client] Configuration: MaxRetries=5, BaseDelay=500ms, MaxDelay=5s

[Gateway] Received request #1
[Gateway] Request #1: Returning 503 Service Unavailable
[Middleware] Status 503 (Service Unavailable) - retryable
[Client] Attempt 1 failed, waiting ~245ms (backoff 500ms) before next retry...

[Gateway] Received request #2
[Gateway] Request #2: Returning 503 Service Unavailable
[Client] Attempt 2 failed, waiting ~892ms (backoff 1s) before next retry...

[Gateway] Received request #3
[Gateway] Request #3: Returning 503 Service Unavailable
[Client] Attempt 3 failed, waiting ~1500ms (backoff 2s) before next retry...

[Gateway] Received request #4
[Gateway] Request #4: Returning 200 OK
[Client] ✓ Payment succeeded!
```

## Part 2: Idempotent API

### Features

- **Duplicate Detection**: Recognizes and deduplicates concurrent requests
- **Request Caching**: Stores results of completed operations
- **In-Progress Handling**: Returns 409 Conflict for concurrent duplicates
- **Atomic Operations**: Thread-safe with proper synchronization
- **Extensible Storage**: Easy to swap Memory store for Redis/Database

### Key Concepts

#### Idempotency Key
A unique identifier (UUID v4) sent by the client in the `Idempotency-Key` header.
The server uses this to:
1. Recognize duplicate requests
2. Return cached results instead of re-executing
3. Prevent double-charging, duplicate records, etc.

#### Request Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   Idempotency Middleware                     │
├─────────────────────────────────────────────────────────────┤
│ 1. Check for Idempotency-Key header                          │
│    ├─ Missing → 400 Bad Request                              │
│    └─ Present → Continue                                     │
│                                                              │
│ 2. Check if key exists in store                              │
│    ├─ Completed → Return cached result (200 OK)             │
│    ├─ Processing → Return 409 Conflict                      │
│    └─ Not found → Continue                                  │
│                                                              │
│ 3. Mark key as "processing"                                  │
│    ├─ Success → Execute handler and save result             │
│    └─ Failure → Another request beat us, check store again  │
└─────────────────────────────────────────────────────────────┘
```

#### State Transitions

```
Key doesn't exist
    ↓
StartProcessing() → "processing"
    ↓
Execute business logic (can take time)
    ↓
Finish() → "completed" (with status + body)
    ↓
Return to client
```

### Components

#### `IdempotencyStore` Interface
Defines the contract for idempotency storage:
```go
type IdempotencyStore interface {
    Get(ctx context.Context, key string) (*CachedResponse, bool)
    StartProcessing(ctx context.Context, key string) bool
    Finish(ctx context.Context, key string, status int, body []byte) error
    Clean(ctx context.Context) error
}
```

#### `MemoryStore` Implementation
Thread-safe in-memory store suitable for:
- Development and testing
- Single-server deployments
- Small traffic volumes

#### `IdempotencyMiddleware`
HTTP middleware that intercepts requests and implements the idempotency logic.

### Running Part 2

```bash
# Run the demo
go run cmd/part2/main.go

# Run tests
go test -v ./part2/...

# Run tests with race detector
go test -race -v ./part2/...
```

### Part 2 Output Example

```
========================================
PART 2: IDEMPOTENT API
Demonstrating Loan Repayment with Idempotency Keys
========================================

--- SCENARIO 1: Single Payment Request ---

[Client] Sending payment with key: payment-001
[Middleware] Processing request with Idempotency-Key: payment-001
[Middleware] First request for this key, executing handler
[Handler] Processing payment request
[Handler] Simulating 2 second payment processing...
[Handler] Payment completed: txn-1710345678901234567
[Middleware] Response sent: 200
[Client] ✓ Status: 200
[Client] Response: {"status":"paid","amount":1000,"transaction_id":"txn-1710345678901234567"}

--- SCENARIO 2: Duplicate Request (User Retries Due to Network Loss) ---

[Client] Sending duplicate with same key: payment-001
[Middleware] Processing request with Idempotency-Key: payment-001
[Middleware] Key already completed, returning cached response
[Client] ✓ Status: 200
[Client] Cached: true
[Client] Response: {"status":"paid","amount":1000,"transaction_id":"txn-1710345678901234567"}

--- SCENARIO 3: Double-Click Attack (Concurrent Duplicate Requests) ---

[Client] Launching 10 concurrent requests with key: payment-002
[Client] This simulates rapid double-clicks or network retries

[Request  1] ✓ 200 OK (EXECUTED - 2.001s)
[Request  2] ⚠ 409 CONFLICT (in progress)
[Request  3] ⚠ 409 CONFLICT (in progress)
[Request  4] ✓ 200 OK (CACHED - 15ms)
[Request  5] ⚠ 409 CONFLICT (in progress)
...

=== STATISTICS ===
Total Time: 2.345s
Total Requests: 10
  - Executed (first): 1
  - Returned from Cache: 5
  - Conflicts (409): 4

=== KEY FINDINGS ===
✓ Only 1 request actually executed the business logic
✓ Other 9 requests were deduplicated
✓ 4 requests were rejected as conflicts (in progress)
✓ 5 requests received cached results
```

## Running Tests with Race Detector

The race detector helps identify concurrency issues:

```bash
# Test Part 1 with race detection
go test -race -v ./part1/...

# Test Part 2 with race detection
go test -race -v ./part2/...

# Test everything
go test -race -v ./...
```

The `-race` flag:
- Instruments the binary to detect data races
- Essential for concurrent code verification
- Adds overhead but catches subtle bugs
- Should be run before committing

## Production Considerations

### Storage Backends

The `IdempotencyStore` interface can be implemented with:

1. **PostgreSQL** (Recommended for production):
```sql
CREATE TABLE idempotency_keys (
    key UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    response_code INT,
    response_body JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

2. **Redis** (For distributed caching):
```go
// Use Redis with TTL
// SET idempotency:<key> "processing" EX 60 NX
// UPDATE to final result when done
```

### Key Expiration

- **"processing" state**: 60 seconds (timeout for hung operations)
- **"completed" state**: 24 hours (allow retries within a day)

### Rate Limiting

Combine with rate limiting middleware:
- Per-user limits
- Per-IP limits
- Exponential backoff for retries

## Performance Metrics

### Part 1: Resilient HTTP Client
- **Success Rate**: 100% with automatic retries
- **Throughput**: No significant impact from retry logic
- **Latency**: 
  - No-failure case: network latency + handler time
  - With retries: includes exponential backoff + jitter

### Part 2: Idempotency
- **Overhead**: Minimal (mostly mutex lock contention)
- **Cache Hit Rate**: Near 100% for duplicate requests within TTL
- **Concurrency**: Thread-safe with proper synchronization

## Common Patterns

### Pattern 1: Automatic Retries on Failure
```go
config := part1.RetryConfig{
    MaxRetries: 5,
    BaseDelay:  500 * time.Millisecond,
    MaxDelay:   5 * time.Second,
}
client := part1.NewPaymentClient(config)

resp, err := client.ExecutePayment(ctx, url, body)
```

### Pattern 2: Ensuring Idempotency
```go
// Client side: Generate and reuse idempotency key
key := uuid.New().String()

// Retry loop
for attempt := 0; attempt < maxAttempts; attempt++ {
    req.Header.Set("Idempotency-Key", key)
    resp, err := client.Do(req)
    if err == nil && resp.StatusCode < 500 {
        break
    }
    // Wait and retry
}
```

### Pattern 3: Combined: Retries + Idempotency
```go
// For payment operations
key := uuid.New().String()
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

req := part1.NewRetryConfig()
client := part1.NewPaymentClient(req)
resp, err := client.ExecutePayment(ctx, url, body)
// Server checks Idempotency-Key header for deduplication
```

## Recommended Reading

- Retries in Go (provided materials)
- How to Make Idempotent APIs in Go (provided materials)
- [AWS: Implement idempotency patterns](https://aws.amazon.com/blogs/architecture/)
- [HTTP Status Codes](https://httpwg.org/specs/rfc9110.html)

## Bibliography

1. **Exponential Backoff and Jitter**
   - "Exponential Backoff And Jitter" - AWS Architecture Blog
   
2. **Idempotency**
   - RFC 9110: HTTP Semantics
   - Stripe API: Idempotent Requests
   
3. **Distributed Systems**
   - "Designing Data-Intensive Applications" by Martin Kleppmann
   - "Release It!" by Michael Nygard (on retry strategies and failure modes)

## Testing Strategy

### Unit Tests
- `IsRetryable()` with various status codes
- `CalculateBackoff()` exponential growth
- Context cancellation handling

### Integration Tests
- Full retry flow with simulated server
- Concurrent request handling
- Cache hit verification

### Race Detection
- `go test -race ./...`
- Detects data races in concurrent code

## Development

### Prerequisites
- Go 1.21 or higher
- Docker (optional, for Redis/Postgres)

### Setup
```bash
# Clone the repo
cd resilient-api

# Download dependencies
go mod download

# Run all tests with race detection
go test -race -v ./...

# Run specific part
go run cmd/part1/main.go
go run cmd/part2/main.go
```

### Common Issues

1. **"Address already in use"**
   - Another process is using the port
   - Kill the process or use a different port

2. **Race detector warnings**
   - Indicates unsafe concurrent access
   - Add synchronization (mutex, channels)
   - Review the provided implementations

3. **Slow tests**
   - Tests include `time.Sleep()` for realism
   - This is intentional to demonstrate backoff behavior
   - Can be sped up with shorter durations

## Key Takeaways

1. **Always classify errors**: Not all errors warrant a retry
2. **Use exponential backoff + jitter**: Prevents thundering herd
3. **Respect context**: Never ignore cancellation signals
4. **Implement idempotency for critical operations**: Essential for distributed systems
5. **Atomic operations**: Always check and write atomically
6. **Test with race detector**: Catch concurrency bugs early

## License

MIT License - Feel free to use in your projects

## Author

Developed for distributed systems education in Go

---

**Pro Tips**:
- Always use idempotency keys for payment-related operations
- Combine retries with idempotency for ultimate reliability
- Monitor retry rates in production
- Adjust backoff parameters based on your service characteristics
- Use structured logging for debugging retry behavior
