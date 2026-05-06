# Custom API Gateway in Go — Learning & Build Plan

> **Goal:** Build a production-grade API gateway from scratch in Go to gain core systems engineering knowledge.  
> **Pace:** Minimum 1 hour/day. Each "Day" = one focused 1-hour session.  
> **How to use:** Open this file at the start of each session. Find your current day (first `[ ]`). Do the work. Mark it `[x]` when done. Ask Copilot: *"I'm on Day X, help me with [topic]"*.  
> **Reference code:** The `internal/`, `cmd/`, `examples/` folders contain reference implementations. **Try to write it yourself first**, then compare.

---

## Current Progress

- **Current Day:** Day 1
- **Started:** 2026-05-06

---

## WEEK 1-2: Prerequisites — Go & Networking Foundations

> Don't skip this. The gateway will use every one of these concepts.

### Day 1: Go Basics Refresher & Project Setup `[ ]`
**Learn:**
- How Go modules work (`go mod init`, `go.mod`, `go.sum`)
- Project layout convention: `cmd/` for entry points, `internal/` for private packages

**Do:**
1. Run `go mod init github.com/custom-api-gateway`
2. Create the folder structure (see Project Structure below)
3. Write a `cmd/gateway/main.go` that prints `"API Gateway starting..."` and exits
4. Build and run it: `go run cmd/gateway/main.go`

**Key question to answer:** *Why does Go use `internal/` — what does it enforce?*

---

### Day 2: The `net/http` Package — Your Foundation `[ ]`
**Learn:**
- `http.Handler` interface — just one method: `ServeHTTP(w, r)`
- `http.HandlerFunc` adapter — turns any function into a Handler
- `http.ListenAndServe` — starts a TCP listener and serves HTTP
- Difference between `http.ServeMux` and a custom Handler

**Do:**
1. Replace your main.go: start an HTTP server on port 8080
2. Add a handler at `/` that returns `{"message": "Hello from API Gateway"}`
3. Add a handler at `/ping` that returns `pong`
4. Test with `curl http://localhost:8080/` and `curl http://localhost:8080/ping`

**Key insight:** *Everything in Go's HTTP world is the `http.Handler` interface. The entire gateway is just nested Handlers wrapping Handlers.*

---

### Day 3: Graceful Shutdown with Signals `[ ]`
**Learn:**
- What `os.Signal`, `syscall.SIGINT`, `syscall.SIGTERM` are
- `signal.Notify` — how Go listens for OS signals
- `http.Server.Shutdown(ctx)` — drains active connections before stopping
- `context.WithTimeout` — sets a deadline for the shutdown

**Do:**
1. Replace `http.ListenAndServe` with a custom `http.Server{}`
2. In a goroutine, start the server with `server.ListenAndServe()`
3. In the main goroutine, block on a signal channel
4. On signal: create a 30-second context, call `server.Shutdown(ctx)`
5. Test: start the server, hit it with curl, then Ctrl+C — it should print "shutting down" and exit cleanly

**Key insight:** *Production servers must drain in-flight requests. A hard kill (`os.Exit`) drops active connections.*

---

### Day 4: Goroutines, Channels & Concurrency Model `[ ]`
**Learn:**
- Goroutines are not threads — they're multiplexed onto OS threads by Go's scheduler
- Channels for communication: `ch := make(chan string)`
- `select` statement — non-blocking multi-channel reads
- `sync.Mutex` and `sync.RWMutex` — when channels aren't the right tool
- The rule: *"Don't communicate by sharing memory; share memory by communicating"*

**Do:**
1. Write a small program (outside the gateway) that:
   - Launches 5 goroutines, each "processing" a request (sleep 1-3s random)
   - Uses a channel to collect results
   - Prints results as they arrive
2. Write another program that uses `sync.Mutex` to safely increment a shared counter from 100 goroutines
3. Answer: *When would you use a Mutex vs a Channel?*

**Key insight:** *Every HTTP request the gateway handles runs in its own goroutine. Your rate limiter and circuit breaker will need mutexes for shared state.*

---

### Day 5: `context.Context` — The Request Lifecycle `[ ]`
**Learn:**
- Every `http.Request` carries a `context.Context` via `r.Context()`
- `context.WithTimeout` — auto-cancel after a duration
- `context.WithCancel` — manual cancellation
- `context.WithValue` — attach request-scoped data (request IDs, user info)
- Contexts propagate cancellation down the call tree

