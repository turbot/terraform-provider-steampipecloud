package steampipecloud

import (
	"context"
	"fmt"
	_nethttp "net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/go-kit/types"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeCloudConnection() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSteampipeCloudConnectionRead,
		Create: resourceSteampipeCloudConnectionCreate,
		Update: resourceSteampipeCloudConnectionUpdate,
		Delete: resourceSteampipeCloudConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudConnectionImport,
		},
		Exists: resourceSteampipeCloudConnectionExists,
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
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"plugin": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceSteampipeCloudConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	IsUser := true
	var org string
	steampipeClient := meta.(*SteampipeClient)
	if steampipeClient.Config != nil {
		if steampipeClient.Config.Org != "" {
			org = steampipeClient.Config.Org
			IsUser = false
		}
	}

	var plugin string
	var connHandle string

	if value, ok := d.GetOk("handle"); ok {
		connHandle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	config := map[string]interface{}{
		"regions":    []string{"us-east-1"},
		"access_key": "redacted",
		"secret_key": "redacted",
	}

	req := openapiclient.TypesCreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
		Config: &config,
	}
	var resp openapiclient.TypesConnection
	var err error
	var actorHandle string
	if IsUser {
		actorHandle, err = getUserHandler(meta)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionCreate. getUserHandler Error: \n%v", err)
		}
		resp, _, err = steampipeClient.APIClient.UserConnectionsApi.CreateUserConnection(context.Background(), actorHandle).Request(req).Execute()
	} else {
		resp, _, err = steampipeClient.APIClient.OrgConnectionsApi.CreateOrgConnection(context.Background(), org).Request(req).Execute()
	}

	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionCreate. Crete connection error \nError %v", err)
	}

	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.SetId(resp.Id)

	return nil
}

func resourceSteampipeCloudConnectionRead(d *schema.ResourceData, meta interface{}) error {
	var org string
	var resp openapiclient.TypesConnection
	var err error
	var actorHandle string
	IsUser := true

	steampipeClient := meta.(*SteampipeClient)
	if steampipeClient.Config != nil {
		if steampipeClient.Config.Org != "" {
			org = steampipeClient.Config.Org
			IsUser = false
		}
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. connection handle not present.")
	}

	if IsUser {
		actorHandle, err = getUserHandler(meta)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. getUserHandler Error: \n%v", err)
		}
		resp, _, err = steampipeClient.APIClient.UserConnectionsApi.GetUserConnection(context.Background(), actorHandle, id).Execute()
	} else {
		resp, _, err = steampipeClient.APIClient.OrgConnectionsApi.GetOrgConnection(context.Background(), org, id).Execute()
	}

	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. \nGetConnection.error %v", err)
	}

	// assign results back into ResourceData
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("handle", resp.Handle)
	d.SetId(resp.Id)
	// d.Set("config", resp.Config)
	// d.Set("created_at", resp.CreatedAt)
	// d.Set("updated_at", resp.UpdatedAt)
	// d.Set("identity", resp.Identity)

	return nil
}

func resourceSteampipeCloudConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	IsUser := true
	var org string
	steampipeClient := meta.(*SteampipeClient)
	if steampipeClient.Config != nil {
		if steampipeClient.Config.Org != "" {
			org = steampipeClient.Config.Org
			IsUser = false
		}
	}
	var conn_handle string
	if value, ok := d.GetOk("handle"); ok {
		conn_handle = value.(string)
	}

	if !IsUser {
		_, _, err := steampipeClient.APIClient.OrgConnectionsApi.DeleteOrgConnection(context.Background(), org, conn_handle).Execute()
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionDelete. DeleteOrgConnection Error: \n%v", err)
		}
	} else {
		actorHandle, err := getUserHandler(meta)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionDelete. getUserHandler Error: \n%v", err)
		}
		_, _, err = steampipeClient.APIClient.UserConnectionsApi.DeleteUserConnection(context.Background(), actorHandle, conn_handle).Execute()
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionDelete. DeleteUserConnection Error: \n%v", err)
		}
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceSteampipeCloudConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	IsUser := true
	var org string
	var actorHandle string
	steampipeClient := meta.(*SteampipeClient)
	if steampipeClient.Config != nil {
		if steampipeClient.Config.Org != "" {
			org = steampipeClient.Config.Org
			IsUser = false
		}
	}

	oldHandle, newHandle := d.GetChange("handle")

	if newHandle.(string) == "" {
		return fmt.Errorf("handle must be configured")
	}
	// config := map[string]interface{}{
	// 	"regions":    []string{"us-east-1"},
	// 	"access_key": "redacted",
	// 	"secret_key": "redacted",
	// }

	req := openapiclient.TypesUpdateConnectionRequest{
		Handle: types.String(newHandle.(string)),
		// Config: &config,
	}

	var resp openapiclient.TypesConnection
	var err error

	if IsUser {
		actorHandle, err = getUserHandler(meta)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. getUserHandler error  %v", err)
		}
		resp, _, err = steampipeClient.APIClient.UserConnectionsApi.UpdateUserConnection(context.Background(), actorHandle, oldHandle.(string)).Request(req).Execute()
	} else {
		// return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. \n newHandle %s \n oldHandle: %s", newHandle.(string), oldHandle.(string))
		resp, _, err = steampipeClient.APIClient.OrgConnectionsApi.UpdateOrgConnection(context.Background(), org, oldHandle.(string)).Request(req).Execute()
	}
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. UpdateConnection error %v", err)
	}

	d.Set("handle", resp.Handle)
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("config", resp.Config)
	d.Set("plugin", resp.Plugin)
	d.SetId(resp.Id)

	return nil
}

func resourceSteampipeCloudConnectionExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	var org string
	var r *_nethttp.Response
	var err error
	var actorHandle string
	IsUser := true

	steampipeClient := meta.(*SteampipeClient)
	if steampipeClient.Config != nil {
		if steampipeClient.Config.Org != "" {
			org = steampipeClient.Config.Org
			IsUser = false
		}
	}

	id := d.Id()
	if id == "" {
		return false, fmt.Errorf("inside resourceSteampipeCloudConnectionExists. connection handle not present.")
	}

	if IsUser {
		actorHandle, err = getUserHandler(meta)
		if err != nil {
			return false, fmt.Errorf("inside resourceSteampipeCloudConnectionExists. getUserHandler Error: \n%v", err)
		}
		_, r, err = steampipeClient.APIClient.UserConnectionsApi.GetUserConnection(context.Background(), actorHandle, id).Execute()
	} else {
		_, r, err = steampipeClient.APIClient.OrgConnectionsApi.GetOrgConnection(context.Background(), org, id).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("inside resourceSteampipeCloudConnectionExists. \nGetConnection.error %v", err)
	}
	return true, nil
}

func resourceSteampipeCloudConnectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudConnectionRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func getUserHandler(meta interface{}) (string, error) {
	steampipeClient := meta.(*SteampipeClient)
	resp, _, err := steampipeClient.APIClient.UsersApi.GetActor(context.Background()).Execute()
	if err != nil {
		return "", err
	}
	return resp.Handle, nil
}