package steampipe

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("STEAMPIPE_CLOUD_TOKEN", ""),
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"insecure_skip_verify": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"steampipe_workspace":       resourceSteampipeCloudWorkspace(),
			"steampipe_user_connection": resourceSteampipeUserConnection(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"steampipe_workspace": dataSourceSteampipeWorkspace(),
			"steampipe_user":      dataSourceSteampipeUser(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	configuration := openapiclient.NewConfiguration()
	config := Config{
		Token:              d.Get("token").(string),
		Org:                d.Get("org").(string),
		Handle:             d.Get("handle").(string),
		InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
		Hostname:           d.Get("hostname").(string),
	}

	configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", config.Token))
	apiClient := openapiclient.NewAPIClient(configuration)

	log.Println("[INFO] Steampipe cloud API client initialized, now validating...", apiClient)
	return apiClient, nil
}

type Config struct {
	Org                string
	Token              string
	Handle             string
	Hostname           string
	InsecureSkipVerify bool
}

// provider "steampipecloud" {
//   org   = "acme"
//   token = "spt_example"
// }

// provider "steampipecloud" {
//   alias = "turbie"
//   # Token is for the user turbie
//   token = "spt_example"
// }

// provider "steampipecloud" {
//   alias = "foo"
//   org   = "foo"
//   token = "spt_example"
// }

// resource "steampipecloud_workspace" "orgdev" {
//   # uses default provider, for org acme
//   handle = "dev"
// }

// resource "steampipecloud_workspace" "userdev" {
//   provider = steampipecloud.turbie
//   handle = "dev"
// }

// resource "steampipecloud_workspace" "orgfoodev" {
//   provider = steampipecloud.foo
//   handle = "dev"
// }