**Do:**
1. Write a handler that simulates a slow backend (sleeps 5 seconds)
2. Use `context.WithTimeout` to cancel it after 2 seconds
3. Detect cancellation with `select` on `ctx.Done()` and return `504 Gateway Timeout`
4. Test: hit the endpoint — it should return 504 after 2s, not hang for 5s

**Key insight:** *The gateway must enforce timeouts on upstream calls. If a backend is slow, the gateway uses context cancellation to give up and respond to the client.*

---

### Day 6: HTTP Deep Dive — Headers, Methods, Status Codes `[ ]`
**Learn:**
- HTTP/1.1 request structure: method, URL, headers, body
- Key headers for proxies: `X-Forwarded-For`, `X-Real-IP`, `X-Forwarded-Host`, `X-Forwarded-Proto`
- Status code families: 2xx success, 3xx redirect, 4xx client error, 5xx server error
- `429 Too Many Requests`, `502 Bad Gateway`, `503 Service Unavailable`, `504 Gateway Timeout`

**Do:**
1. Write a handler that logs ALL incoming headers using `r.Header`
2. Add a handler that reads `Content-Type`, parses JSON body, and echoes it back
3. Return different status codes based on the request path:
   - `/ok` → 200, `/created` → 201, `/bad` → 400, `/error` → 500
4. Test each with curl: `curl -v http://localhost:8080/ok`

**Key insight:** *The gateway's job is to inspect, transform, and forward HTTP requests. You must be fluent in HTTP semantics.*

---

### Day 7: Reverse Proxy vs Forward Proxy & `httputil.ReverseProxy` `[ ]`
**Learn:**
- **Forward proxy:** client → proxy → internet (hides the client). Example: VPN, corporate proxy
- **Reverse proxy:** client → proxy → backend servers (hides the servers). Example: Nginx, API gateway
- `httputil.ReverseProxy` — Go's built-in reverse proxy
- The `Director` function — modifies the request before forwarding

