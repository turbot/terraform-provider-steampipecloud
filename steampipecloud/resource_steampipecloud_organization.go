package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceSteampipeCloudOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeCloudOrganizationCreate,
		Read:   resourceSteampipeCloudOrganizationRead,
		Delete: resourceSteampipeCloudOrganizationDelete,
		Update: resourceSteampipeCloudOrganizationUpdate,
		Exists: resourceSteampipeCloudOrganizationExists,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudOrganizationImport,
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

func resourceSteampipeCloudOrganizationExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*SteampipeClient)
	handle := d.Id()

	_, r, err := client.APIClient.Orgs.Get(context.Background(), handle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceSteampipeCloudOrganizationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudOrganizationRead(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeCloudOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
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

	resp, r, err := client.APIClient.Orgs.Create(context.Background()).Request(req).Execute()
	if err != nil {
		return fmt.Errorf("error creating organization: \n	StatusCode: %d \n	Body: %v", r.StatusCode, r.Body)
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

	return nil
}

func resourceSteampipeCloudOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)
	handle := d.Id()

	resp, r, err := client.APIClient.Orgs.Get(context.Background(), handle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Organization (%s) not found", handle)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading %s: %s", handle, err)
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

	return nil
}

func resourceSteampipeCloudOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
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

	resp, _, err := client.APIClient.Orgs.Update(context.Background(), oldHandle.(string)).Request(req).Execute()
	if err != nil {
		return fmt.Errorf("error updating organization: %s", err)
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

	return nil
}

func resourceSteampipeCloudOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SteampipeClient)
	handle := d.Id()
	log.Printf("\n[DEBUG] Deleting Organization: %s", handle)

	_, _, err := client.APIClient.Orgs.Delete(context.Background(), handle).Execute()
	if err != nil {
		return fmt.Errorf("error deleting organization: %s", err)
	}
	d.SetId("")

	return nil
}
