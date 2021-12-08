package steampipecloud

import (
	"context"
	"encoding/json"
	"fmt"
	_nethttp "net/http"

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

	switch plugin {
	case "aws":
		var awsConfig AwsConnectionConfigWithSecrets
		if value, ok := d.GetOk("regions"); ok {
			var regions []string
			for _, item := range value.([]interface{}) {
				regions = append(regions, item.(string))
			}
			awsConfig.Regions = regions
		}
		if value, ok := d.GetOk("secret_key"); ok {
			awsConfig.SecretKey = value.(string)
		}
		if value, ok := d.GetOk("access_key"); ok {
			awsConfig.AccessKey = value.(string)
		}
		if value, ok := d.GetOk("session_token"); ok {
			awsConfig.SessionToken = value.(string)
		}
		data, _ := json.Marshal(awsConfig)
		json.Unmarshal(data, &config)
	case "gcp":
		var gcpConfig GcpConnectionConfigWithSecrets
		if value, ok := d.GetOk("project"); ok {
			gcpConfig.Project = value.(string)
		}
		if value, ok := d.GetOk("credentials"); ok {
			gcpConfig.Credentials = value.(string)
		}
		data, _ := json.Marshal(gcpConfig)
		json.Unmarshal(data, &config)
	}

	req := openapiclient.TypesCreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
	}

	if config != nil {
		req.SetConfig(config)
	}

	var err error
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
	switch *resp.Plugin {
	case "aws":
		if resp.Config != nil {
			for k, v := range *resp.Config {
				if k == "regions" {
					d.Set(k, v.([]string))
				} else {
					d.Set(k, v.(string))
				}
			}
		}
	case "gcp":
		if resp.Config != nil {
			for k, v := range *resp.Config {
				d.Set(k, v.(string))
			}
		}
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	// if resp.Config != nil {
	// 	configMap := map[string]string{}
	// 	for k, v := range *resp.Config {
	// 		switch item := v.(type) {
	// 		case string:
	// 			configMap[k] = item
	// 		case []string
	// 		}
	// 	}
	// 	d.Set("config", helpers.FormatJson(data.(string)))
	// }
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
	switch *resp.Plugin {
	case "aws":
		if resp.Config != nil {
			for k, v := range *resp.Config {
				if k == "regions" {
					d.Set(k, v.([]interface{}))
				} else {
					d.Set(k, v.(string))
				}
			}
		}
	case "gcp":
		if resp.Config != nil {
			for k, v := range *resp.Config {
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

	plugin := d.Get("plugin")
	if newHandle.(string) == "" {
		return fmt.Errorf("handle must be configured")
	}

	var err error
	var config map[string]interface{}
	var resp openapiclient.TypesConnection

	req := openapiclient.TypesUpdateConnectionRequest{
		Handle: types.String(newHandle.(string)),
	}

	switch plugin.(string) {
	case "aws":
		var awsConfig AwsConnectionConfigWithSecrets
		if value, ok := d.GetOkExists("regions"); ok {
			var regions []string
			for _, item := range value.([]interface{}) {
				regions = append(regions, item.(string))
			}
			awsConfig.Regions = regions
		}
		if value, ok := d.GetOkExists("secretKey"); ok {
			awsConfig.SecretKey = value.(string)
		}
		if value, ok := d.GetOkExists("access_key"); ok {
			awsConfig.AccessKey = value.(string)
		}
		if value, ok := d.GetOkExists("session_token"); ok {
			awsConfig.SessionToken = value.(string)
		}
		data, _ := json.Marshal(awsConfig)
		json.Unmarshal(data, &config)
	case "gcp":
		var gcpConfig GcpConnectionConfigWithSecrets
		if value, ok := d.GetOk("project"); ok {
			gcpConfig.Project = value.(string)
		}
		if value, ok := d.GetOk("credentials"); ok {
			gcpConfig.Credentials = value.(string)
		}
		data, _ := json.Marshal(gcpConfig)
		json.Unmarshal(data, &config)
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
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	switch *resp.Plugin {
	case "aws":
		if resp.Config != nil {
			for k, v := range *resp.Config {
				if k == "regions" {
					d.Set(k, v.([]interface{}))
				} else {
					d.Set(k, v.(string))
				}
			}
		}
	case "gcp":
		if resp.Config != nil {
			for k, v := range *resp.Config {
				d.Set(k, v.(string))
			}
		}
	}
	d.Set("plugin", *resp.Plugin)
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

// data is a json string
// apply standard formatting to old and new data then compare
func suppressIfDataMatches(k, old, new string, d *schema.ResourceData) bool {
	if old == "" || new == "" {
		return false
	}

	oldFormatted := helpers.FormatJson(old)
	newFormatted := helpers.FormatJson(new)
	return oldFormatted == newFormatted
}

// - get properties for a given key from config
// - build a map only including the properties fetched from config
// - convert map to string
func getStringValueForKey(d *schema.ResourceData, key string, readResponse map[string]interface{}) (string, error) {
	propertiesOfKey, err := getPropertiesFromConfig(d, key)
	if err != nil {
		return "", err
	}
	metadata := buildResourceMapFromProperties(readResponse, propertiesOfKey)
	metadataString, err := helpers.MapToJsonString(metadata)
	if err != nil {
		return "", fmt.Errorf("error building resource data: %s", err.Error())
	}
	return metadataString, nil
}

func getPropertiesFromConfig(d *schema.ResourceData, key string) (map[string]string, error) {
	var properties map[string]string = nil
	var err error = nil
	if keyValue, ok := d.GetOk(key); ok {
		if properties, err = helpers.PropertyMapFromJson(keyValue.(string)); err != nil {
			return nil, fmt.Errorf("error retrieving properties: %s", err.Error())
		}
	}
	return properties, nil
}

func buildResourceMapFromProperties(input map[string]interface{}, properties map[string]string) map[string]interface{} {
	for key, _ := range input {
		// delete external keys from response data
		if _, ok := properties[key]; !ok {
			delete(input, key)
		}
	}
	return input
}

// - build a map from the data or full_data property (specified by 'key' parameter)
// - add a `nil` value for deleted properties
// - remove any properties disallowed by the updateSchema
func buildUpdatePayloadForData(d *schema.ResourceData, key string) (map[string]interface{}, error) {
	var err error
	dataMap, err := markPropertiesForDeletion(d, key)
	if err != nil {
		return nil, err
	}
	return dataMap, nil
}

func markPropertiesForDeletion(d *schema.ResourceData, key string) (map[string]interface{}, error) {
	var oldContent, newContent map[string]interface{}
	var err error
	// fetch old(state-file) and new(config) content
	if old, new := d.GetChange(key); old != nil {
		if oldContent, err = helpers.JsonStringToMap(old.(string)); err != nil {
			return nil, fmt.Errorf("error build resource mutation input, failed to unmarshal content: \n%s\nerror: %s", old.(string), err.Error())
		}
		if newContent, err = helpers.JsonStringToMap(new.(string)); err != nil {
			return nil, fmt.Errorf("error build resource mutation input, failed to unmarshal content: \n%s\nerror: %s", new.(string), err.Error())
		}
		// extract keys from old content not in new
		excludeContentProperties := helpers.GetOldMapProperties(oldContent, newContent)
		for _, key := range excludeContentProperties {
			// set keys of old content to `nil` in new content
			// any property which doesn't exist in config is set to nil
			// NOTE: for folder we cannot currently delete the description property
			if _, ok := oldContent[key.(string)]; ok {
				newContent[key.(string)] = nil
			}
		}
	}
	return newContent, nil
}

func ConvertArray(s string) (*[]string, bool) {
	var js []string
	err := json.Unmarshal([]byte(s), &js)
	return &js, err == nil
}

type AwsConnectionConfigWithSecrets struct {
	Regions      []string `json:"regions,omitempty" hcl:"regions"`
	AccessKey    string   `json:"access_key,omitempty" mapstructure:"access_key" hcl:"access_key"`
	SecretKey    string   `json:"secret_key,omitempty" mapstructure:"secret_key" hcl:"secret_key"`
	SessionToken string   `json:"session_token,omitempty" hcl:"session_token" hcle:"omitempty"`
}

type GcpConnectionConfigWithSecrets struct {
	Project     string `json:"project,omitempty" mapstructure:"project" hcl:"project"`
	Credentials string `json:"credentials,omitempty" mapstructure:"credentials" hcl:"credentials"`
}
