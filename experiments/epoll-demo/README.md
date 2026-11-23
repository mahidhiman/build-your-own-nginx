# Linux-only epoll lab

Run `go run .` on Linux and connect using `nc 127.0.0.1 9099`. Type text and observe the echo. Use `Ctrl+C` to stop it.

This lab demonstrates a tiny readiness loop using `syscall` and a newline-independent byte-stream echo. It is intentionally incomplete: it has no HTTP parser, no timer management, and simplified output buffering. Do not expose it outside loopback or use it as a server template. The code makes partial writes and `EAGAIN` visible so you can study state transitions; it is not a hardened network service.
