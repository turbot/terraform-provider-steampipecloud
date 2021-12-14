package steampipecloud

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSteampipeCloudUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSteampipeCloudUserRead,
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

func dataSourceSteampipeCloudUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	steampipeClient := meta.(*SteampipeClient)
	resp, r, err := steampipeClient.APIClient.Actors.Get(context.Background()).Execute()
	if err != nil {
		errBody := make([]map[string]interface{}, 0)
		err = json.NewDecoder(r.Body).Decode(&errBody)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	defer r.Body.Close()

	d.SetId(resp.Handle)
	d.Set("avatar_url", resp.AvatarUrl)
	d.Set("created_at", resp.CreatedAt)
	d.Set("display_name", resp.DisplayName)
	d.Set("email", resp.Email)
	d.Set("handle", resp.Handle)
	d.Set("user_id", resp.Id)
	d.Set("preview_access_mode", resp.PreviewAccessMode)
	d.Set("status", resp.Status)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("url", resp.Url)
	d.Set("version_id", resp.VersionId)

	return diags
}
