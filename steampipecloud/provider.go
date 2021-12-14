package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

// Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Steampipe Cloud authentication token. This is used when connecting to Steampipe Cloud workspaces. You can manage your API tokens from the Settings page for your user account in Steampipe Cloud.",
				DefaultFunc: schema.EnvDefaultFunc("STEAMPIPE_CLOUD_TOKEN", nil),
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"insecure_skip_verify": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Steampipe Cloud host. This is used when connecting to Steampipe Cloud workspaces. The default is cloud.steampipe.io, you only need to set this if you are connecting to a remote Steampipe Cloud database that is NOT hosted in cloud.steampipe.io, such as a dev/test instance.",
				DefaultFunc: schema.EnvDefaultFunc("STEAMPIPE_CLOUD_HOST", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"steampipecloud_connection":           resourceConnection(),
			"steampipecloud_organization":         resourceOrganization(),
			"steampipecloud_workspace":            resourceWorkspace(),
			"steampipecloud_workspace_connection": resourceWorkspaceConnection(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// "steampipecloud_user": dataSourceUser(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{}
	if val, ok := d.GetOk("host"); ok {
		config.Host = val.(string)
	}
	if val, ok := d.GetOk("token"); ok {
		config.Token = val.(string)
	}
	if val, ok := d.GetOk("organization"); ok {
		config.Organization = val.(string)
	}
	if val, ok := d.GetOk("insecure_skip_verify"); ok {
		config.InsecureSkipVerify = val.(bool)
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	apiClient, err := CreateClient(&config, diags)
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
	Organization       string
	Token              string
	Host               string
	InsecureSkipVerify bool
}

/*
	precedence of credentials:
	- token set in config
	- ENV vars {STEAMPIPE_CLOUD_TOKEN}
*/
func CreateClient(config *Config, diags diag.Diagnostics) (*steampipe.APIClient, diag.Diagnostics) {
	configuration := steampipe.NewConfiguration()

	if config.Host != "" {
		configuration.Servers = []steampipe.ServerConfiguration{
			{
				URL: fmt.Sprintf("https://%s/api/v1", config.Host),
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
		return steampipe.NewAPIClient(configuration), diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create HashiCups client",
		Detail:   "Failed to get token to authenticate Steampipecloud client. Please set 'token' in provider config",
	})
	return nil, diags
}
