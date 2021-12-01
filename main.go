package main

import (
	"github.com/Subhajit97/terraform-provider-steampipe/steampipecloud"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: steampipecloud.Provider})
}
