package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceWorkspaceModVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceModVariableCreateSetting,
		ReadContext:   resourceWorkspaceModVariableRead,
		UpdateContext: resourceWorkspaceModVariableUpdateSetting,
		DeleteContext: resourceWorkspaceModVariableDeleteSetting,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_mod_variable_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"default_value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"type": {
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
			"mod_alias": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9_]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
		},
	}
}

func resourceWorkspaceModVariableCreateSetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceModVariable

	workspaceHandle := d.Get("workspace_handle").(string)
	modAlias := d.Get("mod_alias").(string)
	variableName := d.Get("name").(string)
	settingRaw := d.Get("setting_value")
	setting, err := JSONStringToInterface(settingRaw.(string))
	if err != nil {
		return diag.Errorf("error parsing setting for workspace mod variable : %v", setting)
	}

	// Create request
	req := steampipe.CreateWorkspaceModVariableSettingRequest{Name: variableName, Setting: setting}
	log.Printf("\n[DEBUG] Request Setting : %v \n", req.Setting)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceModVariableCreateSetting. getUserHandle error  %v", decodeResponse(r))
		}
		// After Mod installation - it might so happen that the mod variable has yet to be created, which is why we will retry the setting creation
		// logic until the mod is installed and the variables created in the workspace
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			var err error
			resp, r, err = client.APIClient.UserWorkspaceModVariables.CreateSetting(ctx, userHandle, workspaceHandle, modAlias).Request(req).Execute()
			if err != nil {
				return resource.RetryableError(err)
			}
			return nil
		})
	} else {
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			resp, r, err = client.APIClient.OrgWorkspaceModVariables.CreateSetting(ctx, orgHandle, workspaceHandle, modAlias).Request(req).Execute()
			if err != nil {
				return resource.RetryableError(err)
			}
			return nil
		})
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating setting for workspace mod variable : %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Setting created for variable: %s of mod: %s in workspace: %s", variableName, modAlias, workspaceHandle)

	log.Printf("\n[DEBUG] Decoded Response: %v", decodeResponse(r))

	// Set property values
	d.Set("workspace_mod_variable_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("workspace_handle", workspaceHandle)
	d.Set("mod_alias", modAlias)
	d.Set("organization", orgHandle)
	d.Set("default_value", FormatJson(resp.ValueDefault))
	d.Set("setting_value", FormatJson(resp.ValueSetting))
	d.Set("value", FormatJson(resp.Value))

	// If the mod variable belongs to a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias/VariableName" otherwise "WorkspaceHandle/ModAlias/VariableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, modAlias, variableName))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, modAlias, variableName))
	}

	return diags
}

func resourceWorkspaceModVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias, variableName string
	var isUser = false

	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(d.Id(), ":") {
		separator = ":"
	}
	// If mod is installed for a workspace within an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias" otherwise "WorkspaceHandle/ModAlias"
	idParts := strings.Split(d.Id(), separator)
	if len(idParts) < 3 && len(idParts) > 4 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<mod-alias>/<name>", d.Id())
	}

	if len(idParts) == 4 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		modAlias = idParts[2]
		variableName = idParts[3]
	} else if len(idParts) == 3 {
		isUser = true
		workspaceHandle = idParts[0]
		modAlias = idParts[1]
		variableName = idParts[2]
	}

	var resp steampipe.WorkspaceModVariable
	var err error
	var r *http.Response

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceModVariableRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceModVariables.GetSetting(ctx, userHandle, workspaceHandle, modAlias, variableName).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceModVariables.GetSetting(ctx, orgHandle, workspaceHandle, modAlias, variableName).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace mod variable : %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Varible: %s received for Mod: %s in Workspace: %s", variableName, modAlias, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_variable_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("workspace_handle", workspaceHandle)
	d.Set("mod_alias", modAlias)
	d.Set("organization", orgHandle)
	d.Set("default_value", FormatJson(resp.ValueDefault))
	d.Set("setting_value", FormatJson(resp.ValueSetting))
	d.Set("value", FormatJson(resp.Value))

	// If the mod variable belongs to a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias/VariableName" otherwise "WorkspaceHandle/ModAlias/VariableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, modAlias, variableName))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, modAlias, variableName))
	}

	return diags
}

func resourceWorkspaceModVariableUpdateSetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.WorkspaceModVariable

	workspaceHandle := d.Get("workspace_handle").(string)
	modAlias := d.Get("mod_alias").(string)
	variableName := d.Get("name").(string)
	settingRaw := d.Get("setting_value")
	setting, err := JSONStringToInterface(settingRaw.(string))
	if err != nil {
		return diag.Errorf("error parsing setting for workspace mod variable : %v", setting)
	}

	// Create request
	req := steampipe.UpdateWorkspaceModVariableSettingRequest{Setting: setting}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceModVariableUpdateSetting. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceModVariables.UpdateSetting(ctx, userHandle, workspaceHandle, modAlias, variableName).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceModVariables.UpdateSetting(ctx, orgHandle, workspaceHandle, modAlias, variableName).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating setting for workspace mod variable : %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Setting updated for variable: %s of mod: %s in workspace: %s", variableName, modAlias, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_variable_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("workspace_handle", workspaceHandle)
	d.Set("mod_alias", modAlias)
	d.Set("organization", orgHandle)
	d.Set("default_value", FormatJson(resp.ValueDefault))
	d.Set("setting_value", FormatJson(resp.ValueSetting))
	d.Set("value", FormatJson(resp.Value))

	// If the mod variable belongs to a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias/VariableName" otherwise "WorkspaceHandle/ModAlias/VariableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, modAlias, variableName))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, modAlias, variableName))
	}

	return diags
}

func resourceWorkspaceModVariableDeleteSetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias, variableName string
	var isUser = false

	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(d.Id(), ":") {
		separator = ":"
	}
	idParts := strings.Split(d.Id(), separator)
	if len(idParts) < 3 && len(idParts) > 4 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<mod-alias>/<variable-name>", d.Id())
	}

	if len(idParts) == 4 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		modAlias = idParts[2]
		variableName = idParts[3]
	} else if len(idParts) == 3 {
		isUser = true
		workspaceHandle = idParts[0]
		modAlias = idParts[1]
		variableName = idParts[2]
	}

	log.Printf("\n[DEBUG] Setting deleted for variable: %s of mod: %s in workspace: %s", variableName, modAlias, workspaceHandle)

	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceModVariables.DeleteSetting(ctx, actorHandle, workspaceHandle, modAlias, variableName).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceModVariables.DeleteSetting(ctx, orgHandle, workspaceHandle, modAlias, variableName).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting mod variable setting: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
