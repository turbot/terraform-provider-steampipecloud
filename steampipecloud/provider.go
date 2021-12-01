package steampipecloud

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
			"insecure_skip_verify": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"steampipecloud_connection": resourceSteampipeCloudConnection(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"steampipecloud_user": dataSourceSteampipeCloudUser(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	configuration := openapiclient.NewConfiguration()
	config := Config{
		Token:              d.Get("token").(string),
		Org:                d.Get("org").(string),
		InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
		Hostname:           d.Get("hostname").(string),
	}

	// TODO: Write a helper function to extract Token, Handle, Org from env variable and set as per their precedence

	configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", config.Token))
	apiClient := openapiclient.NewAPIClient(configuration)

	log.Println("[INFO] Steampipe cloud API client initialized, now validating...", apiClient)
	return &SteampipeClient{
		APIClient: apiClient,
		Config:    &config,
	}, nil
}

type SteampipeClient struct {
	APIClient *openapiclient.APIClient
	Config    *Config
}

type Config struct {
	Org                string
	Token              string
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
