package steampipecloud

import (
	"context"
	"fmt"
	"log"
	_nethttp "net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeCloudWorkspaceConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeCloudWorkspaceConnectionCreate,
		Read:   resourceSteampipeCloudWorkspaceConnectionRead,
		Delete: resourceSteampipeCloudWorkspaceConnectionDelete,
		Update: resourceSteampipeCloudWorkspaceConnectionUpdate,
		Exists: resourceSteampipeCloudWorkspaceConnectionExists,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudWorkspaceConnectionImport,
		},
		Schema: map[string]*schema.Schema{
			"connection_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_config": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"workspace_created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_hive": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_public_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"workspace_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSteampipeCloudWorkspaceConnectionExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	var r *_nethttp.Response
	var err error
	var userHandler string
	client := meta.(*SteampipeClient)

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return false, fmt.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}
	workspaceHandle := idParts[0]
	connHandle := idParts[1]

	if client.Config != nil && client.Config.Org != "" {
		_, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(context.Background(), client.Config.Org, workspaceHandle, connHandle).Execute()
	} else {
		userHandler, r, err = getUserHandler(client)
		if err != nil {
			return false, fmt.Errorf("inside resourceSteampipeCloudWorkspaceConnectionExists.\ngetHandler Error:	\nstatus_code: %d\n	body: %v", r.StatusCode, r.Body)
		}
		_, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(context.Background(), userHandler, workspaceHandle, connHandle).Execute()
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

func resourceSteampipeCloudWorkspaceConnectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudWorkspaceConnectionRead(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeCloudWorkspaceConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	var r *_nethttp.Response
	client := meta.(*SteampipeClient)
	var resp steampipe.TypesWorkspaceConn
	var userHandler string
	var err error
	workspaceHandle := d.Get("workspace_handle").(string)
	connHandle := d.Get("connection_handle").(string)

	// Create request
	req := steampipe.TypesCreateWorkspaceConnRequest{
		ConnectionHandle: connHandle,
	}

	// Check for organization
	// If 'org' is set in the provider config, workspace will create in that organization
	// else, the user identity is used.
	if client.Config != nil && client.Config.Org != "" {
		resp, _, err = client.APIClient.OrgWorkspaceConnectionAssociations.Create(context.Background(), client.Config.Org, workspaceHandle).Request(req).Execute()
	} else {
		// Get current actor information
		userHandler, r, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceConnectionCreate.\ngetHandler Error:	\nstatus_code: %d\n	body: %v", r.StatusCode, r.Body)
		}
		resp, _, err = client.APIClient.UserWorkspaceConnectionAssociations.Create(context.Background(), userHandler, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error creating workspace connection association: %s", err)
	}
	log.Printf("\n[DEBUG] Workspace Connection Association created: %s", resp.Id)

	// Set property values
	id := fmt.Sprintf("%s/%s", workspaceHandle, resp.Connection.Handle)
	d.SetId(id)
	d.Set("association_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("connection_handle", resp.Connection.Handle)
	d.Set("connection_created_at", resp.Connection.CreatedAt)
	d.Set("connection_updated_at", resp.Connection.UpdatedAt)
	d.Set("connection_identity_id", resp.Connection.IdentityId)
	d.Set("connection_plugin", resp.Connection.Plugin)
	d.Set("connection_type", resp.Connection.Type)
	d.Set("connection_version_id", resp.Connection.VersionId)
	d.Set("connection_config", resp.Connection.Config)

	if resp.Workspace != nil {
		d.Set("workspace_state", resp.Workspace.WorkspaceState)
		d.Set("workspace_created_at", resp.Workspace.CreatedAt)
		d.Set("workspace_database_name", resp.Workspace.DatabaseName)
		d.Set("workspace_hive", resp.Workspace.Hive)
		d.Set("workspace_host", resp.Workspace.Host)
		d.Set("workspace_identity_id", resp.Workspace.IdentityId)
		d.Set("workspace_public_key", resp.Workspace.PublicKey)
		d.Set("workspace_updated_at", resp.Workspace.UpdatedAt)
		d.Set("workspace_version_id", resp.Workspace.VersionId)
	}

	return nil
}

func resourceSteampipeCloudWorkspaceConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return fmt.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}

	workspaceHandle := idParts[0]
	connHandle := idParts[1]

	var resp steampipe.TypesWorkspaceConn
	var err error
	var r *_nethttp.Response
	var userHandle string

	if client.Config != nil && client.Config.Org != "" {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(context.Background(), client.Config.Org, workspaceHandle, connHandle).Execute()
	} else {
		userHandle, r, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceConnectionRead.\ngetHandler Error:	\nstatus_code: %d\n	body: %v", r.StatusCode, r.Body)
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(context.Background(), userHandle, workspaceHandle, connHandle).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Association (%s) not found", resp.Id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("inside resourceSteampipeCloudWorkspaceConnectionRead.\nGetUserWorkspaceConnectionAssociation Error:	\nstatus_code: %d\n	body: %v", r.StatusCode, r.Body)
	}
	log.Printf("\n[DEBUG] Association received: %s", resp.Id)

	d.Set("association_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("connection_handle", resp.Connection.Handle)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("connection_created_at", resp.Connection.CreatedAt)
	d.Set("connection_updated_at", resp.Connection.UpdatedAt)
	d.Set("connection_identity_id", resp.Connection.IdentityId)
	d.Set("connection_plugin", resp.Connection.Plugin)
	d.Set("connection_type", resp.Connection.Type)
	d.Set("connection_version_id", resp.Connection.VersionId)
	d.Set("connection_config", resp.Connection.Config)

	if resp.Workspace != nil {
		d.Set("workspace_state", resp.Workspace.WorkspaceState)
		d.Set("workspace_created_at", resp.Workspace.CreatedAt)
		d.Set("workspace_database_name", resp.Workspace.DatabaseName)
		d.Set("workspace_hive", resp.Workspace.Hive)
		d.Set("workspace_host", resp.Workspace.Host)
		d.Set("workspace_identity_id", resp.Workspace.IdentityId)
		d.Set("workspace_public_key", resp.Workspace.PublicKey)
		d.Set("workspace_updated_at", resp.Workspace.UpdatedAt)
		d.Set("workspace_version_id", resp.Workspace.VersionId)
	}

	return nil
}

func resourceSteampipeCloudWorkspaceConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	workspaceHandle := d.State().Attributes["workspace_handle"]
	connHandle := d.State().Attributes["connection_handle"]

	if d.HasChange("workspace_handle") {
		_, newWorkspaceHandle := d.GetChange("workspace_handle")
		workspaceHandle = newWorkspaceHandle.(string)
	}
	if d.HasChange("connection_handle") {
		_, newConnHandle := d.GetChange("connection_handle")
		connHandle = newConnHandle.(string)
	}

	if workspaceHandle != "" && connHandle != "" {
		id := fmt.Sprintf("%s/%s", workspaceHandle, connHandle)
		d.SetId(id)
		d.Set("workspace_handle", workspaceHandle)
		d.Set("connection_handle", connHandle)
	}

	return nil
}

func resourceSteampipeCloudWorkspaceConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	var err error
	var r *_nethttp.Response
	var userHandle string
	client := meta.(*SteampipeClient)

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return fmt.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}
	workspaceHandle := idParts[0]
	connHandle := idParts[1]

	log.Printf("\n[DEBUG] Deleting Workspace Connection association: %s", fmt.Sprintf("%s/%s", workspaceHandle, connHandle))

	if client.Config != nil && client.Config.Org != "" {
		_, _, err = client.APIClient.OrgWorkspaceConnectionAssociations.Delete(context.Background(), client.Config.Org, workspaceHandle, connHandle).Execute()
	} else {
		userHandle, r, err = getUserHandler(client)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudWorkspaceConnectionDelete.\ngetHandler Error:	\nstatus_code: %d\n	body: %v", r.StatusCode, r.Body)
		}
		_, _, err = client.APIClient.UserWorkspaceConnectionAssociations.Delete(context.Background(), userHandle, workspaceHandle, connHandle).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error deleting workspace connection association:	\n status_code: %d\n	body: %v", r.StatusCode, r.Body)
	}
	d.SetId("")

	return nil
}