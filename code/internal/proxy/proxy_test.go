package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRoundRobinSelector(t *testing.T) {
	s, err := NewSelector([]string{"http://127.0.0.1:9001", "http://127.0.0.1:9002"})
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Next().Host; got != "127.0.0.1:9001" {
		t.Fatalf("first host = %q", got)
	}
	if got := s.Next().Host; got != "127.0.0.1:9002" {
		t.Fatalf("second host = %q", got)
	}
	if got := s.Next().Host; got != "127.0.0.1:9001" {
		t.Fatalf("third host = %q", got)
	}
}

func TestProxyForwardsPathAndBalances(t *testing.T) {
	backendA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "A:"+r.URL.Path) }))
	defer backendA.Close()
	backendB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "B:"+r.URL.Path) }))
	defer backendB.Close()

	h, err := NewHandler([]string{backendA.URL, backendB.URL}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	proxyServer := httptest.NewServer(h)
	defer proxyServer.Close()

	want := []string{"A:/hello", "B:/hello", "A:/hello", "B:/hello"}
	for i, expected := range want {
		resp, err := http.Get(proxyServer.URL + "/hello")
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d status=%d body=%s", i, resp.StatusCode, body)
		}
		if got := strings.TrimSpace(string(body)); got != expected {
			t.Fatalf("request %d body=%q want=%q", i, got, expected)
		}
	}
}

func TestProxyTimeoutReturnsGatewayTimeout(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = io.WriteString(w, "late")
	}))
	defer backend.Close()
	h, err := NewHandler([]string{backend.URL}, 25*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	proxyServer := httptest.NewServer(h)
	defer proxyServer.Close()
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(proxyServer.URL + "/slow")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Fatalf("status=%d want=%d", resp.StatusCode, http.StatusGatewayTimeout)
	}
}
