package steampipecloud

import (
	"context"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceSteampipeCloudUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSteampipeCloudUserRead,
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

func dataSourceSteampipeCloudUserRead(d *schema.ResourceData, meta interface{}) error {
	steampipeClient := meta.(*SteampipeClient)

	resp, _, err := steampipeClient.APIClient.UsersApi.GetActor(context.Background()).Execute()
	if err != nil {
		return err
	}

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

	return nil
}
