package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

type mutableRobot struct {
	mu      sync.RWMutex
	servers []robot.Server
}

func (m *mutableRobot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	servers := append([]robot.Server(nil), m.servers...)
	m.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	out := make([]map[string]robot.Server, 0, len(servers))
	for _, s := range servers {
		out = append(out, map[string]robot.Server{"server": s})
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (m *mutableRobot) set(servers ...robot.Server) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
}

func accConfig(url string, number int) string {
	return fmt.Sprintf("terraform {\n  required_providers {\n    hri = { source = \"withlogicco/hri\" }\n  }\n}\nprovider \"hri\" {\n  username = \"u\"\n  password = \"p\"\n  base_url = %q\n}\nresource \"hri_server\" \"db\" {\n  server_number = %d\n}\n", url, number)
}

func emptyConfig(url string) string {
	return fmt.Sprintf("terraform {\n  required_providers {\n    hri = { source = \"withlogicco/hri\" }\n  }\n}\nprovider \"hri\" {\n  username = \"u\"\n  password = \"p\"\n  base_url = %q\n}\n", url)
}

func factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){"hri": providerserver.NewProtocol6WithError(New("test"))}
}

func TestAccServerResourceLifecycle(t *testing.T) {
	api := &mutableRobot{}
	api.set(robot.Server{ServerNumber: 42, ServerName: "db-old", ServerIP: "192.0.2.1", IPs: []string{"192.0.2.1"}})
	defer api.set()
	server := httptest.NewServer(api)
	defer server.Close()
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: factories(), Steps: []resource.TestStep{
		{Config: accConfig(server.URL, 42), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("hri_server.db", "server_number", "42"), resource.TestCheckResourceAttr("hri_server.db", "server_name", "db-old"))},
		{PreConfig: func() { api.set(robot.Server{ServerNumber: 42, ServerName: "db-new", ServerIP: "192.0.2.2"}) }, Config: accConfig(server.URL, 42), Check: resource.TestCheckResourceAttr("hri_server.db", "server_name", "db-new")},
		{ResourceName: "hri_server.db", ImportState: true, ImportStateId: "42", ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "server_number"},
		{PreConfig: func() {
			api.set(robot.Server{ServerNumber: 42, ServerName: "db-new"}, robot.Server{ServerNumber: 43, ServerName: "next"})
		}, Config: accConfig(server.URL, 43), ExpectError: regexp.MustCompile("Cannot replace server")},
		{PreConfig: func() { api.set() }, Config: emptyConfig(server.URL)},
	}})
}

func TestAccServerResourceRejectsAbsentServer(t *testing.T) {
	api := &mutableRobot{}
	server := httptest.NewServer(api)
	defer server.Close()
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: factories(), Steps: []resource.TestStep{{Config: accConfig(server.URL, 44), ExpectError: regexp.MustCompile("Server not found")}}})
}

func TestAccServerResourceRejectsDestroyWhilePresent(t *testing.T) {
	api := &mutableRobot{}
	api.set(robot.Server{ServerNumber: 45, ServerName: "still-here"})
	defer api.set()
	server := httptest.NewServer(api)
	defer server.Close()
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: factories(), Steps: []resource.TestStep{
		{Config: accConfig(server.URL, 45)},
		{Config: emptyConfig(server.URL), ExpectError: regexp.MustCompile("Cannot destroy server")},
		{PreConfig: func() { api.set() }, Config: emptyConfig(server.URL)},
	}})
}

func TestAccServerResourceRejectsServerDisappearance(t *testing.T) {
	api := &mutableRobot{}
	api.set(robot.Server{ServerNumber: 46, ServerName: "will-disappear"})
	defer api.set()
	server := httptest.NewServer(api)
	defer server.Close()
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: factories(), Steps: []resource.TestStep{
		{Config: accConfig(server.URL, 46)},
		{PreConfig: func() { api.set() }, Config: accConfig(server.URL, 46), ExpectError: regexp.MustCompile("Server not found")},
	}})
}
