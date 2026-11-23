# TinyProxy - companion reference implementation

This is an educational Go reverse proxy paired with **Build Your Own NGINX**.
It is not production-ready and is not a drop-in replacement for NGINX.

## Requirements

- Go 1.23 or newer (the `go` directive in `go.mod` is 1.23)
- Linux, macOS, Windows, or WSL for core labs
- Optional: `curl`, `hey` or `ab`, `ss`, `lsof`, `openssl` for labs

## Run two backends and proxy

Terminal 1:

```sh
go run ./examples/backend -listen 127.0.0.1:9001 -name backend-a
```

Terminal 2:

```sh
go run ./examples/backend -listen 127.0.0.1:9002 -name backend-b -delay 50ms
```

Terminal 3:

```sh
cp config.example.json config.json
go run ./cmd/tinyproxy -config config.json
```

Terminal 4:

```sh
curl -i http://127.0.0.1:8080/hello
curl -i http://127.0.0.1:8080/healthz
```

The `/healthz` route is handled locally; other paths go through the proxy.
Repeated requests should rotate between configured upstreams.

## Verify

```sh
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

## Benchmark safely

Start with low concurrency against loopback only:

```sh
hey -z 20s -c 20 http://127.0.0.1:8080/hello
```

Install benchmark tools from their trusted distribution channels. Record exact commands, OS/kernel, CPU/RAM, Go version, commit, configuration, duration, concurrency, request/error counts, p50/p95/p99 latency, CPU and memory. Avoid exposing the test server beyond localhost.

## Intentional limitations

- No NGINX-compatible configuration parser, modules, process manager, disk cache, or advanced health checking.
- Load balancing is round-robin only; no active health checks or retry policy.
- Response errors are intentionally simple; status handling should be expanded and tested in the course.
- Forwarded-header trust and proxy-chain policy are not production complete.
- Request-body limits, metrics endpoint, TLS listener, security audit, fuzzing coverage, and complete access-log semantics remain course milestones.
- The in-flight limiter bounds handled proxy requests but should not be mistaken for complete connection admission control.

Use this repo to learn and to experiment on localhost. Use maintained software for production ingress.
