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
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"org": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"steampipe_workspace":       resourceSteampipeCloudWorkspace(),
			"steampipe_user_connection": resourceSteampipeUserConnection(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"steampipe_workspace": dataSourceSteampipeWorkspace(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	configuration := openapiclient.NewConfiguration()

	spcToken := d.Get("token").(string)
	configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", spcToken))
	apiClient := openapiclient.NewAPIClient(configuration)

	log.Println("[INFO] Steampipe cloud API client initialized, now validating...", apiClient)
	return apiClient, nil
}
