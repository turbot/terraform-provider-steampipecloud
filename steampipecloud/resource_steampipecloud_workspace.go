package steampipecloud

import (
	"context"
	"fmt"
	"log"
	_nethttp "net/http"
	"regexp"

	"github.com/turbot/go-kit/types"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	openapiclient "github.com/turbot/steampipecloud-sdk-go"
)

func resourceSteampipeCloudWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeCloudWorkspaceCreate,
		Read:   resourceSteampipeCloudWorkspaceRead,
		Delete: resourceSteampipeCloudWorkspaceDelete,
		Update: resourceSteampipeCloudWorkspaceUpdate,
		Exists: resourceSteampipeCloudWorkspaceExists,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudWorkspaceImport,
		},
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "must satisfy regular expression pattern: ^[a-z0-9]{1,23}$"),
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"hive": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"host": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceSteampipeCloudWorkspaceExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*SteampipeClient)
	handle := d.Id()

	var err error
	var r *_nethttp.Response

	if client.Config != nil && client.Config.Org != "" {
		_, r, err = client.APIClient.OrgWorkspacesApi.GetOrgWorkspace(context.Background(), client.Config.Org, handle).Execute()
	} else {
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return false, fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		_, r, err = client.APIClient.UserWorkspacesApi.GetUserWorkspace(context.Background(), userHandler, handle).Execute()
	}

	// Error check
	if err != nil {
		if r.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceSteampipeCloudWorkspaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudWorkspaceRead(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeCloudWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)
	handle := d.Get("handle")

	// Create request
	req := openapiclient.TypesCreateWorkspaceRequest{
		Handle: handle.(string),
	}

	var resp openapiclient.TypesWorkspace
	var err error

	// Check for organization
	// If 'org' is set in the provider config, workspace will create in that organization
	// else, the user identity is used.
	if client.Config != nil && client.Config.Org != "" {
		resp, _, err = client.APIClient.OrgWorkspacesApi.CreateOrgWorkspace(context.Background(), client.Config.Org).Request(req).Execute()
	} else {
		// Get current actor information
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		resp, _, err = client.APIClient.UserWorkspacesApi.CreateUserWorkspace(context.Background(), userHandler).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error creating workspace: %s", err)
	}
	log.Printf("\n[DEBUG] Workspace created: %s", resp.Handle)

	// Set property values
	d.SetId(resp.Handle)
	d.Set("handle", resp.Handle)
	d.Set("workspace_id", resp.Id)
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
	client := meta.(*SteampipeClient)
	handle := d.Id()

	var resp openapiclient.TypesWorkspace
	var err error
	var r *_nethttp.Response

	if client.Config != nil && client.Config.Org != "" {
		resp, r, err = client.APIClient.OrgWorkspacesApi.GetOrgWorkspace(context.Background(), client.Config.Org, handle).Execute()
	} else {
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		resp, r, err = client.APIClient.UserWorkspacesApi.GetUserWorkspace(context.Background(), userHandler, handle).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Workspace (%s) not found", handle)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading %s: %s", handle, err)
	}
	log.Printf("\n[DEBUG] Workspace received: %s", resp.Handle)

	d.Set("handle", handle)
	d.Set("workspace_id", resp.Id)
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

func resourceSteampipeCloudWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)

	oldHandle, newHandle := d.GetChange("handle")

	// Create request
	req := openapiclient.TypesUpdateWorkspaceRequest{
		Handle: types.String(newHandle.(string)),
	}
	log.Printf("\n[DEBUG] Updating Workspace: %s", *req.Handle)

	var resp openapiclient.TypesWorkspace
	var err error

	if client.Config != nil && client.Config.Org != "" {
		resp, _, err = client.APIClient.OrgWorkspacesApi.UpdateOrgWorkspace(context.Background(), client.Config.Org, oldHandle.(string)).Request(req).Execute()
	} else {
		// Get user handler
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		resp, _, err = client.APIClient.UserWorkspacesApi.UpdateUserWorkspace(context.Background(), userHandler, oldHandle.(string)).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error updating workspace: %s", err)
	}
	log.Printf("\n[DEBUG] Workspace updated: %s", resp.Handle)

	// Update state file
	d.SetId(resp.Handle)
	d.Set("handle", resp.Handle)
	d.Set("workspace_id", resp.Id)
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
	client := meta.(*SteampipeClient)
	handle := d.Id()
	log.Printf("\n[DEBUG] Deleting Workspace: %s", handle)

	var err error
	if client.Config != nil && client.Config.Org != "" {
		_, _, err = client.APIClient.OrgWorkspacesApi.DeleteOrgWorkspace(context.Background(), client.Config.Org, handle).Execute()
	} else {
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		_, _, err = client.APIClient.UserWorkspacesApi.DeleteUserWorkspace(context.Background(), userHandler, handle).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error deleting workspace: %s", err)
	}
	d.SetId("")

	return nil
}
