package steampipe

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeUserConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeConnectionCreate,
		Read:   resourceSteampipeConnectionRead,
		Delete: resourceSteampipeConnectionDelete,
		// Exists: resourceExistsItem,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceSteampipeConnectionImport,
		// },
		Schema: map[string]*schema.Schema{
			// aka of the parent resource
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
				// ForceNew: true,
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
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
				// Elem: &schema.Schema{
				// 	Type: schema.TypeString,
				// },
			},
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
	req := openapiclient.TypesCreateConnectionRequest{
		Handle: "aad",
		Plugin: "aws",
		Config: &config,
	}

	resp, r, err := client.UserConnectionsApi.CreateUserConnection(context.Background(), "lalit").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	err = d.Set("id", resp.Id)
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
	err = d.Set("config", resp.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
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
	if value, ok := d.GetOk("conn_handle"); ok {
		conn_handle = value.(string)
	}

	if conn_handle == "" {
		// fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet`: %v\n", err)
		return fmt.Errorf("connection handle not present")
	}

	resp, r, err := client.UserConnectionsApi.GetUserConnection(context.Background(), user_handle, conn_handle).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling resourceSteampipeConnectionRead: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	// assign results back into ResourceData
	err = d.Set("id", resp.Id)
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
	err = d.Set("config", resp.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "After SET`: %v\n", err)
		return err
	}
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
	if value, ok := d.GetOk("conn_handle"); ok {
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

// 	{
//   "plugin": "aws",
//   "config": {
//     "regions": [
//       "us-east-1"
//     ],
//     "access_key": "AKIAQGDRKHTKFBLNOL5N",
//     "secret_key": "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"
//   }
// }

// {
//   "id": "c_c6ivcn9e4mvahd3kqbd0",
//   "handle": "aac",
//   "identity_id": "u_c6flu7pe4mvf26h42ibg",
//   "type": "connection",
//   "plugin": "aws",
//   "config": {
//     "access_key": "AKIAQGDRKHTKFBLNOL5N",
//     "regions": [
//       "us-east-1"
//     ]
//   },
//   "version_id": 1,
//   "created_at": "2021-11-30T10:01:01Z",
//   "updated_at": null
// }

// data := *resp.Items
// {
//   "id": "c_c6ivcn9e4mvahd3kqbd0",
//   "handle": "aac",
//   "identity_id": "u_c6flu7pe4mvf26h42ibg",
//   "type": "connection",
//   "plugin": "aws",
//   "config": {
//     "access_key": "AKIAQGDRKHTKFBLNOL5N",
//     "regions": [
//       "us-east-1"
//     ]
//   },
//   "version_id": 1,
//   "created_at": "2021-11-30T10:01:01Z",
//   "updated_at": null
// }