**Do:**
1. Start a simple backend on port 8081 (reuse Day 2's hello server)
2. In your gateway, create an `httputil.ReverseProxy` that forwards `/api/` to `localhost:8081`
3. Implement the `Director` function: set `req.URL.Scheme`, `req.URL.Host`, add `X-Forwarded-For`
4. Test: `curl http://localhost:8080/api/` should return the backend's response

**Key insight:** *This is the heart of the gateway. Everything else (auth, rate limiting, load balancing) wraps around this reverse proxy.*

---

### Day 8: Interfaces & the Middleware Pattern `[ ]`
**Learn:**
- Go interfaces are satisfied implicitly (no `implements` keyword)
- `http.Handler` interface: `ServeHTTP(http.ResponseWriter, *http.Request)`
- Middleware pattern: `func(http.Handler) http.Handler` — takes a handler, returns a wrapped handler
- Chaining: `Recovery(Logging(CORS(actualHandler)))` — like Russian nesting dolls

**Do:**
1. Define a type: `type Middleware func(http.Handler) http.Handler`
2. Write a `Chain(handler, ...middlewares)` function that applies them in order
3. Write a simple "timer" middleware that logs how long each request takes
4. Apply it: `Chain(myHandler, timerMiddleware)` and verify the logs

**Key insight:** *The middleware pattern is the single most important design pattern in this project. Auth, logging, rate limiting, circuit breaking — all middlewares.*

**Diagram:**
```
Request → [Recovery] → [RequestID] → [CORS] → [Logging] → [RateLimit] → [Auth] → [Proxy] → Backend
                                                                                          ↓
Response ← [Recovery] ← [RequestID] ← [CORS] ← [Logging] ← [RateLimit] ← [Auth] ← [Proxy] ← Backend
```

---

## WEEK 3-4: Building the Core Gateway

### Day 9: Path-Based Router `[ ]`
**Learn:**
- How routing works: match URL path to a handler
- Prefix matching: `/api/users/123` matches route `/api/users`
- Method filtering: only allow GET, POST for a route

**Do:**
1. Create `internal/router/router.go`
2. Implement a `Router` struct with `AddRoute(path, methods, handler)` and `ServeHTTP`
3. Match by longest prefix (iterate routes, check `strings.HasPrefix`)
4. Return `404 Not Found` for unmatched paths, `405 Method Not Allowed` for wrong methods
5. Write a test in `router_test.go` with at least 3 cases

---

### Day 10: Reverse Proxy Package `[ ]`
**Learn:**
- Separating proxy logic into its own package (`internal/proxy/`)
- Path stripping: gateway has `/api/users`, backend expects `/`
- `strings.TrimPrefix` for path manipulation

**Do:**
1. Create `internal/proxy/proxy.go` — a `ReverseProxy` struct wrapping `httputil.ReverseProxy`
2. Constructor: `New(targetURL, stripPath) (*ReverseProxy, error)`
3. The `Director` function should: set scheme/host, strip prefix, add forwarded headers
4. Implement `ServeHTTP` to satisfy `http.Handler`
5. Write a test using `httptest.NewServer` as a fake backend

---

### Day 11: Wiring Router + Proxy Together `[ ]`
**Do:**
1. In `main.go`, create a router
2. For each configured route, create a `proxy.ReverseProxy` pointing to the backend
3. Register: `router.AddRoute("/api/users", proxy)`
4. Start two example backends (port 8081, 8082) and the gateway (port 8080)
5. Test end-to-end: `curl http://localhost:8080/api/users` → proxied to 8081

---

### Day 12: Logging Middleware `[ ]`
**Learn:**
- `log/slog` — Go's structured logging (added in Go 1.21)
- Why structured logging > `fmt.Println`: searchable, parseable, leveled
- Wrapping `http.ResponseWriter` to capture status code

**Do:**
1. Create `internal/middleware/logging.go`
2. Wrap `ResponseWriter` in a custom struct that captures `statusCode` and `bytesWritten`
3. Log: method, path, status, duration_ms, bytes, remote_addr, request_id
4. Add to the middleware chain and verify the JSON log output

---

### Day 13: Recovery & RequestID Middleware `[ ]`
**Do:**
1. **Recovery middleware** (`recovery.go`): use `defer/recover` to catch panics, log the stack trace with `debug.Stack()`, return 500
2. **RequestID middleware** (`requestid.go`): check for `X-Request-ID` header; if missing, generate one with `crypto/rand`; set it on both request and response
3. Write tests: trigger a panic in a handler, verify recovery returns 500; verify request ID is injected

---

### Day 14: CORS Middleware `[ ]`
**Learn:**
- What CORS is: browsers block cross-origin requests unless the server explicitly allows them
- Preflight requests: browser sends `OPTIONS` before the real request
- Headers: `Access-Control-Allow-Origin`, `Allow-Methods`, `Allow-Headers`, `Max-Age`

**Do:**
1. Create `internal/middleware/cors.go`
2. On every response, set CORS headers
3. For `OPTIONS` requests, return `204 No Content` immediately (don't forward to backend)
4. Make it configurable: `CORSConfig{AllowedOrigins, AllowedMethods, ...}`

---

## WEEK 5-6: Security & Traffic Control

### Day 15: Token Bucket Rate Limiter — Theory `[ ]`
**Learn:**
- **Token bucket algorithm:** a bucket holds N tokens, refills at R tokens/sec, each request costs 1 token
- If bucket is empty → reject with `429 Too Many Requests`
- `burst` = bucket capacity (allows short bursts above the steady rate)
- Why per-client (by IP) instead of global

**Draw/write on paper:**
1. Bucket starts with 10 tokens (burst=10), refills at 5/sec
2. Walk through: 12 requests arrive in 1 second. Which succeed? Which get 429?
3. After 2 seconds of no traffic, how many tokens are available?

---

### Day 16: Token Bucket Rate Limiter — Implementation `[ ]`
**Do:**
1. Create `internal/middleware/ratelimit.go`
2. Struct: `TokenBucket { clients map[string]*bucket, rate float64, burst int }`
3. Each `bucket`: `{ tokens float64, lastCheck time.Time }`
4. `Allow(key)` method: calculate elapsed time, refill tokens, check if ≥1 token available
5. Use `sync.Mutex` to protect the map (multiple goroutines will call this concurrently!)
6. Middleware: extract client IP, call `Allow(ip)`, return 429 with `Retry-After` header if denied
7. Write tests: verify burst allows N requests, then rejects; verify refill after waiting

---

### Day 17: API Key Authentication `[ ]`
**Learn:**
- Simplest auth: client sends `X-API-Key: <key>` header, gateway checks against a whitelist
- Use a `map[string]struct{}` for O(1) lookups (not a slice)

**Do:**
1. Create `internal/middleware/auth.go`
2. `APIKeyAuth(validKeys []string) Middleware`
3. Check `X-API-Key` header → missing = 401, invalid = 403
4. Test with curl: `curl -H "X-API-Key: test-key" http://localhost:8080/api/users`

---

### Day 18: JWT Authentication — Theory `[ ]`
**Learn:**
- JWT structure: `header.payload.signature` (base64url encoded)
- Header: `{"alg": "HS256", "typ": "JWT"}`
- Payload: `{"sub": "user123", "exp": 1700000000, "roles": ["admin"]}`
- Signature: `HMAC-SHA256(base64(header) + "." + base64(payload), secret)`
- Verification: recompute the signature and compare — if it matches, the token is authentic
- **Never trust the payload without verifying the signature first!**

**Do:**
1. On paper, manually construct a JWT:
   - Write the header JSON, base64url-encode it
   - Write the payload JSON, base64url-encode it
   - Compute HMAC-SHA256 of `header.payload` with a secret
2. Answer: *Why can't someone tamper with the payload if they don't know the secret?*

---

### Day 19: JWT Authentication — Implementation `[ ]`
**Do:**
1. In `auth.go`, add `JWTAuth(secret string) Middleware`
2. Parse `Authorization: Bearer <token>` header
3. Split token into 3 parts on `.`
4. Verify signature: `hmac.New(sha256.New, secret)` → compute → compare with `hmac.Equal`
5. Decode payload, check `exp` against `time.Now().Unix()`
6. Set `X-User-ID` and `X-User-Roles` headers for downstream handlers
7. Add `RequireRole(roles...) Middleware` — checks `X-User-Roles` header
8. Write tests with a test JWT (create one in the test using the same HMAC logic)

---

### Day 20: Per-Route Auth Configuration `[ ]`
**Do:**
1. Wire auth middleware per route: some routes need API key, some need JWT, some are public
2. Update the route config to include `auth: { required: true, type: "api_key" }`
3. In `main.go`, when building route handlers, conditionally add auth middleware
4. Test: public route works without auth, protected route returns 401 without auth, 200 with valid auth

---

## WEEK 7-8: Load Balancing & Resilience

### Day 21: Round-Robin Load Balancer `[ ]`
**Learn:**
- Multiple backends for the same route → distribute traffic
- Round-robin: backend 1, 2, 3, 1, 2, 3, ...
- Must skip unhealthy backends

**Do:**
1. Create `internal/loadbalancer/balancer.go`
2. Define `Backend` struct: `URL, Weight, Alive, Handler`
3. Define `Strategy` interface: `Next(backends []*Backend) *Backend`
4. Implement `RoundRobin` strategy with an atomic/mutex counter
5. Write tests: 3 backends, verify requests cycle through them

---

### Day 22: Weighted Round-Robin `[ ]`
**Learn:**
- Some backends are more powerful → give them more traffic
- Weights: backend A (weight 3), backend B (weight 1) → A gets 75% of traffic
- Classic algorithm: track `currentWeight`, decrement by GCD each cycle

**Do:**
1. Create `internal/loadbalancer/weighted.go`
2. Implement the weighted round-robin selection algorithm
3. Test: backends with weights [3, 1] — over 100 requests, A should get ~75

---

### Day 23: Least-Connections Balancer `[ ]`
**Learn:**
- Track active connections per backend
- Pick the backend with the fewest active connections
- Use `sync/atomic` for lock-free counter updates

**Do:**
1. Create `internal/loadbalancer/leastconn.go`
2. Use `atomic.LoadInt64` / `atomic.AddInt64` for the connection counter
3. `Next()`: iterate backends, find alive one with min `ActiveConns`
4. Caller must call `Done(backend)` when the request completes to decrement

---

### Day 24: Circuit Breaker — Theory `[ ]`
**Learn:**
- Problem: backend is down, but gateway keeps sending requests → wastes resources, slow responses
- Solution: circuit breaker — stop sending requests to a failing backend
- Three states:
  - **Closed** (normal): requests flow through. Track failures.
  - **Open** (tripped): reject immediately with 503. Wait for recovery timeout.
  - **Half-Open** (testing): allow a few requests through. If they succeed → Closed. If they fail → Open.

**Draw on paper:**
```
     success          failure >= threshold
  ┌──────────┐      ┌──────────────────────┐
  │          │      │                      │
  ▼          │      ▼                      │
CLOSED ──────┘    OPEN ──── timeout ──── HALF-OPEN
  ▲                                        │
  │              success >= threshold       │
  └────────────────────────────────────────┘
```
Walk through a scenario: 5 failures → open → 30s wait → half-open → 3 successes → closed

---

### Day 25: Circuit Breaker — Implementation `[ ]`
**Do:**
1. Create `internal/middleware/circuitbreaker.go`
2. Struct: `CircuitBreaker { state, failures, successes, failureThreshold, recoveryTimeout, ... }`
3. Methods: `State()`, `RecordSuccess()`, `RecordFailure()`
4. State transitions as described in Day 24
5. Middleware: check state before forwarding; after response, record success/failure based on status code (5xx = failure)
6. Write tests: simulate failures to trip the breaker, verify it rejects, verify recovery

---

### Day 26: Health Checks `[ ]`
**Learn:**
- **Active health check:** gateway periodically pings each backend's `/health` endpoint
- **Passive health check:** gateway marks backend unhealthy based on response errors (circuit breaker does this)
- Active checks run in background goroutines using `time.Ticker`

**Do:**
1. Create `internal/health/checker.go`
2. For each backend, start a goroutine that pings `backend.URL + "/health"` every N seconds
3. If response is 200 → `backend.SetAlive(true)`, else → `backend.SetAlive(false)`
4. Use `context.WithCancel` to stop checkers on shutdown
5. Add `/health` endpoint to the example services

---

### Day 27: Admin Endpoints — /health & /metrics `[ ]`
**Do:**
1. Create `internal/admin/admin.go`
2. Start a separate HTTP server on port 9090 (admin port)
3. `/health` — returns JSON: gateway status, list of backends with alive/dead status
4. `/metrics` — returns JSON: uptime, goroutine count, heap memory (`runtime.ReadMemStats`)
5. Test: start gateway, check `curl http://localhost:9090/health`

---

## WEEK 9-10: Configuration, Testing & Polish

### Day 28: YAML Configuration `[ ]`
**Learn:**
- `gopkg.in/yaml.v3` — the standard YAML library for Go
- Struct tags: `yaml:"field_name"` control mapping
- `time.Duration` can be parsed from strings like `"15s"`, `"1m"`

**Do:**
1. Create `internal/config/config.go`
2. Define config structs matching `configs/gateway.yaml` (server, routes, auth, etc.)
3. `Load(path) (*Config, error)` — read file, unmarshal YAML, apply defaults
4. Add env var overrides: `GATEWAY_PORT`, `GATEWAY_JWT_SECRET`, etc.
5. Write tests: load a test YAML, verify parsed values

---

### Day 29: Wire Everything in main.go `[ ]`
**Do:**
1. Load config from `configs/gateway.yaml`
2. For each route in config: create proxy → wrap with circuit breaker → create backend → load balancer → add auth/rate-limit middleware → register on router
3. Apply global middleware: Recovery → RequestID → CORS → Logging
4. Start health checker, admin server, and gateway server
5. Test the complete flow end-to-end

---

### Day 30-31: Unit Tests — Router, Proxy, Middleware `[ ]`
**Learn:**
- `httptest.NewServer` — spins up a real HTTP server for tests
- `httptest.NewRecorder` — captures the response without a real server
- Table-driven tests: `tests := []struct{ name, input, expected }{...}`

**Do:**
1. `router/router_test.go` — test path matching, method filtering, 404, 405
2. `proxy/proxy_test.go` — test forwarding, path stripping, header injection
3. `middleware/*_test.go` — test each middleware in isolation
4. Run: `go test ./... -cover` — target ≥ 80% coverage

---

### Day 32-33: Unit Tests — Load Balancer, Circuit Breaker, Config `[ ]`
**Do:**
1. `loadbalancer/*_test.go` — round-robin cycles, weighted distribution, least-conn selection, skip dead backends
2. `middleware/circuitbreaker_test.go` — state transitions: closed→open→half-open→closed
3. `middleware/ratelimit_test.go` — burst allows N, then rejects, refill works
4. `middleware/auth_test.go` — valid/invalid API keys, valid/expired/tampered JWTs
5. `config/config_test.go` — YAML parsing, env overrides, defaults
6. Run: `go test ./... -cover -v`

---

### Day 34: Benchmark Tests `[ ]`
**Learn:**
- `func BenchmarkXxx(b *testing.B)` — Go's built-in benchmarking
- `b.N` — the framework decides how many iterations to run
- `b.ReportAllocs()` — shows memory allocations per operation

**Do:**
1. Benchmark the rate limiter's `Allow()` method
2. Benchmark the router's `ServeHTTP` with 10 routes
3. Benchmark round-robin `Next()` with 5 backends
4. Run: `go test ./... -bench=. -benchmem`
5. Analyze: are there unnecessary allocations? Can you reduce them?

---

### Day 35: Example Services & End-to-End `[ ]`
**Do:**
1. Finalize `examples/users-service/main.go` — returns mock user data, has `/health`
2. Finalize `examples/orders-service/main.go` — returns mock order data, has `/health`
3. Start all three processes: users (8081), orders (8082), gateway (8080)
4. Test the full flow:
   - `curl http://localhost:8080/api/users` (with API key)
   - `curl http://localhost:8080/api/orders` (no auth needed)
   - `curl http://localhost:9090/health` (admin)
   - Rapid requests to test rate limiting
   - Kill a backend to test circuit breaker + health checks

---

### Day 36: README & Documentation `[ ]`
**Do:**
1. Write `README.md`: what this is, architecture diagram (ASCII), how to run, config reference, API reference
2. Write `CHANGELOG.md`
3. Review all code for clarity, remove any dead code
4. Celebrate — you built an API gateway from scratch! 🎉

---

## Project Structure

```
custom-api-gateway/
├── cmd/
│   └── gateway/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # YAML + env config loader
│   ├── server/
│   │   └── server.go            # HTTP server with graceful shutdown
│   ├── proxy/
│   │   └── proxy.go             # Reverse proxy logic
│   ├── router/
│   │   └── router.go            # Path-based route matching
│   ├── middleware/
│   │   ├── chain.go             # Middleware chaining
│   │   ├── logging.go           # Request/response logger
│   │   ├── recovery.go          # Panic recovery
│   │   ├── cors.go              # CORS headers
│   │   ├── requestid.go         # X-Request-ID injection
│   │   ├── ratelimit.go         # Token-bucket rate limiter
│   │   ├── auth.go              # API-key + JWT auth
│   │   └── circuitbreaker.go    # Circuit breaker
│   ├── loadbalancer/
│   │   ├── balancer.go          # Interface + round-robin
│   │   ├── weighted.go          # Weighted round-robin
│   │   └── leastconn.go         # Least-connections
│   ├── health/
│   │   └── checker.go           # Active health checks
│   └── admin/
│       └── admin.go             # /health, /metrics endpoints
├── configs/
│   └── gateway.yaml             # Default config
├── examples/
│   ├── users-service/
│   │   └── main.go
│   └── orders-service/
│       └── main.go
├── go.mod
├── go.sum
├── README.md
└── CHANGELOG.md
```

---

## Session Rules for Copilot Agent

1. **On each session start:** Read this file. Resume from the first unchecked `[ ]` day.
2. **If user says "I'm on Day X":** Jump to that day, explain the concepts, guide the implementation.
3. **Don't write full code unprompted:** Explain the concept, give the function signature/pseudocode, let the user write it. Provide the full solution only if they ask or are stuck.
4. **After completing a day:** Remind the user to mark `[x]` and commit their work.
5. **Testing:** Remind the user to write tests for every implementation day.
6. **Reference code:** The files in `internal/` are reference implementations — point the user to compare after they attempt it themselves.
7. **Allowed dependencies:** stdlib only + `gopkg.in/yaml.v3`. No third-party routers/frameworks.

---

## Quick Reference: Key Go Packages Used

| Package | Purpose |
|---|---|
| `net/http` | HTTP server, client, Handler interface |
| `net/http/httputil` | `ReverseProxy` |
| `net/http/httptest` | Test servers and response recorders |
| `context` | Timeouts, cancellation, request-scoped values |
| `log/slog` | Structured logging |
| `sync` | Mutex, RWMutex for thread safety |
| `sync/atomic` | Lock-free counters |
| `crypto/hmac`, `crypto/sha256` | JWT signature verification |
| `crypto/rand` | Secure random bytes (request IDs) |
| `encoding/json` | JSON encode/decode |
| `encoding/base64` | JWT base64url encoding |
| `os/signal`, `syscall` | Graceful shutdown |
| `time` | Timeouts, tickers, durations |
| `runtime` | Memory stats, goroutine count |
| `gopkg.in/yaml.v3` | YAML config parsing |

---