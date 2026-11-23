package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Listen              string   `json:"listen"`
	Upstreams           []string `json:"upstreams"`
	ReadHeaderTimeoutMS int      `json:"read_header_timeout_ms"`
	IdleTimeoutMS       int      `json:"idle_timeout_ms"`
	UpstreamTimeoutMS   int      `json:"upstream_timeout_ms"`
	MaxInflight         int      `json:"max_inflight"`
}

func Defaults() Config {
	return Config{
		Listen:              "127.0.0.1:8080",
		Upstreams:           []string{"http://127.0.0.1:9001"},
		ReadHeaderTimeoutMS: 5000,
		IdleTimeoutMS:       60000,
		UpstreamTimeoutMS:   3000,
		MaxInflight:         256,
	}
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()
	cfg := Defaults()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, errors.New("config must contain exactly one JSON value")
		}
		return Config{}, fmt.Errorf("trailing config data: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if _, _, err := net.SplitHostPort(c.Listen); err != nil {
		return fmt.Errorf("listen must be a host:port address: %w", err)
	}
	if len(c.Upstreams) == 0 {
		return errors.New("at least one upstream is required")
	}
	for i, raw := range c.Upstreams {
		u, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("upstream %d: %w", i, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("upstream %d: scheme must be http or https", i)
		}
		if u.Host == "" {
			return fmt.Errorf("upstream %d: host is required", i)
		}
		if u.User != nil {
			return fmt.Errorf("upstream %d: userinfo in URL is not allowed", i)
		}
	}
	if c.ReadHeaderTimeoutMS < 1 || c.ReadHeaderTimeoutMS > int((2*time.Minute)/time.Millisecond) {
		return errors.New("read_header_timeout_ms must be between 1 and 120000")
	}
	if c.IdleTimeoutMS < 1 || c.IdleTimeoutMS > int((10*time.Minute)/time.Millisecond) {
		return errors.New("idle_timeout_ms must be between 1 and 600000")
	}
	if c.UpstreamTimeoutMS < 1 || c.UpstreamTimeoutMS > int((5*time.Minute)/time.Millisecond) {
		return errors.New("upstream_timeout_ms must be between 1 and 300000")
	}
	if c.MaxInflight < 1 || c.MaxInflight > 100000 {
		return errors.New("max_inflight must be between 1 and 100000")
	}
	return nil
}

func (c Config) ReadHeaderTimeout() time.Duration {
	return time.Duration(c.ReadHeaderTimeoutMS) * time.Millisecond
}
func (c Config) IdleTimeout() time.Duration { return time.Duration(c.IdleTimeoutMS) * time.Millisecond }
func (c Config) UpstreamTimeout() time.Duration {
	return time.Duration(c.UpstreamTimeoutMS) * time.Millisecond
}
