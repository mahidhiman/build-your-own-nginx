# Runnable checkpoints

Each directory is a separate, intentionally small Go module. Run its commands from within that directory. Later chapters point from these checkpoints to the full implementation at `../code`.

| Stage | Demonstrates | Run |
|---|---|---|
| `00-hello` | First handler and listener | `go run .` |
| `01-http-server` | Explicit server timeouts and signal-aware shutdown | `go run .` |
| `02-single-upstream` | One configured reverse-proxy target | `UPSTREAM_URL=http://127.0.0.1:9001 go run .` |
| `03-round-robin` | Immutable URL list and atomic rotation | `UPSTREAMS=http://127.0.0.1:9001,http://127.0.0.1:9002 go run .` |

Stages intentionally leave gaps so that the learner can add tests and failure handling. They are learning snapshots, not production programs. The complete reference code is in `code/` and has the broader tests described in the book.
