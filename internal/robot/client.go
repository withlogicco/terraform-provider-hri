package robot

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	Status        int
	Code, Message string
	RetryAfter    time.Duration
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Robot API HTTP %d (%s): %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("Robot API HTTP %d: %s", e.Status, e.Message)
}

type envelope struct {
	Status   int       `json:"status"`
	Code     string    `json:"code"`
	Message  string    `json:"message"`
	Interval int       `json:"interval"`
	Error    *envelope `json:"error"`
}
type serverEnvelope struct {
	Server Server `json:"server"`
}
type subnet struct {
	IP   string `json:"ip"`
	Mask string `json:"mask"`
}
type Server struct {
	ServerNumber  int      `json:"server_number"`
	ServerName    string   `json:"server_name"`
	ServerIP      string   `json:"server_ip"`
	ServerIPv6Net string   `json:"server_ipv6_net"`
	Product       string   `json:"product"`
	DC            string   `json:"dc"`
	Traffic       string   `json:"traffic"`
	Status        string   `json:"status"`
	Cancelled     bool     `json:"cancelled"`
	PaidUntil     string   `json:"paid_until"`
	IPs           []string `json:"ip"`
	Subnets       []string `json:"-"`
}

func (s *Server) UnmarshalJSON(data []byte) error {
	type rawServer struct {
		ServerNumber  int      `json:"server_number"`
		ServerName    string   `json:"server_name"`
		ServerIP      string   `json:"server_ip"`
		ServerIPv6Net string   `json:"server_ipv6_net"`
		Product       string   `json:"product"`
		DC            string   `json:"dc"`
		Traffic       string   `json:"traffic"`
		Status        string   `json:"status"`
		Cancelled     bool     `json:"cancelled"`
		PaidUntil     string   `json:"paid_until"`
		IPs           []string `json:"ip"`
		Subnets       []subnet `json:"subnet"`
	}
	var v rawServer
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*s = Server{ServerNumber: v.ServerNumber, ServerName: v.ServerName, ServerIP: v.ServerIP, ServerIPv6Net: v.ServerIPv6Net, Product: v.Product, DC: v.DC, Traffic: v.Traffic, Status: v.Status, Cancelled: v.Cancelled, PaidUntil: v.PaidUntil, IPs: v.IPs}
	for _, n := range v.Subnets {
		s.Subnets = append(s.Subnets, n.IP+"/"+n.Mask)
	}
	return nil
}

type Client struct {
	base       string
	user, pass string
	http       *http.Client
	sleep      func(time.Duration)
}

func NewClient(user, pass, base string) (*Client, error) {
	u, e := url.Parse(base)
	if e != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base URL must be absolute")
	}
	return &Client{base: strings.TrimRight(base, "/"), user: user, pass: pass, http: &http.Client{Timeout: 30 * time.Second}, sleep: time.Sleep}, nil
}
func (c *Client) ListServers() ([]Server, error) {
	rateRetried := false
	var last error
	for attempt := 0; attempt < 3; {
		req, e := http.NewRequest("GET", c.base+"/server", nil)
		if e != nil {
			return nil, e
		}
		req.SetBasicAuth(c.user, c.pass)
		res, e := c.http.Do(req)
		if e != nil {
			last = e
			attempt++
			if attempt < 3 {
				c.sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
			}
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
		res.Body.Close()
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			var wrapped []serverEnvelope
			if e = json.Unmarshal(body, &wrapped); e != nil {
				return nil, fmt.Errorf("decode Robot inventory: %w", e)
			}
			servers := make([]Server, 0, len(wrapped))
			for _, item := range wrapped {
				servers = append(servers, item.Server)
			}
			return servers, nil
		}
		var env envelope
		_ = json.Unmarshal(body, &env)
		if env.Error != nil {
			env = *env.Error
		}
		api := &APIError{Status: res.StatusCode, Code: env.Code, Message: env.Message}
		if res.StatusCode == http.StatusUnauthorized {
			return nil, api
		}
		if res.StatusCode == http.StatusForbidden && env.Code == "RATE_LIMIT_EXCEEDED" && !rateRetried {
			rateRetried = true
			d := time.Duration(env.Interval) * time.Second
			if h := res.Header.Get("Retry-After"); h != "" {
				if n, e := strconv.Atoi(h); e == nil {
					d = time.Duration(n) * time.Second
				}
			}
			if d > 5*time.Second {
				d = 5 * time.Second
			}
			if d > 0 {
				c.sleep(d)
			}
			continue
		}
		if res.StatusCode == http.StatusServiceUnavailable {
			last = api
			attempt++
			if attempt < 3 {
				c.sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
				continue
			}
			return nil, api
		}
		return nil, api
	}
	return nil, last
}
