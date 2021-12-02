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
		DataSourcesMap: map[string]*schema.Resource{},

		ConfigureFunc: providerConfigure,
	}
}

type ProviderConfig = struct {
	Hostname           string
	InsecureSkipVerify bool
	Org                string
	Token              string
}

type SteampipeCloudClient = struct {
	APIClient *openapiclient.APIClient
	Config    *ProviderConfig
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &ProviderConfig{
		Hostname:           d.Get("hostname").(string),
		InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
		Org:                d.Get("org").(string),
		Token:              d.Get("token").(string),
	}
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
