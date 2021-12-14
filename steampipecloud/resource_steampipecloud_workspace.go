package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_nethttp "net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *_nethttp.Response
	var resp steampipe.TypesWorkspace
	handle := d.Get("handle")

	// Create request
	req := steampipe.TypesCreateWorkspaceRequest{Handle: handle.(string)}

	isUser, orgHandle := isUserConnection(client)
	if isUser {
		var userHandler string
		userHandler, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, _, err = client.APIClient.UserWorkspaces.Create(ctx, userHandler).Request(req).Execute()
	} else {
		resp, _, err = client.APIClient.OrgWorkspaces.Create(ctx, orgHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace: %v", decodeResponse(r))
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

	return diags
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	workspaceHandle := d.Id()
	var resp steampipe.TypesWorkspace
	var err error
	var r *_nethttp.Response

	isUser, orgHandle := isUserConnection(client)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaces.Get(ctx, actorHandle, workspaceHandle).Execute()

	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Get(ctx, orgHandle, workspaceHandle).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Workspace (%s) not found", workspaceHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("error reading %s: %v ", workspaceHandle, decodeResponse(r))
	}

	d.Set("handle", workspaceHandle)
	d.Set("workspace_id", resp.Id)
	d.Set("workspace_state", resp.WorkspaceState)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("identity_id", resp.IdentityId)
	d.Set("version_id", resp.VersionId)

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

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
	var r *http.Response

	isUser, orgHandle := isUserConnection(client)
	if isUser {
		userHandler, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionUpdate. getUserHandler error:	%v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaces.Update(ctx, userHandler, oldHandle.(string)).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Update(ctx, orgHandle, oldHandle.(string)).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating workspace: %v", decodeResponse(r))
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

	return diags
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	client := meta.(*SteampipeClient)
	handle := d.Id()
	log.Printf("\n[DEBUG] Deleting Workspace: %s", handle)

	var err error
	var r *http.Response

	isUser, orgHandle := isUserConnection(client)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaces.Delete(ctx, actorHandle, handle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaces.Delete(ctx, orgHandle, handle).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting workspace: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
