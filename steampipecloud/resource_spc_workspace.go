package steampipe

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeCloudWorkspace() *schema.Resource {
	return &schema.Resource{
		//Create: resourceTurbotFolderCreate,
		Read: resourceTurbotFolderRead,
		// Update: resourceTurbotFolderUpdate,
		// Delete: resourceTurbotFolderDelete,
		// Exists: resourceTurbotFolderExists,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceTurbotFolderImport,
		// },
		Schema: map[string]*schema.Schema{
			// aka of the parent resource
			"handle": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceTurbotFolderRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	resp, r, err := client.UsersWorkspacesApi.ActorWorkspaceGet(context.Background()).Limit(int32(20)).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.ActorWorkspaceGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	data := *resp.Items
	d.Set("resource", data[0].Handle)

	return nil
}
