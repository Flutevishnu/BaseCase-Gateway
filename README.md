# BaseCase Gateway

A high-performance API Gateway built in Go (Golang). This gateway serves as the single public entry point for the microservices architecture, handling routing, authentication, and rate limiting.

## Features

- **Reverse Proxy:** Built on Go's native `httputil.ReverseProxy` to route traffic to internal downstream services (e.g., Express Backend, Go Orchestrator).
- **Authentication:** Middleware that intercepts requests, dynamically fetches Clerk JWKS, and securely verifies JWTs before traffic reaches downstream services.
- **Rate Limiting:** Global Redis-based rate limiter to protect the ecosystem from abuse.
- **Graceful Shutdown:** Safely drains active HTTP connections upon receiving termination signals.
- **Structured Logging:** Uses Go's native `log/slog` for clean, standard JSON logs.
- **Panic Recovery:** Global middleware ensures the gateway never crashes unexpectedly from bad requests.

## Prerequisites

- [Go 1.24+](https://golang.org/dl/)
- [Redis](https://redis.io/download) (or Docker to run Redis locally)
- A Clerk Account (for authentication)

## Getting Started

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Setup your environment variables:**
   Copy the example environment file and fill in your Clerk domain and service URLs.
   ```bash
   cp .env.example .env
   ```

3. **Start a local Redis instance:**
   If you have Docker installed, the easiest way to get Redis running is:
   ```bash
   docker run -d -p 6379:6379 redis:latest
   ```

4. **Run the Gateway:**
   ```bash
   go run cmd/gateway/main.go
   ```
   The gateway will start on `http://localhost:8080`.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PORT` | The port the gateway listens on (default: 8080) |
| `REDIS_ADDR` | Connection string for Redis (default: localhost:6379) |
| `CLERK_ISSUER` | Your Clerk application domain (e.g. `https://your-domain.clerk.accounts.dev`) |
| `CLERK_JWKS_URL` | The JWKS URL from Clerk (e.g. `.../.well-known/jwks.json`) |
| `EXPRESS_BACKEND_URL` | Target URL for the Express backend |
| `ORCHESTRATOR_URL` | Target URL for the Go orchestrator |
| `AGENT_API_URL` | Target URL for the Agent microservice |

## Project Structure

```text
BaseCase-Gateway/
├── cmd/
│   └── gateway/
│       └── main.go               # Entry point (Server, routes, & shutdown logic)
├── internal/
│   ├── config/
│   │   └── config.go             # Environment parsing & routing targets
│   ├── middleware/
│   │   ├── auth.go               # Clerk JWT validation
│   │   ├── chain.go              # Middleware wrapping logic
│   │   ├── logger.go             # slog JSON HTTP request logging
│   │   ├── ratelimit.go          # Redis rate limiting implementation
│   │   └── recovery.go           # Panic recovery wrapper
│   └── proxy/
│       └── proxy.go              # httputil.ReverseProxy implementation
└── .env.example                  # Environment variables template
```
