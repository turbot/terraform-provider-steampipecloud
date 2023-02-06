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

func resourceWorkspacePipeline() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspacePipelineCreate,
		ReadContext:   resourceWorkspacePipelineRead,
		UpdateContext: resourceWorkspacePipelineUpdate,
		DeleteContext: resourceWorkspacePipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_pipeline_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"frequency": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"pipeline": {
				Type:     schema.TypeString,
				Required: true,
			},
			"args": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"tags": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"last_process_id": {
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

func resourceWorkspacePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.Pipeline

	workspaceHandle := d.Get("workspace").(string)
	title := d.Get("title").(string)
	pipeline := d.Get("pipeline").(string)
	var frequency steampipe.PipelineFrequency
	err = json.Unmarshal([]byte(d.Get("frequency").(string)), &frequency)
	if err != nil {
		return diag.Errorf("error parsing frequency for workspace pipeline : %v", d.Get("frequency").(string))
	}
	args, err := JSONStringToInterface(d.Get("args").(string))
	if err != nil {
		return diag.Errorf("error parsing args for workspace pipeline : %v", d.Get("args").(string))
	}
	tags, err := JSONStringToInterface(d.Get("tags").(string))
	if err != nil {
		return diag.Errorf("error parsing tags for workspace pipeline : %v", d.Get("tags").(string))
	}
	log.Printf("\n[DEBUG] Pipeline Frequency: %v", frequency)
	log.Printf("\n[DEBUG] Pipeline Arguments: %v", args)
	log.Printf("\n[DEBUG] Pipeline Tags: %v", tags)

	// Create request
	req := steampipe.CreatePipelineRequest{Title: title, Pipeline: pipeline, Frequency: frequency, Args: args, Tags: tags}

	userHandle := ""
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspacePipelineCreate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspacePipelines.Create(ctx, userHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspacePipelines.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace pipeline: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Pipeline: %s created for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_pipeline_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("title", resp.Title)
	d.Set("frequency", FormatJson(resp.Frequency))
	d.Set("pipeline", resp.Pipeline)
	d.Set("args", FormatJson(resp.Args))
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("last_process_id", resp.LastProcessId)
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

	// If Pipeline is created for a Workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/PipelineID" otherwise "WorkspaceHandle/PipelineID"
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	}

	return diags
}

func resourceWorkspacePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, pipelineId string
	var isUser = false

	// If Pipeline is created for a Workspace within an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/PipelineID" otherwise "WorkspaceHandle/PipelineID"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<pipeline-id>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		pipelineId = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		pipelineId = idParts[1]
	}

	var resp steampipe.Pipeline
	var err error
	var r *http.Response

	userHandle := ""
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspacePipelineRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspacePipelines.Get(ctx, userHandle, workspaceHandle, pipelineId).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspacePipelines.Get(ctx, orgHandle, workspaceHandle, pipelineId).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace pipeline: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] pipeline: %s received for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_pipeline_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("title", resp.Title)
	d.Set("frequency", FormatJson(resp.Frequency))
	d.Set("pipeline", resp.Pipeline)
	d.Set("args", FormatJson(resp.Args))
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("last_process_id", resp.LastProcessId)
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

	// If Pipeline is created for a Workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/PipelineID" otherwise "WorkspaceHandle/PipelineID"
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	}

	return diags
}

func resourceWorkspacePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.Pipeline

	workspaceHandle := d.Get("workspace").(string)
	pipelineId := d.Get("workspace_pipeline_id").(string)
	title := d.Get("title").(string)
	var frequency steampipe.PipelineFrequency
	err = json.Unmarshal([]byte(d.Get("frequency").(string)), &frequency)
	if err != nil {
		return diag.Errorf("error parsing frequency for workspace pipeline : %v", d.Get("frequency").(string))
	}
	args, err := JSONStringToInterface(d.Get("args").(string))
	if err != nil {
		return diag.Errorf("error parsing args for workspace pipeline : %v", d.Get("args").(string))
	}
	tags, err := JSONStringToInterface(d.Get("tags").(string))
	if err != nil {
		return diag.Errorf("error parsing tags for workspace pipeline : %v", d.Get("tags").(string))
	}
	log.Printf("\n[DEBUG] Pipeline Frequency: %v", frequency)
	log.Printf("\n[DEBUG] Pipeline Arguments: %v", args)
	log.Printf("\n[DEBUG] Pipeline Tags: %v", tags)

	// Create request
	req := steampipe.UpdatePipelineRequest{Title: &title, Frequency: &frequency, Args: args, Tags: tags}

	userHandle := ""
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspacePipelineUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspacePipelines.Update(ctx, userHandle, workspaceHandle, pipelineId).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspacePipelines.Update(ctx, orgHandle, workspaceHandle, pipelineId).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating workspace pipeline: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] pipeline: %s updated for Workspace: %s", resp.Id, workspaceHandle)

	// Set property values
	d.Set("workspace_pipeline_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("title", resp.Title)
	d.Set("frequency", FormatJson(resp.Frequency))
	d.Set("pipeline", resp.Pipeline)
	d.Set("args", FormatJson(resp.Args))
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("last_process_id", resp.LastProcessId)
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

	// If Pipeline is created for a Workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/PipelineID" otherwise "WorkspaceHandle/PipelineID"
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	}

	return diags
}

func resourceWorkspacePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, pipelineId string
	var isUser = false

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<pipeline-id>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		pipelineId = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		pipelineId = idParts[1]
	}

	log.Printf("\n[DEBUG] Deleting pipeline: %s for workspace: %s", pipelineId, workspaceHandle)

	var err error
	var r *http.Response

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspacePipelineDelete.getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspacePipelines.Delete(ctx, userHandle, workspaceHandle, pipelineId).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspacePipelines.Delete(ctx, orgHandle, workspaceHandle, pipelineId).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting pipeline: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
