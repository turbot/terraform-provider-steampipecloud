package steampipecloud

import (
	"context"
	"fmt"
	"log"
	_nethttp "net/http"
	"regexp"

	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
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
	var userHandler string
	var r *_nethttp.Response

	if client.Config != nil && client.Config.Org != "" {
		_, r, err = client.APIClient.OrgWorkspaces.Get(context.Background(), client.Config.Org, handle).Execute()
	} else {
		userHandler, _, err = getUserHandler(client)
		if err != nil {
			return false, fmt.Errorf("inside resourceSteampipeCloudWorkspaceExists.\ngetHandler Error: \n%v", err)
		}
		_, r, err = client.APIClient.UserWorkspaces.Get(context.Background(), userHandler, handle).Execute()
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
	req := steampipe.TypesCreateWorkspaceRequest{
		Handle: handle.(string),
	}

	var resp steampipe.TypesWorkspace
	var userHandler string
	var err error

	// Check for organization
	// If 'org' is set in the provider config, workspace will create in that organization
	// else, the user identity is used.
	if client.Config != nil && client.Config.Org != "" {
		resp, _, err = client.APIClient.OrgWorkspaces.Create(context.Background(), client.Config.Org).Request(req).Execute()
	} else {
		// Get current actor information
		userHandler, _, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceCreate.\ngetHandler Error: \n%v", err)
		}
		resp, _, err = client.APIClient.UserWorkspaces.Create(context.Background(), userHandler).Request(req).Execute()
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

	var resp steampipe.TypesWorkspace
	var userHandler string
	var err error
	var r *_nethttp.Response

	if client.Config != nil && client.Config.Org != "" {
		resp, r, err = client.APIClient.OrgWorkspaces.Get(context.Background(), client.Config.Org, handle).Execute()
	} else {
		userHandler, _, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceRead.\ngetHandler Error: \n%v", err)
		}
		resp, r, err = client.APIClient.UserWorkspaces.Get(context.Background(), userHandler, handle).Execute()
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
	req := steampipe.TypesUpdateWorkspaceRequest{
		Handle: types.String(newHandle.(string)),
	}
	log.Printf("\n[DEBUG] Updating Workspace: %s", *req.Handle)

	var resp steampipe.TypesWorkspace
	var userHandler string
	var err error

	if client.Config != nil && client.Config.Org != "" {
		resp, _, err = client.APIClient.OrgWorkspaces.Update(context.Background(), client.Config.Org, oldHandle.(string)).Request(req).Execute()
	} else {
		// Get user handler
		userHandler, _, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceUpdate.\ngetHandler Error: \n%v", err)
		}
		resp, _, err = client.APIClient.UserWorkspaces.Update(context.Background(), userHandler, oldHandle.(string)).Request(req).Execute()
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
	var userHandler string
	if client.Config != nil && client.Config.Org != "" {
		_, _, err = client.APIClient.OrgWorkspaces.Delete(context.Background(), client.Config.Org, handle).Execute()
	} else {
		userHandler, _, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceDelete.\ngetHandler Error: \n%v", err)
		}
		_, _, err = client.APIClient.UserWorkspaces.Delete(context.Background(), userHandler, handle).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error deleting workspace: %s", err)
	}
	d.SetId("")

	return nil
}
