package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	raw := os.Getenv("UPSTREAM_URL")
	if raw == "" { raw = "http://127.0.0.1:9001" }
	target, err := url.Parse(raw)
	if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") {
		log.Fatal("UPSTREAM_URL must be an absolute http(s) URL")
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	srv := &http.Server{Addr: "127.0.0.1:8080", Handler: proxy, ReadHeaderTimeout: 5_000_000_000, IdleTimeout: 60_000_000_000, MaxHeaderBytes: 1 << 20}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { log.Fatal(err) }
}
