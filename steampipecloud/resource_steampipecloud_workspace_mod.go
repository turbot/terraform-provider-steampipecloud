package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceWorkspaceMod() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceModInstall,
		ReadContext:   resourceWorkspaceModRead,
		UpdateContext: resourceWorkspaceModUpdate,
		DeleteContext: resourceWorkspaceModUninstall,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_mod_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"constraint": {
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
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"installed_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"details": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
		},
	}
}

func resourceWorkspaceModInstall(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceMod

	workspaceHandle := d.Get("workspace_handle").(string)
	path := d.Get("path").(string)
	var constraint string

	if value, ok := d.GetOk("handle"); ok {
		constraint = value.(string)
	}

	// Create request
	req := steampipe.CreateWorkspaceModRequest{Path: path, Constraint: &constraint}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceMods.Install(ctx, userHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceMods.Install(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace mod: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Mod: %s installed for Workspace: %s", *resp.Path, workspaceHandle)
	log.Printf("\n[DEBUG] Mod Alias : %s ", *resp.Alias)

	// Set property values
	d.Set("workspace_mod_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("constraint", resp.Constraint)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("alias", resp.Alias)
	d.Set("installed_version", resp.InstalledVersion)
	d.Set("state", resp.State)
	d.Set("path", resp.Path)
	d.Set("details", resp.Details)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If mod is installed for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle:WorkspaceHandle:ModAlias" otherwise "WorkspaceHandle:ModAlias"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s:%s:%s", orgHandle, workspaceHandle, *resp.Alias))
	} else {
		d.SetId(fmt.Sprintf("%s:%s", workspaceHandle, *resp.Alias))
	}

	return diags
}

func resourceWorkspaceModRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias string
	var isUser = false

	// If mod is installed for a workspace within an Organization the id will be of the
	// format "OrganizationHandle:WorkspaceHandle:ModAlias" otherwise "WorkspaceHandle:ModAlias"
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>:<mod-alias>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		modAlias = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		modAlias = idParts[1]
	}

	var resp steampipe.WorkspaceMod
	var err error
	var r *http.Response

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceModRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceMods.Get(ctx, userHandle, workspaceHandle, modAlias).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceMods.Get(ctx, orgHandle, workspaceHandle, modAlias).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace mod: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Mod: %s received for Workspace: %s", *resp.Path, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("constraint", resp.Constraint)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("alias", resp.Alias)
	d.Set("installed_version", resp.InstalledVersion)
	d.Set("state", resp.State)
	d.Set("path", resp.Path)
	d.Set("details", resp.Details)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If mod is installed for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle:WorkspaceHandle:ModAlias" otherwise "WorkspaceHandle:ModAlias"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s:%s:%s", orgHandle, workspaceHandle, *resp.Alias))
	} else {
		d.SetId(fmt.Sprintf("%s:%s", workspaceHandle, *resp.Alias))
	}

	return diags
}

func resourceWorkspaceModUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceMod

	workspaceHandle := d.Get("workspace_handle").(string)
	modAlias := d.Get("alias").(string)
	constraint := d.Get("constraint").(string)

	// Create request
	req := steampipe.UpdateWorkspaceModRequest{Constraint: constraint}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceMods.Update(ctx, userHandle, workspaceHandle, modAlias).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceMods.Update(ctx, orgHandle, workspaceHandle, modAlias).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace mod: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Mod: %s installed for Workspace: %s", *resp.Path, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("constraint", resp.Constraint)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("alias", resp.Alias)
	d.Set("installed_version", resp.InstalledVersion)
	d.Set("state", resp.State)
	d.Set("path", resp.Path)
	d.Set("details", resp.Details)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If mod is installed for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle:WorkspaceHandle:ModAlias" otherwise "WorkspaceHandle:ModAlias"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s:%s:%s", orgHandle, workspaceHandle, *resp.Alias))
	} else {
		d.SetId(fmt.Sprintf("%s:%s", workspaceHandle, *resp.Alias))
	}

	return diags
}

func resourceWorkspaceModUninstall(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias string
	var isUser = false

	idParts := strings.Split(d.Id(), ":")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>:<mod-alias>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		modAlias = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		modAlias = idParts[1]
	}

	log.Printf("\n[DEBUG] Uninstalling mod: %s for workspace: %s", modAlias, workspaceHandle)

	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceMods.Uninstall(ctx, actorHandle, workspaceHandle, modAlias).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceMods.Uninstall(ctx, orgHandle, workspaceHandle, modAlias).Execute()
	}

	if err != nil {
		return diag.Errorf("error uninstalling mod: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
