//go:build linux

// Command epoll-demo is a deliberately small Linux-only line-protocol echo
// server for learning readiness notification. It is not an HTTP server.
// Bind is fixed to loopback; do not remove that restriction and expose it.
package main

import (
	"fmt"
	"log"
	"syscall"
)

type client struct {
	fd  int
	out []byte
}

func main() {
	listener, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, 0)
	must(err)
	defer syscall.Close(listener)
	must(syscall.SetsockoptInt(listener, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1))
	addr := &syscall.SockaddrInet4{Port: 9099, Addr: [4]byte{127, 0, 0, 1}}
	must(syscall.Bind(listener, addr))
	must(syscall.Listen(listener, 128))

	epfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	must(err)
	defer syscall.Close(epfd)
	must(syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, listener, &syscall.EpollEvent{Events: syscall.EPOLLIN, Fd: int32(listener)}))

	log.Println("loopback echo lab listening on 127.0.0.1:9099; send lines, Ctrl+C to stop")
	clients := map[int]*client{}
	events := make([]syscall.EpollEvent, 128)
	buf := make([]byte, 4096)

	for {
		n, err := syscall.EpollWait(epfd, events, -1)
		if err == syscall.EINTR {
			continue
		}
		must(err)
		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)
			if fd == listener {
				for {
					cfd, _, acceptErr := syscall.Accept4(listener, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC)
					if acceptErr == syscall.EAGAIN || acceptErr == syscall.EWOULDBLOCK {
						break
					}
					if acceptErr == syscall.EINTR {
						continue
					}
					if acceptErr != nil {
						log.Printf("accept: %v", acceptErr)
						break
					}
					c := &client{fd: cfd}
					clients[cfd] = c
					if ctlErr := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, cfd, &syscall.EpollEvent{Events: syscall.EPOLLIN | syscall.EPOLLRDHUP | syscall.EPOLLERR | syscall.EPOLLHUP, Fd: int32(cfd)}); ctlErr != nil {
						log.Printf("register client: %v", ctlErr)
						closeClient(epfd, clients, cfd)
					}
				}
				continue
			}

			c, ok := clients[fd]
			if !ok {
				continue
			}
			if events[i].Events&(syscall.EPOLLERR|syscall.EPOLLHUP|syscall.EPOLLRDHUP) != 0 {
				closeClient(epfd, clients, fd)
				continue
			}

			if events[i].Events&syscall.EPOLLIN != 0 {
				closed := false
				for {
					readN, readErr := syscall.Read(fd, buf)
					if readN > 0 {
						// Simple echo, bounded per event turn; a production event loop
						// would need a bounded output queue and partial-write state.
						c.out = append(c.out, buf[:readN]...)
					}
					if readErr == syscall.EINTR {
						continue
					}
					if readErr == syscall.EAGAIN || readErr == syscall.EWOULDBLOCK {
						break
					}
					if readErr != nil {
						log.Printf("read fd=%d: %v", fd, readErr)
						closed = true
						break
					}
					if readN == 0 {
						closed = true
						break
					}
				}
				if closed {
					closeClient(epfd, clients, fd)
					continue
				}
			}

			if len(c.out) > 0 {
				for len(c.out) > 0 {
					written, writeErr := syscall.Write(fd, c.out)
					if written > 0 {
						c.out = c.out[written:]
					}
					if writeErr == syscall.EINTR {
						continue
					}
					if writeErr == syscall.EAGAIN || writeErr == syscall.EWOULDBLOCK {
						break
					}
					if writeErr != nil {
						log.Printf("write fd=%d: %v", fd, writeErr)
						closeClient(epfd, clients, fd)
						c = nil
						break
					}
					if written == 0 {
						break
					}
				}
				if c == nil {
					continue
				}
			}

			mask := uint32(syscall.EPOLLIN | syscall.EPOLLRDHUP | syscall.EPOLLERR | syscall.EPOLLHUP)
			if len(c.out) > 0 {
				mask |= syscall.EPOLLOUT
			}
			if modErr := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Events: mask, Fd: int32(fd)}); modErr != nil {
				log.Printf("modify fd=%d: %v", fd, modErr)
				closeClient(epfd, clients, fd)
			}
		}
	}
}

func closeClient(epfd int, clients map[int]*client, fd int) {
	_ = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_DEL, fd, nil)
	_ = syscall.Close(fd)
	delete(clients, fd)
}

func must(err error) {
	if err != nil {
		log.Fatal(fmt.Errorf("epoll lab: %w", err))
	}
}
