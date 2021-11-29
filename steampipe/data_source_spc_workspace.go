package steampipe

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	apiClient "github.com/turbot/steampipe-cloud-sdk-go"
)

func dataSourceSteampipeWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSteampipeWorkspaceRead,
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// "type": {
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// },
			// "resource": {
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// },
			// "state": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			// "reason": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			// "details": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
		},
	}
}

func dataSourceSteampipeWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.APIClient)

	resp, r, err := client.UsersWorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet(context.Background(), "terraformtest1234", "subhajit97").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.ActorWorkspaceGet`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

		return err
	}

	d.Set("handle", resp.Handle)

	return nil
}
