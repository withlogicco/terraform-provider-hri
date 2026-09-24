package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

func TestAccServersDataSourceUsesSingleInventoryRequest(t *testing.T) {
	servers := make([]map[string]robot.Server, 50)
	for i := range servers {
		servers[i] = map[string]robot.Server{"server": {ServerNumber: 1000 + i, ServerName: fmt.Sprintf("db-%02d", i), DC: "NBG1-DC1", Product: "DS 3000"}}
	}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servers)
	}))
	defer httpServer.Close()
	config := fmt.Sprintf("terraform {\n required_providers {\n hri = { source = \"withlogicco/hri\" }\n }\n}\nprovider \"hri\" {\n username = \"u\"\n password = \"p\"\n base_url = %q\n}\ndata \"hri_servers\" \"all\" {\n dc = \"NBG1-DC1\"\n product = \"DS 3000\"\n name_regex = \"^db-\"\n}\ndata \"hri_server\" \"first\" { server_number = 1001 }\n", httpServer.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"hri": providerserver.NewProtocol6WithError(New("test"))},
		Steps: []resource.TestStep{{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.hri_servers.all", "server_numbers.#", "50"),
			resource.TestCheckResourceAttr("data.hri_server.first", "server_name", "db-01"),
		)}},
	})
}
