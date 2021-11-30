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
		Create: resourceSteampipeCloudWorkspaceCreate,
		Read:   resourceSteampipeCloudWorkspaceRead,
		Delete: resourceSteampipeCloudWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudWorkspaceImport,
		},
		Schema: map[string]*schema.Schema{
			// aka of the parent resource
			"handle": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			// "id": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
		},
	}
}

func resourceSteampipeCloudWorkspaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// if err := resourceSteampipeCloudWorkspaceRead(d, meta); err != nil {
	// 	return nil, err
	// }
	// d.Set("handle", "terraformtest1234")
	d.State().Attributes["handle"] = "terraformtest1234"
	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeCloudWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	req := openapiclient.TypesCreateWorkspaceRequest{
		Handle: "terraformtest1234",
	}
	resp, r, err := client.UsersWorkspacesApi.UserUserHandleWorkspacePost(context.Background(), "subhajit97").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// data := *resp.Items
	d.SetId(resp.Id)
	d.Set("handle", resp.Handle)

	return nil
}

func resourceSteampipeCloudWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	handle := d.State().Attributes["handle"]

	_, r, err := client.UsersWorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet(context.Background(), handle, "subhajit97").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// data := *resp.Items
	d.Set("handle", handle)

	return nil
}

func resourceSteampipeCloudWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	_, r, err := client.UsersWorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete(context.Background(), "terraformtest2021", "subhajit97").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	return nil
}
