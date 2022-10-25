package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceUserPreferences() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserPreferencesRead,
		ReadContext:   resourceUserPreferencesRead,
		UpdateContext: resourceUserPreferencesUpdate,
		DeleteContext: resourceUserPreferencesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"communication_community_updates": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"communication_product_updates": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"communication_tips_and_tricks": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUserPreferencesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*SteampipeClient)

	user, r, err := client.APIClient.Actors.Get(context.Background()).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Actor information not found")
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading actor information: %v", decodeResponse(r))
	}

	resp, r, err := client.APIClient.Users.GetPreferences(context.Background(), user.Handle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] User Preferences not found")
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading user preferences: %v", decodeResponse(r))
	}
	log.Printf("\n[INFO] Received User Preferences : %v", resp)

	d.SetId(fmt.Sprintf("%s/preferences", user.Handle))
	d.Set("communication_community_updates", resp.CommunicationCommunityUpdates)
	d.Set("communication_product_updates", resp.CommunicationProductUpdates)
	d.Set("communication_tips_and_tricks", resp.CommunicationTipsAndTricks)
	d.Set("created_at", resp.CreatedAt)
	if resp.UpdatedAt != nil {
		d.Set("updated_at", resp.UpdatedAt)
	}
	d.Set("version_id", resp.VersionId)

	return diags
}

func resourceUserPreferencesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var userHandle string

	client := meta.(*SteampipeClient)

	id := d.Id()

	if len(id) > 0 {
		ids := strings.Split(id, "/")
		if len(ids) == 2 {
			userHandle = ids[0]
		}
	} else {
		user, r, err := client.APIClient.Actors.Get(context.Background()).Execute()
		if err != nil {
			if r.StatusCode == 404 {
				log.Printf("\n[WARN] Actor information not found")
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading actor information: %v", decodeResponse(r))
		}
		userHandle = user.Handle
	}

	var req steampipe.UpdateUserPreferencesRequest
	if value, ok := d.GetOk("communication_community_updates"); ok {
		req.CommunicationCommunityUpdates = types.String(value.(string))
	}
	if value, ok := d.GetOk("communication_product_updates"); ok {
		req.CommunicationProductUpdates = types.String(value.(string))
	}
	if value, ok := d.GetOk("communication_tips_and_tricks"); ok {
		req.CommunicationTipsAndTricks = types.String(value.(string))
	}

	resp, r, err := client.APIClient.Users.UpdatePreferences(context.Background(), userHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating user preferences: %v", decodeResponse(r))
	}

	d.SetId(fmt.Sprintf("%s/preferences", userHandle))
	d.Set("communication_community_updates", resp.CommunicationCommunityUpdates)
	d.Set("communication_product_updates", resp.CommunicationProductUpdates)
	d.Set("communication_tips_and_tricks", resp.CommunicationTipsAndTricks)
	d.Set("created_at", resp.CreatedAt)
	if resp.UpdatedAt != nil {
		d.Set("updated_at", resp.UpdatedAt)
	}
	d.Set("version_id", resp.VersionId)

	return diags
}

func resourceUserPreferencesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var userHandle string

	client := meta.(*SteampipeClient)

	ids := strings.Split(d.Id(), "/")
	if len(ids) == 2 {
		userHandle = ids[0]
	}

	var req steampipe.UpdateUserPreferencesRequest
	req.CommunicationCommunityUpdates = types.String("enabled")
	req.CommunicationProductUpdates = types.String("enabled")
	req.CommunicationTipsAndTricks = types.String("enabled")

	_, r, err := client.APIClient.Users.UpdatePreferences(context.Background(), userHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error resetting user preferences: %v", decodeResponse(r))
	}
	log.Printf("\n[INFO] Setting ID to blank string")
	d.SetId("")

	return diags
}
