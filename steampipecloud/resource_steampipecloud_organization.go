package steampipecloud

import (
	"context"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceOrganization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationCreate,
		ReadContext:   resourceOrganizationRead,
		UpdateContext: resourceOrganizationUpdate,
		DeleteContext: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,37}[a-z0-9]$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single hyphens, and cannot begin or end with a hyphen."),
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
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

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := meta.(*SteampipeClient)
	handle := d.Get("handle")

	// Create request
	req := steampipe.TypesCreateOrgRequest{
		Handle: handle.(string),
	}

	if value, ok := d.GetOk("avatar_url"); ok {
		req.AvatarUrl = types.String(value.(string))
	}

	if value, ok := d.GetOk("display_name"); ok {
		req.DisplayName = types.String(value.(string))
	}

	if value, ok := d.GetOk("url"); ok {
		req.Url = types.String(value.(string))
	}

	resp, r, err := client.APIClient.Orgs.Create(ctx).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error creating organization: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization created: %s", resp.Handle)

	// Set property values
	d.SetId(resp.Handle)
	d.Set("handle", handle)
	d.Set("avatar_url", resp.AvatarUrl)
	d.Set("created_at", resp.CreatedAt)
	d.Set("display_name", resp.DisplayName)
	d.Set("organization_id", resp.Id)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("url", resp.Url)
	d.Set("version_id", resp.VersionId)

	return diags
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := meta.(*SteampipeClient)
	handle := d.Id()

	resp, r, err := client.APIClient.Orgs.Get(context.Background(), handle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Organization (%s) not found", handle)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading organization %s: %v", handle, decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization received: %s", resp.Handle)

	d.Set("handle", handle)
	d.Set("avatar_url", resp.AvatarUrl)
	d.Set("created_at", resp.CreatedAt)
	d.Set("display_name", resp.DisplayName)
	d.Set("organization_id", resp.Id)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("url", resp.Url)
	d.Set("version_id", resp.VersionId)

	return diags
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := meta.(*SteampipeClient)

	oldHandle, newHandle := d.GetChange("handle")

	// Create request
	req := steampipe.TypesUpdateOrgRequest{
		Handle: types.String(newHandle.(string)),
	}

	if value, ok := d.GetOk("avatar_url"); ok {
		req.AvatarUrl = types.String(value.(string))
	}

	if value, ok := d.GetOk("display_name"); ok {
		req.DisplayName = types.String(value.(string))
	}

	if value, ok := d.GetOk("url"); ok {
		req.Url = types.String(value.(string))
	}

	log.Printf("\n[DEBUG] Updating Organization: %s", *req.Handle)

	resp, r, err := client.APIClient.Orgs.Update(ctx, oldHandle.(string)).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceOrganizationUpdate. Update organization %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization updated: %s", resp.Handle)

	// Update state file
	d.SetId(resp.Handle)
	d.Set("handle", resp.Handle)
	d.Set("avatar_url", resp.AvatarUrl)
	d.Set("created_at", resp.CreatedAt)
	d.Set("display_name", resp.DisplayName)
	d.Set("organization_id", resp.Id)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("url", resp.Url)
	d.Set("version_id", resp.VersionId)

	return diags
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)
	handle := d.Id()

	_, r, err := client.APIClient.Orgs.Delete(ctx, handle).Execute()
	if err != nil {
		return diag.Errorf("Deleting organization error: %v", decodeResponse(r))
	}
	d.SetId("")

	return nil
}
