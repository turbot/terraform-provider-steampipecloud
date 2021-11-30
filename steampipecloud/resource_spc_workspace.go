package steampipe

import (
	"context"
	"fmt"
	"os"

	"github.com/turbot/go-kit/types"

	"github.com/hashicorp/terraform/helper/schema"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeCloudWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeCloudWorkspaceCreate,
		Read:   resourceSteampipeCloudWorkspaceRead,
		Delete: resourceSteampipeCloudWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudWorkspaceImport,
		},
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hive": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"host": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceSteampipeCloudWorkspaceExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*openapiclient.APIClient)

	handle := d.State().Attributes["handle"]
	userHandler := getUserHandler(meta)

	_, r, err := client.UserWorkspacesApi.GetUserWorkspace(context.Background(), handle, userHandler).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
	resp, r, err := client.UserWorkspacesApi.CreateUserWorkspace(context.Background(), "subhajit97").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// data := *resp.Items
	d.SetId(resp.Id)
	d.Set("handle", resp.Handle)
	d.Set("workspace_state", resp.WorkspaceState)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("identity_id", resp.IdentityId)
	d.Set("version_id", resp.VersionId)

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

	d.Set("handle", handle)
	d.Set("workspace_state", resp.WorkspaceState)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("identity_id", resp.IdentityId)
	d.Set("version_id", resp.VersionId)

	return nil
}

func resourceSteampipeCloudWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	handle := d.State().Attributes["handle"]
	userHandler := getUserHandler(meta)

	_, r, err := client.UserWorkspacesApi.DeleteUserWorkspace(context.Background(), handle, userHandler).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	return nil
}
