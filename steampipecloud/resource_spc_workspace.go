package steampipe

import (
	"context"
	"fmt"
	"log"

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
<<<<<<< HEAD:steampipecloud/resource_spc_workspace.go

=======
	handle := d.Get("handle")

	// Empty check
	if handle.(string) == "" {
		return fmt.Errorf("handle can not be empty")
	}

	// Create request
>>>>>>> a8895d8 (Fix update to store all properties in state file):steampipe/resource_spc_workspace.go
	req := openapiclient.TypesCreateWorkspaceRequest{
		Handle: "terraformtest1234",
	}
<<<<<<< HEAD:steampipecloud/resource_spc_workspace.go
	resp, r, err := client.UserWorkspacesApi.CreateUserWorkspace(context.Background(), "subhajit97").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// data := *resp.Items
=======

	// Get current user/org information
	userHandler := getUserHandler(meta)

	// Create resource
	resp, _, err := client.UserWorkspacesApi.CreateUserWorkspace(context.Background(), userHandler).Request(req).Execute()
	if err != nil {
		return fmt.Errorf("error creating workspace: %s", err)
	}
	log.Printf("\n[DEBUG] Workspace created: %s", resp.Handle)

	// Set property values
>>>>>>> a8895d8 (Fix update to store all properties in state file):steampipe/resource_spc_workspace.go
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
<<<<<<< HEAD:steampipecloud/resource_spc_workspace.go
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
=======
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Workspace (%s) not found", handle)
			d.SetId("")
			return nil
		}
		log.Printf("\n[DEBUG] Workspace received: %s", resp.Handle)
>>>>>>> a8895d8 (Fix update to store all properties in state file):steampipe/resource_spc_workspace.go
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

<<<<<<< HEAD:steampipecloud/resource_spc_workspace.go
=======
func resourceSteampipeCloudWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	oldHandle, newHandle := d.GetChange("handle")

	if newHandle.(string) == "" {
		return fmt.Errorf("handle can not be empty")
	}

	// Get user handler
	userHandler := getUserHandler(meta)

	// Create request
	req := openapiclient.TypesUpdateWorkspaceRequest{
		Handle: types.String(newHandle.(string)),
	}
	log.Printf("\n[DEBUG] Updating Workspace: %s", *req.Handle)

	resp, _, err := client.UserWorkspacesApi.UpdateUserWorkspace(context.Background(), userHandler, oldHandle.(string)).Request(req).Execute()
	if err != nil {
		return fmt.Errorf("error updating workspace: %s", err)
	}
	log.Printf("\n[DEBUG] Workspace updated: %s", resp.Handle)

	// Update state file
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

>>>>>>> a8895d8 (Fix update to store all properties in state file):steampipe/resource_spc_workspace.go
func resourceSteampipeCloudWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	handle := d.State().Attributes["handle"]
	userHandler := getUserHandler(meta)

<<<<<<< HEAD:steampipecloud/resource_spc_workspace.go
	_, r, err := client.UserWorkspacesApi.DeleteUserWorkspace(context.Background(), handle, userHandler).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
=======
	// Empty check
	if handle == "" {
		return fmt.Errorf("handle can not be empty")
	}
	log.Printf("\n[DEBUG] Deleting Workspace: %s", handle)

	_, _, err := client.UserWorkspacesApi.DeleteUserWorkspace(context.Background(), handle, userHandler).Execute()
	if err != nil {
		return fmt.Errorf("error deleting workspace: %s", err)
>>>>>>> a8895d8 (Fix update to store all properties in state file):steampipe/resource_spc_workspace.go
	}

	return nil
}
<<<<<<< HEAD:steampipecloud/resource_spc_workspace.go
=======

func getUserHandler(meta interface{}) string {
	client := meta.(*openapiclient.APIClient)
	resp, _, err := client.UsersApi.GetActor(context.Background()).Execute()
	if err != nil {
		return ""
	}
	return resp.Handle
}
>>>>>>> a8895d8 (Fix update to store all properties in state file):steampipe/resource_spc_workspace.go
