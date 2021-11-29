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
				Type:     schema.TypeString,
				Required: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"spc_workspace": resourceSteampipeCloudWorkspace(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	configuration := openapiclient.NewConfiguration()

	spCloudToken := d.Get("token").(string)
	configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", spCloudToken))
	apiClient := openapiclient.NewAPIClient(configuration)

	log.Println("[INFO] Steampipe cloud API client initialized, now validating...", apiClient)
	return apiClient, nil
}
