package steampipe

import (
	"context"
	"fmt"
	"os"

	// "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	// "github.com/hashicorp/terraform/helper/structure"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/go-kit/types"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeUserConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeConnectionCreate,
		Read:   resourceSteampipeConnectionRead,
		Delete: resourceSteampipeConnectionDelete,
		Update: resourceSteampipeConnectionUpdate,
		// Exists: resourceExistsItem,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeConnectionImport,
		},
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"plugin": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// "config": {
			// 	Type:         schema.TypeString,
			// 	Optional:     true,
			// 	ValidateFunc: validation.StringIsJSON,
			// 	StateFunc: func(v interface{}) string {
			// 		json, _ := structure.NormalizeJsonString(v)
			// 		return json
			// 	},
			// },
		},
	}
}

func resourceSteampipeConnectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeConnectionRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	config := map[string]interface{}{
		"regions":    []string{"us-east-1"},
		"access_key": "AKIAQGDRKHTKFBLNOL5N",
		"secret_key": "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",
	}

	var plugin string
	var conn_handle string
	if value, ok := d.GetOk("handle"); ok {
		conn_handle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	// return fmt.Errorf("Data \nconn_handle: %s \nplugin: %s ", conn_handle, plugin)

	req := openapiclient.TypesCreateConnectionRequest{
		Handle: conn_handle,
		Plugin: plugin,
		Config: &config,
	}

	resp, r, err := client.UserConnectionsApi.CreateUserConnection(context.Background(), "lalit").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.SetId(resp.Id)
	// err = d.Set("config", resp.Config)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
	// 	return err
	// }

	return nil
}

func resourceSteampipeConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)
	id := d.Id()
	var user_handle = "lalit"

	if id == "" {
		return fmt.Errorf("connection handle not present. conn_handle: %s", id)
	}

	resp, _, err := client.UserConnectionsApi.GetUserConnection(context.Background(), user_handle, id).Execute()
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeConnectionRead. \nGetUserConnection.error %v", err)
	}

	// assign results back into ResourceData
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	// d.Set("config", resp.Config)
	d.Set("plugin", resp.Plugin)
	d.Set("handle", resp.Handle)
	// d.Set("created_at", resp.CreatedAt)
	// d.Set("updated_at", resp.UpdatedAt)
	// d.Set("identity", resp.Identity)

	return nil
}

func resourceSteampipeConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)
	var user_handle = "lalit"
	var conn_handle string
	if value, ok := d.GetOk("handle"); ok {
		conn_handle = value.(string)
	}

	_, _, err := client.UserConnectionsApi.DeleteUserConnection(context.Background(), user_handle, conn_handle).Execute()
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceSteampipeConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)
	oldHandle, newHandle := d.GetChange("handle")

	if newHandle.(string) == "" {
		return fmt.Errorf("handle must be configured")
	}
	// config := map[string]interface{}{
	// 	"regions":    []string{"us-east-1"},
	// 	"access_key": "AKIAQGDRKHTKFBLNOL5N",
	// 	"secret_key": "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",
	// }

	req := openapiclient.TypesUpdateConnectionRequest{
		Handle: types.String(newHandle.(string)),
		// Config: &config,
	}

	// Get user handler
	user_handle := getUserHandler(meta)
	resp, _, err := client.UserConnectionsApi.UpdateUserConnection(context.Background(), user_handle, oldHandle.(string)).Request(req).Execute()
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeConnectionUpdate: %v", err)
	}

	d.Set("handle", resp.Handle)
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("config", resp.Config)
	d.Set("plugin", resp.Plugin)

	return nil
}

func getUserHandler(meta interface{}) string {
	client := meta.(*openapiclient.APIClient)
	resp, _, err := client.UsersApi.GetActor(context.Background()).Execute()
	if err != nil {
		return ""
	}
	return resp.Handle
}
