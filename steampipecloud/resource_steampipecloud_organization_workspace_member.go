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

func resourceOrganizationWorkspaceMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationWorkspaceMemberCreate,
		ReadContext:   resourceOrganizationWorkspaceMemberRead,
		DeleteContext: resourceOrganizationWorkspaceMemberDelete,
		UpdateContext: resourceOrganizationWorkspaceMemberUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization_workspace_member_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"email"},
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_handle"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
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
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

// CRUD functions

func resourceOrganizationWorkspaceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	client := meta.(*SteampipeClient)

	// Get the organization handle
	org := d.Get("organization").(string)

	// Get the workspace handle
	workspace := d.Get("workspace_handle").(string)

	// Create request
	req := steampipe.InviteOrgWorkspaceUserRequest{
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
	_, r, err := client.APIClient.OrgWorkspaceMembers.Invite(ctx, org, workspace).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error inviting member: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member invited: %v", decodeResponse(r))

	/*
	 * If a member is invited using user handle, use `OrgWorkspaceMembers.Get` to fetch the user details
	 * If a member is invited using an email;
	   * List the invited users, and find the requested user; if found return the requested user
	   * else, list the accepted users, and find the requested user; if found return the requested user
	 * TODO:: As of Dec 15, 2021, SDK doesn't support `email` in `OrgWorkspaceMembers.Get` API. If the API supports `email`, list operations can be ignored.
	*/
	var orgWorkspaceMemberDetails steampipe.OrgWorkspaceUser
	if req.Handle != nil {
		resp, r, err := client.APIClient.OrgWorkspaceMembers.Get(ctx, org, workspace, *req.Handle).Execute()
		if err != nil {
			if r.StatusCode == 404 {
				return diag.Errorf("requested member %s not found", *req.Handle)
			}
			return diag.Errorf("error reading member %s.\nerr: %s", *req.Handle, decodeResponse(r))
		}
		orgWorkspaceMemberDetails = resp
	} else {
		data, err := listOrganizationWorkspaceMembersInvited(d, meta, req.Handle, req.Email)
		if data.Id == "" {
			data, err = listOrganizationWorkspaceMembersAccepted(d, meta, req.Handle, req.Email)
		}

		if err != nil {
			return diag.Errorf("error fetching member from the list.\nerr: %s", decodeResponse(r))
		}
		orgWorkspaceMemberDetails = data
	}

	// Set property values
	d.SetId(fmt.Sprintf("%s:%s:%s", org, workspace, orgWorkspaceMemberDetails.UserHandle))
	d.Set("organization_workspace_member_id", orgWorkspaceMemberDetails.Id)
	d.Set("organization_id", orgWorkspaceMemberDetails.OrgId)
	d.Set("workspace_id", orgWorkspaceMemberDetails.WorkspaceId)
	d.Set("user_id", orgWorkspaceMemberDetails.UserId)
	d.Set("user_handle", orgWorkspaceMemberDetails.UserHandle)
	if orgWorkspaceMemberDetails.User != nil {
		d.Set("display_name", orgWorkspaceMemberDetails.User.DisplayName)
	}
	d.Set("email", orgWorkspaceMemberDetails.Email)
	d.Set("status", orgWorkspaceMemberDetails.Status)
	d.Set("role", orgWorkspaceMemberDetails.Role)
	d.Set("created_at", orgWorkspaceMemberDetails.CreatedAt)
	d.Set("updated_at", orgWorkspaceMemberDetails.UpdatedAt)
	if orgWorkspaceMemberDetails.CreatedBy != nil {
		d.Set("created_by", orgWorkspaceMemberDetails.CreatedBy.Handle)
	}
	if orgWorkspaceMemberDetails.UpdatedBy != nil {
		d.Set("updated_by", orgWorkspaceMemberDetails.UpdatedBy.Handle)
	}
	d.Set("version_id", orgWorkspaceMemberDetails.VersionId)

	return diags
}

func resourceOrganizationWorkspaceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	idParts := strings.Split(id, ":")
	if len(idParts) < 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>:<workspace_handle>:<user_handle>", id)
	}
	org := idParts[0]
	workspace := idParts[1]

	if strings.Contains(idParts[2], "@") {
		return diag.Errorf("invalid user_handle. Please provide valid user_handle to import")
	}
	user := idParts[2]

	orgWorkspaceMemberDetails, r, err := client.APIClient.OrgWorkspaceMembers.Get(context.Background(), org, workspace, user).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Member (%s) not found in workspace (%s) of organization (%s)", user, workspace, org)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading %s:%s.\nerr: %s", org, user, decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization Workspace Member received: %s", id)

	// Set the property values
	d.SetId(id)
	d.Set("organization_workspace_member_id", orgWorkspaceMemberDetails.Id)
	d.Set("organization_id", orgWorkspaceMemberDetails.OrgId)
	d.Set("workspace_id", orgWorkspaceMemberDetails.WorkspaceId)
	d.Set("user_id", orgWorkspaceMemberDetails.UserId)
	d.Set("user_handle", orgWorkspaceMemberDetails.UserHandle)
	if orgWorkspaceMemberDetails.User != nil {
		d.Set("display_name", orgWorkspaceMemberDetails.User.DisplayName)
	}
	d.Set("email", orgWorkspaceMemberDetails.Email)
	d.Set("status", orgWorkspaceMemberDetails.Status)
	d.Set("role", orgWorkspaceMemberDetails.Role)
	d.Set("created_at", orgWorkspaceMemberDetails.CreatedAt)
	d.Set("updated_at", orgWorkspaceMemberDetails.UpdatedAt)
	if orgWorkspaceMemberDetails.CreatedBy != nil {
		d.Set("created_by", orgWorkspaceMemberDetails.CreatedBy.Handle)
	}
	if orgWorkspaceMemberDetails.UpdatedBy != nil {
		d.Set("updated_by", orgWorkspaceMemberDetails.UpdatedBy.Handle)
	}
	d.Set("version_id", orgWorkspaceMemberDetails.VersionId)

	return diags
}

func resourceOrganizationWorkspaceMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Get the organization handle
	org := d.Get("organization").(string)
	// Get the workspace handle
	workspace := d.Get("workspace_handle").(string)
	// Get the handle of the user which needs to be updated
	user := d.Get("user_handle").(string)
	// We can only update the role of a user in an organization workspace for now
	role := d.Get("role").(string)

	// Create request
	req := steampipe.UpdateOrgWorkspaceUserRequest{
		Role: role,
	}

	log.Printf("\n[DEBUG] Updating membership: '%s:%s:%s'", org, workspace, user)

	orgWorkspaceMemberDetails, r, err := client.APIClient.OrgWorkspaceMembers.Update(context.Background(), org, workspace, user).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating membership: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Membership updated: %s:%s:%s", org, workspace, user)

	// Update state file
	id := fmt.Sprintf("%s:%s:%s", org, workspace, user)
	d.SetId(id)
	d.Set("organization_workspace_member_id", orgWorkspaceMemberDetails.Id)
	d.Set("organization_id", orgWorkspaceMemberDetails.OrgId)
	d.Set("workspace_id", orgWorkspaceMemberDetails.WorkspaceId)
	d.Set("user_id", orgWorkspaceMemberDetails.UserId)
	d.Set("user_handle", orgWorkspaceMemberDetails.UserHandle)
	if orgWorkspaceMemberDetails.User != nil {
		d.Set("display_name", orgWorkspaceMemberDetails.User.DisplayName)
	}
	d.Set("email", orgWorkspaceMemberDetails.Email)
	d.Set("status", orgWorkspaceMemberDetails.Status)
	d.Set("role", orgWorkspaceMemberDetails.Role)
	d.Set("created_at", orgWorkspaceMemberDetails.CreatedAt)
	d.Set("updated_at", orgWorkspaceMemberDetails.UpdatedAt)
	if orgWorkspaceMemberDetails.CreatedBy != nil {
		d.Set("created_by", orgWorkspaceMemberDetails.CreatedBy.Handle)
	}
	if orgWorkspaceMemberDetails.UpdatedBy != nil {
		d.Set("updated_by", orgWorkspaceMemberDetails.UpdatedBy.Handle)
	}
	d.Set("version_id", orgWorkspaceMemberDetails.VersionId)

	return diags
}

func resourceOrganizationWorkspaceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SteampipeClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	idParts := strings.Split(id, ":")
	if len(idParts) < 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>:<workspace_handle>:<user_handle>", id)
	}
	org := idParts[0]
	workspace := idParts[1]
	user := idParts[2]

	log.Printf("\n[DEBUG] Removing membership: %s", id)

	_, r, err := client.APIClient.OrgWorkspaceMembers.Delete(context.Background(), org, workspace, user).Execute()
	if err != nil {
		return diag.Errorf("error removing membership %s: %s", id, decodeResponse(r))
	}
	d.SetId("")

	return diags
}

// List all the members who has been invited to the org.
func listOrganizationWorkspaceMembersInvited(d *schema.ResourceData, meta interface{}, handle *string, email *string) (steampipe.OrgWorkspaceUser, error) {
	client := meta.(*SteampipeClient)

	// Get the organization handle
	org := d.Get("organization").(string)

	// Get the workspace handle
	workspace := d.Get("workspace_handle").(string)

	pagesLeft := true
	var resp steampipe.ListOrgWorkspaceUsersResponse
	var err error

	for pagesLeft {
		if resp.NextToken != nil {
			resp, _, err = client.APIClient.OrgWorkspaceMembers.ListInvited(context.Background(), org, workspace).NextToken(*resp.NextToken).Execute()
		} else {
			resp, _, err = client.APIClient.OrgWorkspaceMembers.ListInvited(context.Background(), org, workspace).Execute()
		}

		if err != nil {
			return steampipe.OrgWorkspaceUser{}, err
		}

		for _, i := range *resp.Items {
			if (email != nil && i.Email == *email) || (handle != nil && i.UserHandle == *handle) {
				return i, nil
			}
		}
	}

	return steampipe.OrgWorkspaceUser{}, nil
}

// List all the members who has accepted the request.
func listOrganizationWorkspaceMembersAccepted(d *schema.ResourceData, meta interface{}, handle *string, email *string) (steampipe.OrgWorkspaceUser, error) {
	client := meta.(*SteampipeClient)

	// Get the organization handle
	org := d.Get("organization").(string)

	// Get the workspace handle
	workspace := d.Get("workspace_handle").(string)

	pagesLeft := true
	var resp steampipe.ListOrgWorkspaceUsersResponse
	var err error

	for pagesLeft {
		if resp.NextToken != nil {
			resp, _, err = client.APIClient.OrgWorkspaceMembers.ListAccepted(context.Background(), org, workspace).NextToken(*resp.NextToken).Execute()
		} else {
			resp, _, err = client.APIClient.OrgWorkspaceMembers.ListAccepted(context.Background(), org, workspace).Execute()
		}

		if err != nil {
			return steampipe.OrgWorkspaceUser{}, err
		}

		for _, i := range *resp.Items {
			if (email != nil && i.Email == *email) || (handle != nil && i.UserHandle == *handle) {
				return i, nil
			}
		}
	}

	return steampipe.OrgWorkspaceUser{}, nil
}
