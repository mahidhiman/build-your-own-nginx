package proxy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

var ErrNoUpstreams = errors.New("no upstreams configured")

type Selector struct {
	urls []*url.URL
	next atomic.Uint64
}

func NewSelector(raw []string) (*Selector, error) {
	if len(raw) == 0 {
		return nil, ErrNoUpstreams
	}
	s := &Selector{}
	for _, item := range raw {
		u, err := url.Parse(item)
		if err != nil {
			return nil, err
		}
		if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, errors.New("upstream must be an absolute http(s) URL")
		}
		s.urls = append(s.urls, u)
	}
	return s, nil
}

func (s *Selector) Next() *url.URL {
	i := s.next.Add(1) - 1
	u := *s.urls[i%uint64(len(s.urls))]
	return &u
}

func NewHandler(rawUpstreams []string, timeout time.Duration) (http.Handler, error) {
	selector, err := NewSelector(rawUpstreams)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 256
	transport.MaxIdleConnsPerHost = 64
	transport.MaxConnsPerHost = 128
	transport.IdleConnTimeout = 60 * time.Second
	transport.ResponseHeaderTimeout = timeout
	transport.TLSHandshakeTimeout = 5 * time.Second
	transport.ExpectContinueTimeout = 1 * time.Second

	proxy := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(pr *httputil.ProxyRequest) {
			target := selector.Next()
			pr.SetURL(target)
			pr.Out.Host = pr.In.Host
			// Trust the direct socket peer only. This tiny lab intentionally
			// does not implement multi-proxy trusted-forwarding configuration.
			pr.Out.Header.Set("X-Forwarded-Proto", schemeOf(pr.In))
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if errors.Is(r.Context().Err(), context.Canceled) {
				return // the client went away; there may be no response to send
			}
			var netErr net.Error
			if errors.Is(r.Context().Err(), context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
				http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
				return
			}
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
		},
	}
	return proxy, nil
}

func schemeOf(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
