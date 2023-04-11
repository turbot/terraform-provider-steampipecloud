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

func resourceWorkspaceAggregator() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceAggregatorCreate,
		ReadContext:   resourceWorkspaceAggregatorRead,
		UpdateContext: resourceWorkspaceAggregatorUpdate,
		DeleteContext: resourceWorkspaceAggregatorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_aggregator_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"plugin": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connections": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
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
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
		},
	}
}

func resourceWorkspaceAggregatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceAggregator

	workspaceHandle := d.Get("workspace").(string)
	aggregatorHandle := d.Get("handle").(string)
	plugin := d.Get("plugin").(string)
	connections := d.Get("connections").([]string)

	log.Printf("\n[DEBUG] Workspace Handle: %v", workspaceHandle)
	log.Printf("\n[DEBUG] Aggregator Handle: %v", aggregatorHandle)
	log.Printf("\n[DEBUG] Aggregator Plugin: %v", plugin)
	log.Printf("\n[DEBUG] Aggregator Connections: %v", connections)

	// Create request
	req := steampipe.CreateWorkspaceAggregatorRequest{Handle: aggregatorHandle, Plugin: plugin, Connections: connections}

	userHandle := ""
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceAggregatorCreate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceAggregators.Create(ctx, userHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceAggregators.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace aggregator: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Aggregator: %s created for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_aggregator_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("connections", FormatJson(resp.Connections))
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)

	// If an aggregator is created for a workspace inside an organization then the ID will be of the
	// format "OrganizationHandle/WorkspaceHandle/AggregatorHandle" otherwise "WorkspaceHandle/AggregatorHandle".
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Handle))
	}

	return diags
}

func resourceWorkspaceAggregatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, aggregatorHandle string
	var isUser = false

	// If an aggregator is created for a workspace inside an organization then the ID will be of the
	// format "OrganizationHandle/WorkspaceHandle/AggregatorHandle" otherwise "WorkspaceHandle/AggregatorHandle".
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<aggregator-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		aggregatorHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		aggregatorHandle = idParts[1]
	}

	var resp steampipe.WorkspaceAggregator
	var err error
	var r *http.Response

	userHandle := ""
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceAggregatorRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceAggregators.Get(ctx, userHandle, workspaceHandle, aggregatorHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceAggregators.Get(ctx, orgHandle, workspaceHandle, aggregatorHandle).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace aggregator: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Aggregator: %s received for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_aggregator_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("connections", FormatJson(resp.Connections))
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)

	// If an aggregator is created for a workspace inside an organization then the ID will be of the
	// format "OrganizationHandle/WorkspaceHandle/AggregatorHandle" otherwise "WorkspaceHandle/AggregatorHandle".
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Handle))
	}

	return diags
}

func resourceWorkspaceAggregatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceAggregator

	workspaceHandle := d.Get("workspace").(string)
	aggregatorHandle := d.Get("handle").(string)
	connections := d.Get("connections").([]string)

	log.Printf("\n[DEBUG] Workspace Handle: %v", workspaceHandle)
	log.Printf("\n[DEBUG] Aggregator Handle: %v", aggregatorHandle)
	log.Printf("\n[DEBUG] Aggregator Connections: %v", connections)

	// Create request
	req := steampipe.UpdateWorkspaceAggregatorRequest{Handle: &aggregatorHandle, Connections: &connections}

	userHandle := ""
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceAggregatorUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceAggregators.Update(ctx, userHandle, workspaceHandle, aggregatorHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceAggregators.Update(ctx, orgHandle, workspaceHandle, aggregatorHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating workspace aggregator: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Aggregator: %s updated for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_aggregator_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("connections", FormatJson(resp.Connections))
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)

	// If an aggregator is created for a workspace inside an organization then the ID will be of the
	// format "OrganizationHandle/WorkspaceHandle/AggregatorHandle" otherwise "WorkspaceHandle/AggregatorHandle".
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Handle))
	}

	return diags
}

func resourceWorkspaceAggregatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, aggregatorHandle string
	var isUser = false

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<aggregator-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		aggregatorHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		aggregatorHandle = idParts[1]
	}

	log.Printf("\n[DEBUG] Deleting Aggregator: %s for Workspace: %s", aggregatorHandle, workspaceHandle)

	var err error
	var r *http.Response

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceAggregatorDelete.getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceAggregators.Delete(ctx, userHandle, workspaceHandle, aggregatorHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceAggregators.Delete(ctx, orgHandle, workspaceHandle, aggregatorHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting aggregator: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
