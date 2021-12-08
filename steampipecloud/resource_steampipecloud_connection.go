package steampipecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	_nethttp "net/http"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/go-kit/types"
	openapiclient "github.com/turbot/steampipecloud-sdk-go"
	"github.com/turbot/terraform-provider-steampipecloud/helpers"
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
			// "identity": {
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// 	Computed: true,
			// },
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// AWS connection config arguments
			"regions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"session_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// GCP connection config arguments
			"project": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"credentials": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// AZURE connection config arguments
			"environment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subscription_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"client_secret": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Connection argument for DO, Github
			"token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bearer_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Airtable
			"database_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tables": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			//  OCI
			"user_ocid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenancy_ocid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// "config": {
			// 	Type:     schema.TypeMap,
			// 	Optional: true,
			// 	// DiffSuppressFunc: suppressIfDataMatches,
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
	var config map[string]interface{}

	if value, ok := d.GetOk("handle"); ok {
		connHandle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	// Get config to create connection
	var connConfig ConnectionConfig
	if value, ok := d.GetOk("regions"); ok {
		var regions []string
		for _, item := range value.([]interface{}) {
			regions = append(regions, item.(string))
		}
		connConfig.Regions = regions
	}
	if value, ok := d.GetOk("secret_key"); ok {
		connConfig.SecretKey = value.(string)
	}
	if value, ok := d.GetOk("access_key"); ok {
		connConfig.AccessKey = value.(string)
	}
	if value, ok := d.GetOk("session_token"); ok {
		connConfig.SessionToken = value.(string)
	}
	if value, ok := d.GetOk("project"); ok {
		connConfig.Project = value.(string)
	}
	if value, ok := d.GetOk("credentials"); ok {
		creds := value.(string)
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, []byte(creds)); err != nil {
			log.Println(err)
		}
		connConfig.Credentials = buffer.String()
	}
	if value, ok := d.GetOk("environment"); ok {
		connConfig.Environment = value.(string)
	}
	if value, ok := d.GetOk("tenant_id"); ok {
		connConfig.TenantID = value.(string)
	}
	if value, ok := d.GetOk("subscription_id"); ok {
		connConfig.SubscriptionID = value.(string)
	}
	if value, ok := d.GetOk("client_id"); ok {
		connConfig.ClientID = value.(string)
	}
	if value, ok := d.GetOk("client_secret"); ok {
		connConfig.ClientSecret = value.(string)
	}
	if value, ok := d.GetOk("token"); ok {
		connConfig.Token = value.(string)
	}
	if value, ok := d.GetOk("bearer_token"); ok {
		connConfig.BearerToken = value.(string)
	}
	if value, ok := d.GetOk("database_id"); ok {
		connConfig.DatabaseID = value.(string)
	}
	if value, ok := d.GetOk("tables"); ok {
		var tables []string
		for _, item := range value.([]interface{}) {
			tables = append(tables, item.(string))
		}
		connConfig.Tables = tables
	}
	if value, ok := d.GetOk("private_key"); ok {
		privateKey := value.(string)
		connConfig.PrivateKey = strings.ReplaceAll(privateKey, "\r\n", "\\n")
	}
	if value, ok := d.GetOk("user_ocid"); ok {
		connConfig.UserOCID = value.(string)
	}
	if value, ok := d.GetOk("fingerprint"); ok {
		connConfig.Fingerprint = value.(string)
	}
	if value, ok := d.GetOk("tenancy_ocid"); ok {
		connConfig.TenancyOCID = value.(string)
	}
	configByteData, err := json.Marshal(connConfig)
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionCreate. Marshalling connection config error  %v", err)
	}
	err = json.Unmarshal(configByteData, &config)
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionCreate. Unmarshalling connection config error  %v", err)
	}

	req := openapiclient.TypesCreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
	}

	if config != nil {
		req.SetConfig(config)
	}

	var resp openapiclient.TypesConnection
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
		return fmt.Errorf("inside resourceSteampipeCloudConnectionCreate. Create connection error \nError %v", err)
	}

	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.SetId(resp.Id)
	// Save the config
	if resp.Config != nil {
		for k, v := range *resp.Config {
			if helpers.SliceContains([]string{"regions", "Regions", "tables"}, k) {
				d.Set(strings.ToLower(k), v.([]interface{}))
			} else {
				d.Set(k, v.(string))
			}
		}
	}

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
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	// d.Set("identity", resp.Identity)
	if resp.Config != nil {

		for k, v := range *resp.Config {
			if helpers.SliceContains([]string{"regions", "Regions", "tables"}, k) {
				d.Set(strings.ToLower(k), v.([]interface{}))
			} else {
				d.Set(k, v.(string))
			}
		}
	}
	d.SetId(resp.Id)
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

	var err error
	var config map[string]interface{}
	var resp openapiclient.TypesConnection

	req := openapiclient.TypesUpdateConnectionRequest{
		Handle: types.String(newHandle.(string)),
	}

	// Get config to create connection
	var connConfig ConnectionConfig
	if value, ok := d.GetOk("regions"); ok {
		var regions []string
		for _, item := range value.([]interface{}) {
			regions = append(regions, item.(string))
		}
		connConfig.Regions = regions
	}
	if value, ok := d.GetOk("secret_key"); ok {
		connConfig.SecretKey = value.(string)
	}
	if value, ok := d.GetOk("access_key"); ok {
		connConfig.AccessKey = value.(string)
	}
	if value, ok := d.GetOk("session_token"); ok {
		connConfig.SessionToken = value.(string)
	}
	if value, ok := d.GetOk("project"); ok {
		connConfig.Project = value.(string)
	}
	if value, ok := d.GetOk("credentials"); ok {
		creds := value.(string)
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, []byte(creds)); err != nil {
			log.Println(err)
		}
		connConfig.Credentials = buffer.String()
	}
	if value, ok := d.GetOk("environment"); ok {
		connConfig.Environment = value.(string)
	}
	if value, ok := d.GetOk("tenant_id"); ok {
		connConfig.TenantID = value.(string)
	}
	if value, ok := d.GetOk("subscription_id"); ok {
		connConfig.SubscriptionID = value.(string)
	}
	if value, ok := d.GetOk("client_id"); ok {
		connConfig.ClientID = value.(string)
	}
	if value, ok := d.GetOk("client_secret"); ok {
		connConfig.ClientSecret = value.(string)
	}
	if value, ok := d.GetOk("token"); ok {
		connConfig.Token = value.(string)
	}
	if value, ok := d.GetOk("bearer_token"); ok {
		connConfig.BearerToken = value.(string)
	}
	if value, ok := d.GetOk("database_id"); ok {
		connConfig.DatabaseID = value.(string)
	}
	if value, ok := d.GetOk("tables"); ok {
		var tables []string
		for _, item := range value.([]interface{}) {
			tables = append(tables, item.(string))
		}
		connConfig.Tables = tables
	}
	if value, ok := d.GetOk("private_key"); ok {
		privateKey := value.(string)
		connConfig.PrivateKey = strings.ReplaceAll(privateKey, "\r\n", "\\n")
	}
	if value, ok := d.GetOk("user_ocid"); ok {
		connConfig.UserOCID = value.(string)
	}
	if value, ok := d.GetOk("fingerprint"); ok {
		connConfig.Fingerprint = value.(string)
	}
	if value, ok := d.GetOk("tenancy_ocid"); ok {
		connConfig.TenancyOCID = value.(string)
	}
	data, err := json.Marshal(connConfig)
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. Marshalling connection config error  %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. Unmarshalling connection config error  %v", err)
	}

	if config != nil {
		req.SetConfig(config)
	}

	if IsUser {
		actorHandle, err = getUserHandler(meta)
		if err != nil {
			return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. getUserHandler error  %v", err)
		}
		resp, _, err = steampipeClient.APIClient.UserConnectionsApi.UpdateUserConnection(context.Background(), actorHandle, oldHandle.(string)).Request(req).Execute()
	} else {
		resp, _, err = steampipeClient.APIClient.OrgConnectionsApi.UpdateOrgConnection(context.Background(), org, oldHandle.(string)).Request(req).Execute()
	}
	if err != nil {
		return fmt.Errorf("inside resourceSteampipeCloudConnectionUpdate. UpdateConnection error %v", err)
	}

	d.Set("handle", resp.Handle)
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("plugin", *resp.Plugin)
	if resp.Config != nil {
		for k, v := range *resp.Config {
			if helpers.SliceContains([]string{"regions", "Regions", "tables"}, k) {
				d.Set(strings.ToLower(k), v.([]interface{}))
			} else {
				d.Set(k, v.(string))
			}
		}
	}

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

