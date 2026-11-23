# Build Your Own NGINX — companion code

Runnable reference code for the book **Build Your Own NGINX: A Systems Programming Textbook in Go** by Manish Dhiman.

```sh
git clone https://github.com/mahidhiman/build-your-own-nginx
cd build-your-own-nginx/code
go test ./...
```

## Repository map

```
code/                    # TinyProxy: the complete educational proxy and its tests
stages/                  # incremental runnable checkpoints, one per build stage
experiments/epoll-demo/  # Linux-only readiness (epoll) demonstration
labs/                    # benchmark harness template and lab notes
scripts/test_all.sh      # gofmt, go vet, go test -race and go build for everything
```

This is an educational project, not a production reverse proxy or NGINX replacement.

## Requirements

- Go 1.23 or newer
- Optional lab tools: `curl`, `hey` or `ab`, `ss`, `lsof`, and `openssl`
- The `epoll` demo is Linux-only

## Run locally

Open three terminals from `code/`:

```sh
go run ./examples/backend -listen 127.0.0.1:9001 -name backend-a
```

```sh
go run ./examples/backend -listen 127.0.0.1:9002 -name backend-b -delay 50ms
```

```sh
cp config.example.json config.json
go run ./cmd/tinyproxy -config config.json
```

Test in another terminal:

```sh
curl -i http://127.0.0.1:8080/hello
curl -i http://127.0.0.1:8080/healthz
```

Keep listeners on `127.0.0.1` for initial experiments. Do not expose the educational proxy or raw `epoll` demo to the public Internet.

## Validate the project

From the repository root:

```sh
./scripts/test_all.sh
```

The script runs `gofmt`, `go test -race`, `go vet`, and `go build` for the main proxy plus each standalone stage and the Linux experiment.

## Intentional limitations

The project does not implement an NGINX-compatible configuration language, master/worker supervisor, full HTTP protocol stack, HTTP/2 or HTTP/3 server, standards-complete cache, active health checks, retry policy, complete forwarded-header trust policy, production-grade admission controls, or a complete security/operations program. Consult the book and official docs for the boundary between the learning implementation and production behavior.

## License

MIT. See LICENSE.
