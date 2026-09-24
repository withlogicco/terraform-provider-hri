package main

import (
	"context"
	"flag"
	"log"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/withlogicco/terraform-provider-hri/internal/provider"
)

var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable provider debug mode")
	flag.Parse()
	err := providerserver.Serve(context.Background(), func() frameworkprovider.Provider { return provider.New(version) }, providerserver.ServeOpts{Address: "registry.terraform.io/withlogicco/hri", Debug: debug})
	if err != nil {
		log.Fatal(err)
	}
}
