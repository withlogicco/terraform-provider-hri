package robot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIndexFetchesOnce(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		servers := make([]map[string]Server, 50)
		for n := range servers {
			servers[n] = map[string]Server{"server": {ServerNumber: n + 1}}
		}
		_ = json.NewEncoder(w).Encode(servers)
	}))
	defer s.Close()
	c, _ := NewClient("u", "p", s.URL)
	idx := NewIndex(c)
	if _, _, e := idx.Get(1); e != nil {
		t.Fatal(e)
	}
	if _, _, e := idx.Get(50); e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}
