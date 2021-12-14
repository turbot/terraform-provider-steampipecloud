package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/turbot/terraform-provider-steampipecloud/steampipecloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: steampipecloud.Provider})
}
