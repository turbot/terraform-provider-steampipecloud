package main

import (
	"github.com/Subhajit97/terraform-provider-steampipe/steampipe"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: steampipe.Provider})
}
