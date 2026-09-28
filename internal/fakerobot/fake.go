package fakerobot

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

// API is a mutable in-memory implementation of the Robot GET /server endpoint.
type API struct {
	mu       sync.RWMutex
	servers  []robot.Server
	status   int
	code     string
	message  string
	interval int
	requests atomic.Int64
}

func New() *API { return &API{} }

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/server" {
		http.NotFound(w, r)
		return
	}
	a.requests.Add(1)
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if a.status != 0 {
		w.WriteHeader(a.status)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"status": a.status, "code": a.code, "message": a.message, "interval": a.interval}})
		return
	}
	response := make([]map[string]robot.Server, 0, len(a.servers))
	for _, server := range a.servers {
		response = append(response, map[string]robot.Server{"server": server})
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (a *API) SetServers(servers ...robot.Server) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.servers = append([]robot.Server(nil), servers...)
}

func (a *API) SetError(status int, code, message string, interval int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.status, a.code, a.message, a.interval = status, code, message, interval
}

func (a *API) ResetError() { a.SetError(0, "", "", 0) }

func (a *API) RequestCount() int64 { return a.requests.Load() }
