package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/turbot/go-kit/types"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrganizationMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationMemberCreate,
		ReadContext:   resourceOrganizationMemberRead,
		DeleteContext: resourceOrganizationMemberDelete,
		UpdateContext: resourceOrganizationMemberUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"email"},
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_handle"},
			},
			"association_id": {
				Type:     schema.TypeString,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// CRUD functions

func resourceOrganizationMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	client := meta.(*SteampipeClient)

	// Return if `organization` is not given in config
	if client.Config != nil && client.Config.Organization == "" {
		return diag.Errorf("failed to get organization. Please set 'organization' in provider config")
	}

	// Create request
	req := steampipe.InviteOrgUserRequest{
		Role: d.Get("role").(string),
	}

	if value, ok := d.GetOk("user_handle"); ok {
		req.Handle = types.String(value.(string))
	}
	if value, ok := d.GetOk("email"); ok {
		req.Email = types.String(value.(string))
	}

	// Return if both handle and email are empty
	if req.Handle == nil && req.Email == nil {
		return diag.Errorf("either 'user_handle' or 'email' must be set in provider config")
	}

	// Invite requested member
	resp, r, err := client.APIClient.OrgMembers.Invite(ctx, client.Config.Organization).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error inviting member: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member invited: %v", resp)

	/*
		 * If a member is invited using user handle, use `OrgMembers.Get` to fetch the user details
		 * If a member is invited using an email;
		   * List the invited users, and find the requested user; if found return the requested user
			 * else, list the accepted users, and find the requested user; if found return the requested user
		 * TODO:: As of Dec 15, 2021, SDK doesn't support `email` in `OrgMembers.Get` API. If the API supports `email`, list operations can be ignored.
	*/
	var orgMemberDetails steampipe.OrgUser
	if req.Handle != nil {
		resp, r, err := client.APIClient.OrgMembers.Get(ctx, client.Config.Organization, *req.Handle).Execute()
		if err != nil {
			if r.StatusCode == 404 {
				return diag.Errorf("requested member %s not found", *req.Handle)
			}
			return diag.Errorf("error reading member %s.\nerr: %s", *req.Handle, decodeResponse(r))
		}
		orgMemberDetails = resp
	} else {
		data, err := listOrganizationMembersInvited(d, meta, req.Handle, req.Email)
		if data.Id == "" {
			data, err = listOrganizationMembersAccepted(d, meta, req.Handle, req.Email)
		}

		if err != nil {
			return diag.Errorf("error fetching member from the list.\nerr: %s", decodeResponse(r))
		}
		orgMemberDetails = data
	}

	// Set property values
	d.SetId(fmt.Sprintf("%s:%s", client.Config.Organization, orgMemberDetails.UserHandle))
	d.Set("user_handle", orgMemberDetails.UserHandle)
	d.Set("created_at", orgMemberDetails.CreatedAt)
	d.Set("email", orgMemberDetails.Email)
	d.Set("association_id", orgMemberDetails.Id)
	d.Set("organization_id", orgMemberDetails.OrgId)
	d.Set("role", orgMemberDetails.Role)
	d.Set("status", orgMemberDetails.Status)
	d.Set("updated_at", orgMemberDetails.UpdatedAt)
	d.Set("user_id", orgMemberDetails.UserId)
	d.Set("version_id", orgMemberDetails.VersionId)

	if orgMemberDetails.User != nil {
		d.Set("display_name", orgMemberDetails.User.DisplayName)
	}

	return diags
}

func resourceOrganizationMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Return if `org` is not given in config
	if client.Config != nil && client.Config.Organization == "" {
		return diag.Errorf("failed to get organization. Please set 'organization' in provider config")
	}

	id := d.Id()
	idParts := strings.Split(id, ":")
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
	}

	if idParts[0] != client.Config.Organization {
		return diag.Errorf("given organization_handle does not match with the organization in the provider config")
	}

	if strings.Contains(idParts[1], "@") {
		return diag.Errorf("invalid user_handle. Please provide valid user_handle to import")
	}
	userHandle := idParts[1]

	resp, r, err := client.APIClient.OrgMembers.Get(context.Background(), client.Config.Organization, userHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Member (%s) not found", userHandle)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading %s:%s.\nerr: %s", client.Config.Organization, userHandle, decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization Member received: %s", id)

	d.SetId(id)
	d.Set("user_handle", resp.UserHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("email", resp.Email)
	d.Set("association_id", resp.Id)
	d.Set("organization_id", resp.OrgId)
	d.Set("role", resp.Role)
	d.Set("status", resp.Status)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("user_id", resp.UserId)
	d.Set("version_id", resp.VersionId)

	if resp.User != nil {
		d.Set("display_name", resp.User.DisplayName)
	}

	return diags
}

func resourceOrganizationMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Return if `org` is not given in config
	if client.Config != nil && client.Config.Organization == "" {
		return diag.Errorf("failed to get organization. Please set 'org' in provider config")
	}

	userHandle := d.Get("user_handle").(string)
	role := d.Get("role").(string)

	// Create request
	req := steampipe.UpdateOrgUserRequest{
		Role: role,
	}

	log.Printf("\n[DEBUG] Updating membership: '%s:%s'", client.Config.Organization, userHandle)

	resp, r, err := client.APIClient.OrgMembers.Update(context.Background(), client.Config.Organization, userHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating membership: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Membership updated: %s:%s", client.Config.Organization, resp.UserHandle)

	// Update state file
	id := fmt.Sprintf("%s:%s", client.Config.Organization, resp.UserHandle)
	d.SetId(id)
	d.Set("user_handle", resp.UserHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("email", resp.Email)
	d.Set("association_id", resp.Id)
	d.Set("organization_id", resp.OrgId)
	d.Set("role", resp.Role)
	d.Set("status", resp.Status)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("user_id", resp.UserId)
	d.Set("version_id", resp.VersionId)

	if resp.User != nil {
		d.Set("display_name", resp.User.DisplayName)
	}

	return diags
}

func resourceOrganizationMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Return if `org` is not given in config
	if client.Config != nil && client.Config.Organization == "" {
		return diag.Errorf("failed to get organization. Please set 'organization' in provider config")
	}

	id := d.Id()
	idParts := strings.Split(id, ":")
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
	}
	log.Printf("\n[DEBUG] Removing membership: %s", id)

	_, r, err := client.APIClient.OrgMembers.Delete(context.Background(), client.Config.Organization, idParts[1]).Execute()
	if err != nil {
		return diag.Errorf("error removing membership %s: %s", id, decodeResponse(r))
	}
	d.SetId("")

	return diags
}

// List all the members who has been invited to the org.
func listOrganizationMembersInvited(d *schema.ResourceData, meta interface{}, handle *string, email *string) (steampipe.OrgUser, error) {
	client := meta.(*SteampipeClient)

	// Return if `org` is not given in config
	if client.Config != nil && client.Config.Organization == "" {
		return steampipe.OrgUser{}, fmt.Errorf("failed to get organization. Please set 'organization' in provider config")
	}

	pagesLeft := true
	var resp steampipe.ListOrgUsersResponse
	var err error

	for pagesLeft {
		if resp.NextToken != nil {
			resp, _, err = client.APIClient.OrgMembers.ListInvited(context.Background(), client.Config.Organization).NextToken(*resp.NextToken).Execute()
		} else {
			resp, _, err = client.APIClient.OrgMembers.ListInvited(context.Background(), client.Config.Organization).Execute()
		}

		if err != nil {
			return steampipe.OrgUser{}, err
		}

		for _, i := range *resp.Items {
			if (email != nil && i.Email == *email) || (handle != nil && i.UserHandle == *handle) {
				return i, nil
			}
		}
	}

	return steampipe.OrgUser{}, nil
}

// List all the members who has accepted the request.
func listOrganizationMembersAccepted(d *schema.ResourceData, meta interface{}, handle *string, email *string) (steampipe.OrgUser, error) {
	client := meta.(*SteampipeClient)

	// Return if `org` is not given in config
	if client.Config != nil && client.Config.Organization == "" {
		return steampipe.OrgUser{}, fmt.Errorf("failed to get organization. Please set 'organization' in provider config")
	}

	pagesLeft := true
	var resp steampipe.ListOrgUsersResponse
	var err error

	for pagesLeft {
		if resp.NextToken != nil {
			resp, _, err = client.APIClient.OrgMembers.ListAccepted(context.Background(), client.Config.Organization).NextToken(*resp.NextToken).Execute()
		} else {
			resp, _, err = client.APIClient.OrgMembers.ListAccepted(context.Background(), client.Config.Organization).Execute()
		}

		if err != nil {
			return steampipe.OrgUser{}, err
		}

		for _, i := range *resp.Items {
			if (email != nil && i.Email == *email) || (handle != nil && i.UserHandle == *handle) {
				return i, nil
			}
		}
	}

	return steampipe.OrgUser{}, nil
}
