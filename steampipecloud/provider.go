package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
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
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Steampipe Cloud host. This is used when connecting to Steampipe Cloud workspaces. The default is https://cloud.steampipe.io, you only need to set this if you are connecting to a remote Steampipe Cloud database that is NOT hosted in https://cloud.steampipe.io, such as a dev/test instance.",
				DefaultFunc: schema.EnvDefaultFunc("STEAMPIPE_CLOUD_HOST", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"steampipecloud_connection":                    resourceConnection(),
			"steampipecloud_organization":                  resourceOrganization(),
			"steampipecloud_organization_member":           resourceOrganizationMember(),
			"steampipecloud_organization_workspace_member": resourceOrganizationWorkspaceMember(),
			"steampipecloud_user_preferences":              resourceUserPreferences(),
			"steampipecloud_workspace":                     resourceWorkspace(),
			"steampipecloud_workspace_connection":          resourceWorkspaceConnection(),
			"steampipecloud_workspace_mod":                 resourceWorkspaceMod(),
			"steampipecloud_workspace_mod_variable":        resourceWorkspaceModVariable(),
			"steampipecloud_workspace_snapshot":            resourceWorkspaceSnapshot(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"steampipecloud_organization": dataSourceOrganization(),
			"steampipecloud_user":         dataSourceUser(),
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
	Token string
	Host  string
	// InsecureSkipVerify bool
}

/*
precedence of credentials:
1. token set in config
2. ENV vars {STEAMPIPE_CLOUD_TOKEN}
*/
func CreateClient(config *Config, diags diag.Diagnostics) (*steampipe.APIClient, diag.Diagnostics) {
	configuration := steampipe.NewConfiguration()
	if config.Host != "" {
		parsedAPIURL, parseErr := url.Parse(config.Host)
		if parseErr != nil {
			return nil, diag.Errorf(`invalid host: %v`, parseErr)
		}
		if parsedAPIURL.Host == "" {
			return nil, diag.Errorf(`missing protocol or host : %v`, config.Host)
		}
		configuration.Servers = []steampipe.ServerConfiguration{
			{
				URL: fmt.Sprintf("https://%s/api/v0", parsedAPIURL.Host),
			},
		}
	}

	var steampipeCloudToken string
	if config.Token != "" {
		steampipeCloudToken = config.Token
	} else {
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
		Summary:  "Unable to create Steampipe Cloud client",
		Detail:   "Failed to get token to authenticate Steampipe Cloud client. Please set 'token' in provider config",
	})
	return nil, diags
}
