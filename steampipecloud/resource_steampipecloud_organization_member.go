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
			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_member_id": {
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

	// Get the organization
	org := d.Get("organization").(string)

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
		return diag.Errorf("either 'user_handle' or 'email' must be set in resource config")
	}

	// Invite requested member
	_, r, err := client.APIClient.OrgMembers.Invite(ctx, org).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error inviting member: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member invited: %v", decodeResponse(r))

	/*
	 * If a member is invited using user handle, use `OrgMembers.Get` to fetch the user details
	 * If a member is invited using an email;
	   * List the invited users, and find the requested user; if found return the requested user
	   * else, list the accepted users, and find the requested user; if found return the requested user
	 * TODO:: As of Dec 15, 2021, SDK doesn't support `email` in `OrgMembers.Get` API. If the API supports `email`, list operations can be ignored.
	*/
	var orgMemberDetails steampipe.OrgUser
	if req.Handle != nil {
		resp, r, err := client.APIClient.OrgMembers.Get(ctx, org, *req.Handle).Execute()
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
	d.SetId(fmt.Sprintf("%s:%s", org, orgMemberDetails.UserHandle))
	d.Set("user_handle", orgMemberDetails.UserHandle)
	d.Set("created_at", orgMemberDetails.CreatedAt)
	d.Set("email", orgMemberDetails.Email)
	d.Set("organization_member_id", orgMemberDetails.Id)
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

	id := d.Id()
	idParts := strings.Split(id, ":")
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
	}
	org := idParts[0]

	if strings.Contains(idParts[1], "@") {
		return diag.Errorf("invalid user_handle. Please provide valid user_handle to import")
	}
	userHandle := idParts[1]

	resp, r, err := client.APIClient.OrgMembers.Get(context.Background(), org, userHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Member (%s) not found", userHandle)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading %s:%s.\nerr: %s", org, userHandle, decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization Member received: %s", id)

	d.SetId(id)
	d.Set("user_handle", resp.UserHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("email", resp.Email)
	d.Set("organization_member_id", resp.Id)
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

	// Get the organization
	org := d.Get("organization").(string)

	userHandle := d.Get("user_handle").(string)
	role := d.Get("role").(string)

	// Create request
	req := steampipe.UpdateOrgUserRequest{
		Role: role,
	}

	log.Printf("\n[DEBUG] Updating membership: '%s:%s'", org, userHandle)

	resp, r, err := client.APIClient.OrgMembers.Update(context.Background(), org, userHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating membership: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Membership updated: %s:%s", org, resp.UserHandle)

	// Update state file
	id := fmt.Sprintf("%s:%s", org, resp.UserHandle)
	d.SetId(id)
	d.Set("user_handle", resp.UserHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("email", resp.Email)
	d.Set("organization_member_id", resp.Id)
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

	id := d.Id()
	idParts := strings.Split(id, ":")
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
	}
	org := idParts[0]

	log.Printf("\n[DEBUG] Removing membership: %s", id)

	_, r, err := client.APIClient.OrgMembers.Delete(context.Background(), org, idParts[1]).Execute()
	if err != nil {
		return diag.Errorf("error removing membership %s: %s", id, decodeResponse(r))
	}
	d.SetId("")

	return diags
}

// List all the members who has been invited to the org.
func listOrganizationMembersInvited(d *schema.ResourceData, meta interface{}, handle *string, email *string) (steampipe.OrgUser, error) {
	client := meta.(*SteampipeClient)

	// Get the organization
	org := d.Get("organization").(string)

	pagesLeft := true
	var resp steampipe.ListOrgUsersResponse
	var err error

	for pagesLeft {
		if resp.NextToken != nil {
			resp, _, err = client.APIClient.OrgMembers.ListInvited(context.Background(), org).NextToken(*resp.NextToken).Execute()
		} else {
			resp, _, err = client.APIClient.OrgMembers.ListInvited(context.Background(), org).Execute()
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

	// Get the organization
	org := d.Get("organization").(string)

	pagesLeft := true
	var resp steampipe.ListOrgUsersResponse
	var err error

	for pagesLeft {
		if resp.NextToken != nil {
			resp, _, err = client.APIClient.OrgMembers.ListAccepted(context.Background(), org).NextToken(*resp.NextToken).Execute()
		} else {
			resp, _, err = client.APIClient.OrgMembers.ListAccepted(context.Background(), org).Execute()
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
