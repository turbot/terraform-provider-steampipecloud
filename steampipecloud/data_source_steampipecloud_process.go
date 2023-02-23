package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func dataSourceProcess() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProcessRead,
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},
			"workspace": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},
			"process_id": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pipeline_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceProcessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var resp steampipe.SpProcess
	var r *http.Response
	var err error

	// Retrieve the process_id(mandatory) and workspace passed(if any)
	processId := d.Get("process_id").(string)
	workspace := d.Get("workspace").(string)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("dataSourceProcessRead.getUserHandler error  %v", decodeResponse(r))
		}
		log.Printf("\n[DEBUG] Process get context-> identity:'%s'; workspace:'%s'; process:'%s'", actorHandle, workspace, processId)
		// If a workspace is not passed we can assume that it is an identity process
		if workspace == "" {
			resp, r, err = client.APIClient.UserProcesses.Get(ctx, actorHandle, processId).Execute()
		} else {
			resp, r, err = client.APIClient.UserWorkspaceProcesses.Get(ctx, actorHandle, workspace, processId).Execute()
		}
	} else {
		log.Printf("\n[DEBUG] Process get context-> identity:'%s'; workspace:'%s'; process:'%s'", orgHandle, workspace, processId)
		//  If a workspace is not passed we can assume that it is an identity process
		if workspace == "" {
			resp, r, err = client.APIClient.OrgProcesses.Get(ctx, orgHandle, processId).Execute()
		} else {
			resp, r, err = client.APIClient.OrgWorkspaceProcesses.Get(ctx, orgHandle, workspace, processId).Execute()
		}
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("%v", decodeResponse(r)))
	}

	log.Printf("\n[DEBUG] Process Received: %v", resp)

	d.Set("process_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("pipeline_id", resp.PipelineId)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("updated_at", resp.UpdatedAt)
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace", workspace)
	d.SetId(resp.Id)

	return diags
}
