package steampipecloud

import (
	"bytes"
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
	"github.com/turbot/go-kit/helpers"
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

			// Specific plugin configs arguments
			// AWS, Alicloud
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// AWS, Alicloud
			"secret_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// AWS
			"session_token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// GCP
			"project": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// GCP
			"credentials": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// AZURE, AzureAD
			"environment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// AZURE, AzureAD
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// AZURE, AzureAD
			"subscription_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// AZURE, AzureAD
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// AZURE, AzureAD
			"client_secret": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// Digital Ocean, GitHub, Airtable, Jira, Linode, Slack
			"token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// Digital Ocean
			"bearer_token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// Airtable
			"database_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			//  OCI
			"user_ocid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			//  OCI
			"fingerprint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			//  OCI
			"tenancy_ocid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			//  OCI
			"private_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// Jira, Bitbucket
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Jira, Bitbucket
			"base_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Bitbucket
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			// Hacker News
			"max_items": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			// hackernews, Zendesk, DataDog, IBM, Cloudflare, Stripe
			"api_key": {
				Type:      schema.TypeInt,
				Sensitive: true,
				Optional:  true,
			},
			// Zendesk, Cloudflare
			"subdomain": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			// Zendesk, Cloudflare
			"email": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			// AWS, OCI, Alicloud, IBM
			"regions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Airtable
			"tables": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var plugin, connHandle string
	var config map[string]interface{}

	if value, ok := d.GetOk("handle"); ok {
		connHandle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	// Get config to create connection
	connConfig, err := CreateConnectionConfiguration(d)
	if err != nil {
		return diag.Errorf("resourceConnectionUpdate. Error while creating connection:  %v", err)
	}

	configByteData, err := json.Marshal(connConfig)
	if err != nil {
		return diag.Errorf("resourceConnectionCreate. Marshalling connection config error  %v", err)
	}
	err = json.Unmarshal(configByteData, &config)
	if err != nil {
		return diag.Errorf("resourceConnectionCreate. Unmarshalling connection config error  %v", err)
	}

	req := steampipe.TypesCreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
	}

	if config != nil {
		req.SetConfig(config)
	}

	client := meta.(*SteampipeClient)
	var resp steampipe.TypesConnection
	var r *http.Response

	isUser, orgHandle := isUserConnection(client)
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
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.SetId(resp.Handle)
	// Save the config
	if resp.Config != nil {
		for k, v := range *resp.Config {
			if v != nil {
				if helpers.StringSliceContains([]string{"regions", "Regions", "tables"}, k) {
					d.Set(strings.ToLower(k), v.([]interface{}))
				} else {
					d.Set(k, v.(string))
				}
			}
		}
	}

	return diags
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp steampipe.TypesConnection

	connectionHandle := d.Id()
	if connectionHandle == "" {
		return diag.Errorf("resourceConnectionRead. Connection handle not present.")
	}

	isUser, orgHandle := isUserConnection(client)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionRead. getUserHandler error  %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserConnections.Get(context.Background(), actorHandle, connectionHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgConnections.Get(context.Background(), orgHandle, connectionHandle).Execute()
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
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("handle", resp.Handle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.SetId(resp.Handle)
	if resp.Config != nil {
		for k, v := range *resp.Config {
			if v != nil {
				if helpers.StringSliceContains([]string{"regions", "Regions", "tables"}, k) {
					d.Set(strings.ToLower(k), v.([]interface{}))
				} else {
					d.Set(k, v.(string))
				}
			}
		}
	}

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	oldConnectionHandle, newConnectionHandle := d.GetChange("handle")
	if newConnectionHandle.(string) == "" {
		return diag.Errorf("handle must be configured")
	}

	var r *http.Response
	var resp steampipe.TypesConnection
	var err error
	var config map[string]interface{}

	// Create connection config to be updated
	connConfig, err := CreateConnectionConfiguration(d)
	if err != nil {
		return diag.Errorf("resourceConnectionUpdate. Error while creating connection:  %v", err)
	}
	data, err := json.Marshal(connConfig)
	if err != nil {
		return diag.Errorf("resourceConnectionUpdate. Marshalling connection config error  %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return diag.Errorf("resourceConnectionUpdate. Unmarshalling connection config error  %v", err)
	}

	req := steampipe.TypesUpdateConnectionRequest{Handle: types.String(newConnectionHandle.(string))}
	if config != nil {
		req.SetConfig(config)
	}

	isUser, orgHandle := isUserConnection(client)
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
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("plugin", *resp.Plugin)
	d.SetId(resp.Handle)
	if resp.Config != nil {
		for k, v := range *resp.Config {
			if v != nil {
				if helpers.StringSliceContains([]string{"regions", "Regions", "tables"}, k) {
					d.Set(strings.ToLower(k), v.([]interface{}))
				} else {
					d.Set(k, v.(string))
				}
			}
		}
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
	isUser, orgHandle := isUserConnection(client)
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

func CreateConnectionConfiguration(d *schema.ResourceData) (ConnectionConfig, error) {
	var connConfig ConnectionConfig

	if value, ok := d.GetOk("access_key"); ok {
		connConfig.AccessKey = value.(string)
	}
	if value, ok := d.GetOk("api_key"); ok {
		connConfig.ApiKey = value.(string)
	}
	if value, ok := d.GetOk("base_url"); ok {
		connConfig.BaseURL = value.(string)
	}
	if value, ok := d.GetOk("bearer_token"); ok {
		connConfig.BearerToken = value.(string)
	}
	if value, ok := d.GetOk("client_id"); ok {
		connConfig.ClientID = value.(string)
	}
	if value, ok := d.GetOk("client_secret"); ok {
		connConfig.ClientSecret = value.(string)
	}
	if value, ok := d.GetOk("credentials"); ok {
		creds := value.(string)
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, []byte(creds)); err != nil {
			log.Println(err)
		}
		connConfig.Credentials = buffer.String()
	}
	if value, ok := d.GetOk("database_id"); ok {
		connConfig.DatabaseID = value.(string)
	}
	if value, ok := d.GetOk("email"); ok {
		connConfig.Email = value.(string)
	}
	if value, ok := d.GetOk("environment"); ok {
		connConfig.Environment = value.(string)
	}
	if value, ok := d.GetOk("fingerprint"); ok {
		connConfig.Fingerprint = value.(string)
	}
	if value, ok := d.GetOk("max_items"); ok {
		connConfig.MaxItems = value.(int)
	}
	if value, ok := d.GetOk("password"); ok {
		connConfig.Password = value.(string)
	}
	if value, ok := d.GetOk("private_key"); ok {
		privateKey := value.(string)
		connConfig.PrivateKey = strings.ReplaceAll(privateKey, "\r\n", "\\n")
	}
	if value, ok := d.GetOk("project"); ok {
		connConfig.Project = value.(string)
	}
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
	if value, ok := d.GetOk("session_token"); ok {
		connConfig.SessionToken = value.(string)
	}
	if value, ok := d.GetOk("subdomain"); ok {
		connConfig.Subdomain = value.(string)
	}
	if value, ok := d.GetOk("subscription_id"); ok {
		connConfig.SubscriptionID = value.(string)
	}
	if value, ok := d.GetOk("tables"); ok {
		var tables []string
		for _, item := range value.([]interface{}) {
			tables = append(tables, item.(string))
		}
		connConfig.Tables = tables
	}
	if value, ok := d.GetOk("tenancy_ocid"); ok {
		connConfig.TenancyOCID = value.(string)
	}
	if value, ok := d.GetOk("tenant_id"); ok {
		connConfig.TenantID = value.(string)
	}
	if value, ok := d.GetOk("token"); ok {
		connConfig.Token = value.(string)
	}
	if value, ok := d.GetOk("user_ocid"); ok {
		connConfig.UserOCID = value.(string)
	}

	return connConfig, nil
}

type ConnectionConfig struct {
	AccessKey      string   `json:"access_key,omitempty"`
	ApiKey         string   `json:"api_key,omitempty"`
	BaseURL        string   `json:"base_url,omitempty"`
	BearerToken    string   `json:"bearer_token,omitempty"`
	ClientID       string   `json:"client_id,omitempty"`
	ClientSecret   string   `json:"client_secret,omitempty"`
	Credentials    string   `json:"credentials,omitempty"`
	DatabaseID     string   `json:"database_id,omitempty"`
	Email          string   `json:"email,omitempty"`
	Environment    string   `json:"environment,omitempty"`
	Fingerprint    string   `json:"fingerprint,omitempty"`
	MaxItems       int      `json:"max_items,omitempty"`
	Password       string   `json:"password,omitempty"`
	PrivateKey     string   `json:"private_key,omitempty"`
	Project        string   `json:"project,omitempty"`
	Regions        []string `json:"regions,omitempty"`
	SecretKey      string   `json:"secret_key,omitempty"`
	SessionToken   string   `json:"session_token,omitempty"`
	Subdomain      string   `json:"subdomain,omitempty"`
	SubscriptionID string   `json:"subscription_id,omitempty"`
	Tables         []string `json:"tables,omitempty"`
	TenancyOCID    string   `json:"tenancy_ocid,omitempty"`
	TenantID       string   `json:"tenant_id,omitempty"`
	Token          string   `json:"token,omitempty"`
	UserOCID       string   `json:"user_ocid,omitempty"`
}
