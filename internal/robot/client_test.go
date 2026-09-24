package robot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestListServersRobotEnvelopeAndBasicAuth(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		u, p, ok := r.BasicAuth()
		if !ok || u != "user" || p != "pass" {
			t.Errorf("bad auth: %q %q %v", u, p, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"server":{"server_number":42,"server_name":"db","ip":["192.0.2.1"],"subnet":[{"ip":"203.0.113.0","mask":"24"}]}}]`))
	}))
	defer s.Close()
	c, _ := NewClient("user", "pass", s.URL)
	items, e := c.ListServers()
	if e != nil {
		t.Fatal(e)
	}
	if calls != 1 || len(items) != 1 || items[0].ServerNumber != 42 || items[0].ServerName != "db" || items[0].IPs[0] != "192.0.2.1" || items[0].Subnets[0] != "203.0.113.0/24" {
		t.Fatalf("unexpected inventory %#v calls=%d", items, calls)
	}
}
func TestListServersAuthFailsOnce(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error":{"status":401,"code":"AUTH_FAILED","message":"no"}}`))
	}))
	defer s.Close()
	c, _ := NewClient("u", "p", s.URL)
	_, e := c.ListServers()
	if e == nil || calls != 1 || !strings.Contains(e.Error(), "401") || !strings.Contains(e.Error(), "AUTH_FAILED") {
		t.Fatalf("err=%v calls=%d", e, calls)
	}
}
func TestRateLimitRetryOnceAndFailure(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"slow","interval":2}}`))
	}))
	defer s.Close()
	c, _ := NewClient("u", "p", s.URL)
	waits := 0
	c.sleep = func(d time.Duration) {
		waits++
		if d != 2*time.Second {
			t.Errorf("waited %s", d)
		}
	}
	_, e := c.ListServers()
	if e == nil || calls != 2 || waits != 1 {
		t.Fatalf("err=%v calls=%d waits=%d", e, calls, waits)
	}
}
func TestMaintenanceRetriesThreeTimes(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"error":{"code":"MAINTENANCE","message":"later"}}`))
	}))
	defer s.Close()
	c, _ := NewClient("u", "p", s.URL)
	c.sleep = func(time.Duration) {}
	_, e := c.ListServers()
	if e == nil || calls != 3 {
		t.Fatalf("err=%v calls=%d", e, calls)
	}
}
