# Practice 6 — Go Concurrency

## How to run each solution

```bash
# Problem 1 — Thread-safe maps
go run -race problem1_syncmap.go
go run -race problem1_rwmutex.go

# Problem 2 — Concurrent counter
go run -race problem2.go

# Problem 3 — Fan-In metrics aggregation
go run -race problem3_fanin.go
```

## Problem 1 — Thread-Safe Maps

Two approaches are provided:

| File | Technique |
|------|-----------|
| `problem1_syncmap.go` | `sync.Map` — concurrent-safe map from the standard library |
| `problem1_rwmutex.go` | `sync.RWMutex` wrapping a regular `map[string]int` |

`sync.Map` is best when keys are written once and read many times.  
`sync.RWMutex` gives more control and works well when you need atomic read-modify-write patterns.

## Problem 2 — Concurrent Counter

**One-sentence explanation for team lead:**  
The `counter++` operation is not atomic — it expands to a read, increment, and write, so concurrent goroutines can overwrite each other's updates and the final value is unpredictable.

**Race detection:**
```bash
go run -race problem2.go
```

Two fixes (no channels):
1. `sync.Mutex` — wraps the increment in a critical section
2. `sync/atomic` (`atomic.AddInt64`) — hardware-level atomic increment, no lock needed

## Problem 3 — Fan-In Pattern

`FanIn` accepts a variadic `...<-chan string`, spawns one forwarding goroutine per
channel, and uses a `sync.WaitGroup` to close the merged channel only after every
source has closed. Context cancellation propagates automatically because
`startServer` closes its output channel when `ctx.Done()` fires.
