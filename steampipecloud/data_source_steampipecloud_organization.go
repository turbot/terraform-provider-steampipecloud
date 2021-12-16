package steampipecloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrganizationRead,
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"handle", "organization_id"},
			},
			"organization_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"handle", "organization_id"},
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
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

func dataSourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var handle string
	handle = d.Get("handle").(string)
	if strings.TrimSpace(handle) == "" {
		handle = d.Get("organization_id").(string)
	}
	if strings.TrimSpace(handle) == "" {
		return diags
	}

	resp, r, err := client.APIClient.Orgs.Get(ctx, handle).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("%v", decodeResponse(r)))
	}

	if err := d.Set("handle", resp.Handle); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("organization_id", resp.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("avatar_url", resp.AvatarUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", resp.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", resp.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", resp.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", resp.Url); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version_id", resp.VersionId); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Id)

	return diags
}
