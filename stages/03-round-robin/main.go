package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

type selector struct { targets []*url.URL; next atomic.Uint64 }
func newSelector(raw []string) (*selector, error) {
	if len(raw) == 0 { return nil, errors.New("at least one upstream is required") }
	s := &selector{}
	for _, item := range raw {
		u, err := url.Parse(item)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil { return nil, errors.New("upstream must be a valid http(s) URL without userinfo") }
		s.targets = append(s.targets, u)
	}
	return s, nil
}
func (s *selector) pick() *url.URL { i := s.next.Add(1)-1; cp := *s.targets[i%uint64(len(s.targets))]; return &cp }
func main() {
	raw := os.Getenv("UPSTREAMS")
	if raw == "" { raw = "http://127.0.0.1:9001,http://127.0.0.1:9002" }
	s, err := newSelector(strings.Split(raw, ",")); if err != nil { log.Fatal(err) }
	p := &httputil.ReverseProxy{Rewrite: func(pr *httputil.ProxyRequest) { pr.SetURL(s.pick()); pr.Out.Host = pr.In.Host }}
	srv := &http.Server{Addr: "127.0.0.1:8080", Handler: p, ReadHeaderTimeout: 5_000_000_000, IdleTimeout: 60_000_000_000, MaxHeaderBytes: 1 << 20}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { log.Fatal(err) }
}
