package steampipe

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeConnectionCreate,
		Read:   resourceSteampipeConnectionRead,
		Delete: resourceSteampipeConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeConnectionImport,
		},
		Schema: map[string]*schema.Schema{
			// aka of the parent resource
			"handle": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSteampipeCloudWorkspaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudWorkspaceRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	req := openapiclient.TypesCreateWorkspaceRequest{
		Handle: "terraformtest1234",
	}
	resp, r, err := client.UserWorkspacesApi.CreateUserWorkspace(context.Background(), "lalit").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}
	// data := *resp.Items
	d.Set("handle", resp.Handle)

	return nil
}

func resourceSteampipeCloudWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	resp, r, err := client.UserWorkspacesApi.GetUserWorkspace(context.Background(), "terraformtest1234", "lalit").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// data := *resp.Items
	d.Set("handle", resp.Handle)

	return nil
}

func resourceSteampipeCloudWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	_, r, err := client.UserWorkspacesApi.DeleteUserWorkspace(context.Background(), "terraformtest2021", "lalit").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	return nil
}
