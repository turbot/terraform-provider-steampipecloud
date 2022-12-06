package steampipecloud

import (
	"context"
	"encoding/json"
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

func resourceWorkspaceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceSnapshotCreate,
		ReadContext:   resourceWorkspaceSnapshotRead,
		UpdateContext: resourceWorkspaceSnapshotUpdate,
		DeleteContext: resourceWorkspaceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_snapshot_id": {
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
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dashboard_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dashboard_title": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"schema_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"inputs": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": {
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
			"expires_at": {
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
			"data": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
		},
	}
}

func resourceWorkspaceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceSnapshot
	var data steampipe.WorkspaceSnapshotData

	workspaceHandle := d.Get("workspace_handle").(string)
	err = json.Unmarshal([]byte(d.Get("data").(string)), &data)
	if err != nil {
		return diag.Errorf("error parsing data for workspace snapshot : %v", d.Get("data").(string))
	}
	tags, err := JSONStringToInterface(d.Get("tags").(string))
	if err != nil {
		return diag.Errorf("error parsing tags for workspace snapshot : %v", d.Get("tags").(string))
	}
	visibility := d.Get("visibility").(string)
	log.Printf("\n[DEBUG] Snapshot Data: %v", data)

	// Create request
	req := steampipe.CreateWorkspaceSnapshotRequest{Data: data, Tags: tags, Visibility: &visibility}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceSnapshotCreate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceSnapshots.Create(ctx, userHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceSnapshots.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace snapshot: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Snapshot: %s created for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_snapshot_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("state", resp.State)
	d.Set("visibility", resp.Visibility)
	d.Set("dashboard_name", resp.DashboardName)
	d.Set("dashboard_title", resp.DashboardTitle)
	d.Set("schema_version", resp.SchemaVersion)
	d.Set("inputs", resp.Inputs)
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("expires_at", resp.ExpiresAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If snapshot is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/SnapshotID" otherwise "WorkspaceHandle/SnapshotID"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	}

	return diags
}

func resourceWorkspaceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, snapshotId string
	var isUser = false

	// If snapshot is created for a workspace within an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/SnapshotID" otherwise "WorkspaceHandle/SnapshotID"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<snapshot-id>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		snapshotId = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		snapshotId = idParts[1]
	}

	var resp steampipe.WorkspaceSnapshot
	var err error
	var r *http.Response

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceSnapshotRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceSnapshots.Get(ctx, userHandle, workspaceHandle, snapshotId).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceSnapshots.Get(ctx, orgHandle, workspaceHandle, snapshotId).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace snapshot: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Snapshot: %s received for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_snapshot_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("state", resp.State)
	d.Set("visibility", resp.Visibility)
	d.Set("dashboard_name", resp.DashboardName)
	d.Set("dashboard_title", resp.DashboardTitle)
	d.Set("schema_version", resp.SchemaVersion)
	d.Set("inputs", resp.Inputs)
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("expires_at", resp.ExpiresAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If snapshot is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/SnapshotID" otherwise "WorkspaceHandle/SnapshotID"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	}

	return diags
}

func resourceWorkspaceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceSnapshot

	workspaceHandle := d.Get("workspace_handle").(string)
	snapshotId := d.Get("workspace_snapshot_id").(string)
	tags, err := JSONStringToInterface(d.Get("tags").(string))
	if err != nil {
		return diag.Errorf("error parsing tags for workspace snapshot : %v", d.Get("tags").(string))
	}
	visibility := d.Get("visibility").(string)

	// Create request
	req := steampipe.UpdateWorkspaceSnapshotRequest{Tags: tags, Visibility: &visibility}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceSnapshotUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceSnapshots.Update(ctx, userHandle, workspaceHandle, snapshotId).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceSnapshots.Update(ctx, orgHandle, workspaceHandle, snapshotId).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating workspace snapshot: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Snapshot: %s updated for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_snapshot_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("state", resp.State)
	d.Set("visibility", resp.Visibility)
	d.Set("dashboard_name", resp.DashboardName)
	d.Set("dashboard_title", resp.DashboardTitle)
	d.Set("schema_version", resp.SchemaVersion)
	d.Set("inputs", resp.Inputs)
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("expires_at", resp.ExpiresAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If snapshot is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/SnapshotID" otherwise "WorkspaceHandle/SnapshotID"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	}

	return diags
}

func resourceWorkspaceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, snapshotId string
	var isUser = false

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<snapshot-id>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		snapshotId = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		snapshotId = idParts[1]
	}

	log.Printf("\n[DEBUG] Deleting snapshot: %s for workspace: %s", snapshotId, workspaceHandle)

	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceSnapshotDelete.getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceSnapshots.Delete(ctx, actorHandle, workspaceHandle, snapshotId).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceSnapshots.Delete(ctx, orgHandle, workspaceHandle, snapshotId).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting snapshot: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
