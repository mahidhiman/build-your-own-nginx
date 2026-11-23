package config

import "testing"

func TestValidate(t *testing.T) {
	good := Defaults()
	if err := good.Validate(); err != nil {
		t.Fatalf("defaults invalid: %v", err)
	}
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"empty upstream list", func(c *Config) { c.Upstreams = nil }, true},
		{"unsupported scheme", func(c *Config) { c.Upstreams = []string{"file:///etc/passwd"} }, true},
		{"userinfo forbidden", func(c *Config) { c.Upstreams = []string{"http://user:pass@example.com"} }, true},
		{"invalid listen", func(c *Config) { c.Listen = "localhost" }, true},
		{"zero inflight", func(c *Config) { c.MaxInflight = 0 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Defaults()
			tt.mutate(&c)
			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
