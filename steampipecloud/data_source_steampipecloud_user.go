package steampipecloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Description: "",
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"preview_access_mode": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	steampipeClient := meta.(*SteampipeClient)
	resp, r, err := steampipeClient.APIClient.Actors.Get(context.Background()).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("%v", decodeResponse(r)))
	}
	defer r.Body.Close()

	d.SetId(resp.Handle)
	if err := d.Set("avatar_url", resp.AvatarUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", resp.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", resp.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", resp.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("handle", resp.Handle); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_id", resp.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("preview_access_mode", resp.PreviewAccessMode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", resp.Status); err != nil {
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

	return diags
}
