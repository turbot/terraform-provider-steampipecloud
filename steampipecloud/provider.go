package steampipecloud

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:     schema.TypeString,
				Optional: true,
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
			"steampipecloud_connection":                       resourceSteampipeCloudConnection(),
			"steampipecloud_organization":                     resourceSteampipeCloudOrganization(),
			"steampipecloud_workspace":                        resourceSteampipeCloudWorkspace(),
			"steampipecloud_workspace_connection_association": resourceSteampipeCloudWorkspaceConnectionAssociation(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"steampipecloud_user": dataSourceSteampipeCloudUser(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:              d.Get("token").(string),
		Org:                d.Get("org").(string),
		InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
		Hostname:           d.Get("hostname").(string),
	}

	apiClient, err := CreateClient(&config)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO] Steampipe cloud API client initialized, now validating...", apiClient)
	return &SteampipeClient{
		APIClient: apiClient,
		Config:    &config,
	}, nil
}

type SteampipeClient struct {
	APIClient *steampipe.APIClient
	Config    *Config
}

type Config struct {
	Org                string
	Token              string
	Hostname           string
	InsecureSkipVerify bool
}

/*
	precedence of credentials:
	- token set in config
	- ENV vars {STEAMPIPE_CLOUD_TOKEN}
*/
func CreateClient(config *Config) (*steampipe.APIClient, error) {
	configuration := steampipe.NewConfiguration()

	if config.Hostname != "" {
		configuration.Servers = []steampipe.ServerConfiguration{
			{
				URL: config.Hostname,
			},
		}
	}
	var steampipeCloudToken string
	if config.Token != "" {
		steampipeCloudToken = config.Token
	} else {
		// return nil, fmt.Errorf("failed to get token to authenticate. Please set 'token' in provider to config. STEAMPIPE_CLOUD_TOKEN")
		if token, ok := os.LookupEnv("STEAMPIPE_CLOUD_TOKEN"); ok {
			steampipeCloudToken = token
		}
	}
	if steampipeCloudToken != "" {
		configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", steampipeCloudToken))
		return steampipe.NewAPIClient(configuration), nil
	}

	return nil, fmt.Errorf("failed to get token to authenticate. Please set 'token' in provider config")
}