func ConvertArray(s string) (*[]string, bool) {
	var js []string
	err := json.Unmarshal([]byte(s), &js)
	return &js, err == nil
}

type ConnectionConfig struct {
	Regions        []string `json:"regions,omitempty"`
	Tables         []string `json:"tables,omitempty"`
	AccessKey      string   `json:"access_key,omitempty"`
	SecretKey      string   `json:"secret_key,omitempty"`
	SessionToken   string   `json:"session_token,omitempty"`
	Project        string   `json:"project,omitempty"`
	Credentials    string   `json:"credentials,omitempty"`
	Environment    string   `json:"environment,omitempty"`
	TenantID       string   `json:"tenant_id,omitempty"`
	SubscriptionID string   `json:"subscription_id,omitempty"`
	ClientID       string   `json:"client_id,omitempty"`
	ClientSecret   string   `json:"client_secret,omitempty"`
	Token          string   `json:"token,omitempty"`
	BearerToken    string   `json:"bearer_token,omitempty"`
	DatabaseID     string   `json:"database_id,omitempty"`
	PrivateKey     string   `json:"private_key,omitempty"`
	UserOCID       string   `json:"user_ocid,omitempty"`
	Fingerprint    string   `json:"fingerprint,omitempty"`
	TenancyOCID    string   `json:"tenancy_ocid,omitempty"`
}
