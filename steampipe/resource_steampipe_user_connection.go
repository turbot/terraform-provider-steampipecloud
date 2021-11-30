package steampipe

import (
	"context"
	"fmt"
	"os"

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
		// Importer: &schema.ResourceImporter{
		// 	State: resourceSteampipeConnectionImport,
		// },
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
			// 	Type:     schema.TypeMap,
			// 	Optional: true,
			// 	Elem: &schema.Schema{

			// 		Type: schema.TypeString,
			// 	},
			// },
		},
	}
}

// func resourceSteampipeConnectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// 	if err := resourceSteampipeConnectionRead(d, meta); err != nil {
// 		return nil, err
// 	}
// 	return []*schema.ResourceData{d}, nil
// }

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

	err = d.Set("connection_id", resp.Id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	err = d.Set("identity_id", resp.IdentityId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	err = d.Set("type", resp.Type)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	// err = d.Set("config", resp.Config)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
	// 	return err
	// }
	err = d.Set("plugin", resp.Plugin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	d.SetId(resp.Id)

	return nil
}

func resourceSteampipeConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)
	var user_handle = "lalit"
	var conn_handle string
	if value, ok := d.GetOk("handle"); ok {
		conn_handle = value.(string)
	}

	if conn_handle == "" {
		// fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet`: %v\n", err)
		return fmt.Errorf("connection handle not present")
	}
	// type TypesUpdateConnectionRequest struct {
	// 	Config *map[string]interface{} `json:"config,omitempty"`
	// 	Handle *string `json:"handle,omitempty"`
	// }

	resp, r, err := client.UserConnectionsApi.GetUserConnection(context.Background(), user_handle, conn_handle).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling resourceSteampipeConnectionRead: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	// assign results back into ResourceData
	err = d.Set("connection_id", resp.Id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	err = d.Set("identity_id", resp.IdentityId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	err = d.Set("type", resp.Type)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	// err = d.Set("config", resp.Config)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
	// 	return err
	// }
	err = d.Set("plugin", resp.Plugin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	// data := *resp.Items
	// d.Set("handle", resp.Handle)

	return nil
}

func resourceSteampipeConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)
	var user_handle = "lalit"
	var conn_handle string
	if value, ok := d.GetOk("handle"); ok {
		conn_handle = value.(string)
	}

	_, r, err := client.UserConnectionsApi.DeleteUserConnection(context.Background(), user_handle, conn_handle).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
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
		return fmt.Errorf("AM I HERE %v", err)
	}

	err = d.Set("handle", resp.Handle)
	err = d.Set("connection_id", resp.Id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	err = d.Set("identity_id", resp.IdentityId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	err = d.Set("type", resp.Type)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	// err = d.Set("config", resp.Config)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
	// 	return err
	// }
	err = d.Set("plugin", resp.Plugin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
	// d.SetId(resp.Id)

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
