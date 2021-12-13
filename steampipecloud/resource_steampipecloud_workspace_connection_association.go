package steampipecloud

import (
	"context"
	"fmt"
	"log"
	_nethttp "net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeCloudWorkspaceConnectionAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeCloudWorkspaceConnectionAssociationCreate,
		Read:   resourceSteampipeCloudWorkspaceConnectionAssociationRead,
		Delete: resourceSteampipeCloudWorkspaceConnectionAssociationDelete,
		Exists: resourceSteampipeCloudWorkspaceConnectionAssociationExists,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudWorkspaceConnectionAssociationImport,
		},
		Schema: map[string]*schema.Schema{
			"connection_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{1,37}[a-z0-9]$`), "must satisfy regular expression pattern: ^[a-z][a-z0-9_]{1,37}[a-z0-9]$"),
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "must satisfy regular expression pattern: ^[a-z0-9]{1,23}$"),
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

func resourceSteampipeCloudWorkspaceConnectionAssociationExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*SteampipeClient)

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return false, fmt.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}
	workspaceHandle := idParts[0]
	connHandle := idParts[1]

	var err error
	var r *_nethttp.Response

	if client.Config != nil && client.Config.Org != "" {
		_, r, err = client.APIClient.OrgWorkspaceConnectionAssociationsApi.GetOrgWorkspaceConnectionAssociation(context.Background(), client.Config.Org, workspaceHandle, connHandle).Execute()
	} else {
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return false, fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		_, r, err = client.APIClient.UserWorkspaceConnectionAssociationsApi.GetUserWorkspaceConnectionAssociation(context.Background(), userHandler, workspaceHandle, connHandle).Execute()
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

func resourceSteampipeCloudWorkspaceConnectionAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudWorkspaceConnectionAssociationRead(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeCloudWorkspaceConnectionAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)
	workspaceHandle := d.Get("workspace_handle").(string)
	connHandle := d.Get("connection_handle").(string)

	// Create request
	req := openapiclient.TypesCreateWorkspaceConnRequest{
		ConnectionHandle: connHandle,
	}

	var resp openapiclient.TypesWorkspaceConn
	var err error

	// Check for organization
	// If 'org' is set in the provider config, workspace will create in that organization
	// else, the user identity is used.
	if client.Config != nil && client.Config.Org != "" {
		resp, _, err = client.APIClient.OrgWorkspaceConnectionAssociationsApi.CreateOrgWorkspaceConnectionAssociation(context.Background(), client.Config.Org, workspaceHandle).Request(req).Execute()
	} else {
		// Get current actor information
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		resp, _, err = client.APIClient.UserWorkspaceConnectionAssociationsApi.CreateUserWorkspaceConnectionAssociation(context.Background(), userHandler, workspaceHandle).Request(req).Execute()
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

func resourceSteampipeCloudWorkspaceConnectionAssociationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return fmt.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}

	workspaceHandle := idParts[0]
	connHandle := idParts[1]

	var resp openapiclient.TypesWorkspaceConn
	var err error
	var r *_nethttp.Response

	if client.Config != nil && client.Config.Org != "" {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionAssociationsApi.GetOrgWorkspaceConnectionAssociation(context.Background(), client.Config.Org, workspaceHandle, connHandle).Execute()
	} else {
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionAssociationsApi.GetUserWorkspaceConnectionAssociation(context.Background(), userHandler, workspaceHandle, connHandle).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Association (%s) not found", resp.Id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading %s: %s", d.Id(), err)
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

func resourceSteampipeCloudWorkspaceConnectionAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return fmt.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}
	workspaceHandle := idParts[0]
	connHandle := idParts[1]

	log.Printf("\n[DEBUG] Deleting Workspace Connection association: %s", fmt.Sprintf("%s/%s", workspaceHandle, connHandle))

	var err error

	if client.Config != nil && client.Config.Org != "" {
		_, _, err = client.APIClient.OrgWorkspaceConnectionAssociationsApi.DeleteOrgWorkspaceConnectionAssociation(context.Background(), client.Config.Org, workspaceHandle, connHandle).Execute()
	} else {
		userHandler, userErr := getUserHandler(meta)
		if userErr != nil {
			return fmt.Errorf("failed to get user handle. Verify the token has been set correctly, error %s", userErr)
		}
		_, _, err = client.APIClient.UserWorkspaceConnectionAssociationsApi.DeleteUserWorkspaceConnectionAssociation(context.Background(), userHandler, workspaceHandle, connHandle).Execute()
	}

	// Error check
	if err != nil {
		return fmt.Errorf("error deleting workspace connection association: %s", err)
	}
	d.SetId("")

	return nil
}
