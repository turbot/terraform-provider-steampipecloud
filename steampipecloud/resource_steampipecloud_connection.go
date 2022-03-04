package steampipecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"plugin": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
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
				Optional: true,
				Computed: true,
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: connectionJSONStringsEqual,
			},
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var plugin, connHandle, configString string
	var config map[string]interface{}
	var err error

	if value, ok := d.GetOk("handle"); ok {
		connHandle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	if value, ok := d.GetOk("config"); ok {
		configString, config = formatConnectionJSONString(plugin, value.(string))
	}

	req := steampipe.CreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
	}

	if config != nil {
		req.SetConfig(config)
	}

	client := meta.(*SteampipeClient)
	var resp steampipe.Connection
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserConnections.Create(ctx, actorHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgConnections.Create(ctx, orgHandle).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceConnectionCreate. Create connection api error  %v", decodeResponse(r))
	}

	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("organization", orgHandle)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if config != nil {
		d.Set("config", configString)
	}

	// If connection is created inside an Organization the id will be of the
	// format "OrganizationHandle:ConnectionHandle" otherwise "ConnectionHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s:%s", orgHandle, resp.Handle))
	} else {
		d.SetId(resp.Handle)
	}

	return diags
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var connectionHandle, orgHandle string
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.Connection
	var isUser = false
	id := d.Id()

	// If connection exists inside an Organization the id will be of the
	// format "OrganizationHandle:ConnectionHandle" otherwise "ConnectionHandle"
	ids := strings.Split(id, ":")
	if len(ids) == 2 {
		orgHandle = ids[0]
		connectionHandle = ids[1]
	} else if len(ids) == 1 {
		isUser = true
		connectionHandle = ids[0]
	}

	if connectionHandle == "" {
		return diag.Errorf("resourceConnectionRead. Connection handle not present.")
	}

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserConnections.Get(context.Background(), actorHandle, connectionHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgConnections.Get(context.Background(), orgHandle, connectionHandle).Execute()
	}
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Connection (%s) not found", connectionHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceConnectionRead. Get connection error: %v", decodeResponse(r))
	}

	// assign results back into ResourceData
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("organization", orgHandle)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("handle", resp.Handle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	var plugin, configString string
	var r *http.Response
	var resp steampipe.Connection
	var err error
	var config map[string]interface{}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	oldConnectionHandle, newConnectionHandle := d.GetChange("handle")
	if newConnectionHandle.(string) == "" {
		return diag.Errorf("handle must be configured")
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	if value, ok := d.GetOk("config"); ok {
		configString, config = formatConnectionJSONString(plugin, value.(string))
	}

	req := steampipe.UpdateConnectionRequest{Handle: types.String(newConnectionHandle.(string))}
	if config != nil {
		req.SetConfig(config)
	}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionUpdate. getUserHandler error:	%v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserConnections.Update(context.Background(), actorHandle, oldConnectionHandle.(string)).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgConnections.Update(context.Background(), orgHandle, oldConnectionHandle.(string)).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceConnectionUpdate. Update connection error: %v", decodeResponse(r))
	}

	d.Set("handle", resp.Handle)
	d.Set("organization", orgHandle)
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("plugin", *resp.Plugin)

	// If connection exists inside an Organization the id will be of the
	// format "OrganizationHandle:ConnectionHandle" otherwise "ConnectionHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s:%s", orgHandle, resp.Handle))
	} else {
		d.SetId(resp.Handle)
	}
	if config != nil {
		d.Set("config", configString)
	}
	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var connectionHandle string
	if value, ok := d.GetOk("handle"); ok {
		connectionHandle = value.(string)
	}

	var err error
	var r *http.Response
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserConnections.Delete(ctx, actorHandle, connectionHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgConnections.Delete(ctx, orgHandle, connectionHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("resourceConnectionDelete. Delete connection error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}

// config is a json string
// apply standard formatting to old and new data then compare
func connectionJSONStringsEqual(k, old, new string, d *schema.ResourceData) bool {
	if old == "" || new == "" {
		return false
	}
	var plugin string
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}
	oldFormatted, _ := formatConnectionJSONString(plugin, old)
	newFormatted, _ := formatConnectionJSONString(plugin, new)
	return oldFormatted == newFormatted
}

// apply standard formatting to a json string by unmarshalling into a map then marshalling back to JSON
func formatConnectionJSONString(plugin, body string) (string, map[string]interface{}) {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, []byte(body))
	if err != nil {
		return body, nil
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(buffer.Bytes(), &data); err != nil {
		// ignore error and just return original body
		return body, nil
	}

	body, err = mapToJSONString(data)
	if err != nil {
		// ignore error and just return original body
		return body, data
	}
	return body, data
}
