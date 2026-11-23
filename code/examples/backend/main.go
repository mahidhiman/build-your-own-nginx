package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:9001", "listen address")
	name := flag.String("name", "backend-a", "backend name")
	delay := flag.Duration("delay", 0, "artificial response delay")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *delay > 0 {
			time.Sleep(*delay)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "served by %s\npath=%s\n", *name, r.URL.Path)
	})
	log.Printf("backend %s listening on %s", *name, *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
