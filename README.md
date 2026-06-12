# proxy-go

A lightweight, production-minded reverse proxy and API gateway written in Go.

Single static binary, configured with one YAML file. Built to sit in front of
upstream HTTP services and provide the resilience and observability layer they
shouldn't have to implement themselves.

## Features

- **Routing** — host- and path-based routing to multiple upstreams
- **Load balancing** — round-robin across upstream replicas, with passive health checks
- **Rate limiting** — token bucket per client IP, configurable per route
- **Circuit breaker** — stop hammering an unhealthy upstream, recover automatically (half-open probing)
- **Retries** — configurable retry policy for idempotent requests, with exponential backoff
- **Timeouts** — sane defaults everywhere; every request carries a deadline via `context.Context`
- **Observability** — Prometheus metrics endpoint (`/metrics`), structured logs (JSON), request IDs
- **Graceful shutdown** — drains in-flight requests on `SIGTERM`, no dropped connections on deploy
- **Hot config reload** — `SIGHUP` re-reads the config without dropping traffic _(planned)_
- **TLS termination** _(planned)_

## Why

Built as a learning project to write idiomatic, concurrent Go — and as a real
tool: it runs in front of a production e-commerce backend. The goal is a
codebase small enough to read in an afternoon, but correct enough to trust
with real traffic.

Not a replacement for nginx, Envoy or Traefik. If you need those, use those.

## Quick start

```bash
# build
go build -o proxy-go ./cmd/proxy

# run with a config file
./proxy-go -config config.yaml
```

Minimal `config.yaml`:

```yaml
server:
  listen: ":8080"
  read_timeout: 10s
  write_timeout: 30s

metrics:
  listen: ":9090" # /metrics for Prometheus

routes:
  - match:
      host: "api.example.com"
      path_prefix: "/"
    upstreams:
      - "http://10.0.0.10:3000"
      - "http://10.0.0.11:3000"
    rate_limit:
      requests_per_second: 50
      burst: 100
    retry:
      attempts: 2
      methods: [GET, HEAD]
    circuit_breaker:
      failure_threshold: 5
      open_duration: 30s
```

## Architecture

```
                        ┌──────────────────────────────────────────┐
                        │                proxy-go                  │
                        │                                          │
 client ──── HTTP ────► │  router ─► middleware chain ─► forwarder │ ────► upstream A
                        │             │                            │ ────► upstream B
                        │             ├─ request ID                │
                        │             ├─ logging                   │
                        │             ├─ rate limiter              │
                        │             ├─ circuit breaker           │
                        │             └─ retry                     │
                        │                                          │
                        │  :9090 /metrics (Prometheus)             │
                        └──────────────────────────────────────────┘
```

Each request flows through a middleware chain built with standard
`func(http.Handler) http.Handler` composition — no framework, just the
standard library plus a Prometheus client.

## Project layout

```
cmd/proxy/            entry point (flag parsing, wiring, signal handling)
internal/config/      YAML config loading + validation
internal/router/      route matching (host / path prefix)
internal/proxy/       the forwarder (httputil.ReverseProxy wrapper, load balancing)
internal/middleware/  rate limit, circuit breaker, retry, logging, request ID
internal/metrics/     Prometheus collectors
internal/health/      passive upstream health tracking
```

## Development

```bash
go test ./...          # unit tests
go vet ./...           # static analysis
golangci-lint run      # lint (https://golangci-lint.run)
```

Load testing a local setup:

```bash
# terminal 1: a dummy upstream
go run ./testdata/upstream

# terminal 2: the proxy
go run ./cmd/proxy -config testdata/config.yaml

# terminal 3: traffic
hey -z 30s -c 50 http://localhost:8080/
```

## Roadmap

- [ ] v0.1 — routing + forwarding + graceful shutdown
- [ ] v0.2 — rate limiting + structured logs + request IDs
- [ ] v0.3 — circuit breaker + retries + passive health checks
- [ ] v0.4 — Prometheus metrics + Grafana dashboard example
- [ ] v1.0 — running in front of real production traffic
- [ ] later — hot reload (SIGHUP), TLS termination, active health checks

## License

MIT

proxy-go/
├── go.mod
├── README.md
├── config.yaml # config d'exemple
├── cmd/
│ └── proxy/
│ └── main.go # parsing flags, wiring, signaux
(SIGTERM)
├── internal/
│ ├── config/
│ │ ├── config.go # structs + chargement YAML + validation
│ │ └── config_test.go
│ ├── router/
│ │ ├── router.go # matching host / path prefix → route
│ │ └── router_test.go
│ ├── proxy/
│ │ ├── forwarder.go # wrapper httputil.ReverseProxy
│ │ └── balancer.go # round-robin entre upstreams
│ ├── middleware/
│ │ ├── chain.go # composition func(http.Handler)
http.Handler
│ │ ├── requestid.go
│ │ ├── logging.go # logs JSON via log/slog (stdlib)
│ │ ├── ratelimit.go # token bucket par IP
│ │ ├── circuitbreaker.go
│ │ └── retry.go
│ ├── metrics/
│ │ └── metrics.go # collectors Prometheus
│ └── health/
│ └── tracker.go # santé passive des upstreams
└── testdata/
├── config.yaml # config pour les tests/dev local
└── upstream/
└── main.go # faux upstream pour tester en local
