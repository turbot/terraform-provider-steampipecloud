package steampipecloud

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strings"

// 	"github.com/turbot/go-kit/types"

// 	"github.com/hashicorp/terraform/helper/schema"
// 	"github.com/turbot/steampipe-cloud-sdk-go"
// )

// func resourceSteampipeCloudOrganizationMember() *schema.Resource {
// 	return &schema.Resource{
// 		Create: resourceSteampipeCloudOrganizationMemberCreate,
// 		Read:   resourceSteampipeCloudOrganizationMemberRead,
// 		Delete: resourceSteampipeCloudOrganizationMemberDelete,
// 		Update: resourceSteampipeCloudOrganizationMemberUpdate,
// 		Exists: resourceSteampipeCloudOrganizationMemberExists,
// 		Importer: &schema.ResourceImporter{
// 			State: resourceSteampipeCloudOrganizationMemberImport,
// 		},
// 		Schema: map[string]*schema.Schema{
// 			"user_handle": {
// 				Type:          schema.TypeString,
// 				Optional:      true,
// 				ConflictsWith: []string{"email"},
// 			},
// 			"role": {
// 				Type:     schema.TypeString,
// 				Required: true,
// 			},
// 			"email": {
// 				Type:          schema.TypeString,
// 				Optional:      true,
// 				ConflictsWith: []string{"user_handle"},
// 			},
// 		},
// 	}
// }

// func resourceSteampipeCloudOrganizationMemberExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
// 	client := meta.(*SteampipeClient)

// 	// Return if `org` is not given in config
// 	if client.Config != nil && client.Config.Org == "" {
// 		return false, fmt.Errorf("failed to get organization. Please set 'org' in provider config")
// 	}

// 	id := d.Id()
// 	idParts := strings.Split(id, ":")
// 	if len(idParts) < 2 {
// 		return false, fmt.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
// 	}

// 	orgHandle := idParts[0]
// 	userHandle := idParts[1]

// 	_, r, err := client.APIClient.OrgMembersApi.GetOrgMember(context.Background(), orgHandle, userHandle).Execute()
// 	if err != nil {
// 		if r.StatusCode == 404 {
// 			return false, nil
// 		}
// 		return false, fmt.Errorf("%s", r.Status)
// 	}
// 	return true, nil
// }

// func resourceSteampipeCloudOrganizationMemberImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// 	if err := resourceSteampipeCloudOrganizationRead(d, meta); err != nil {
// 		return nil, err
// 	}

// 	return []*schema.ResourceData{d}, nil
// }

// func resourceSteampipeCloudOrganizationMemberCreate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*SteampipeClient)

// 	// Return if `org` is not given in config
// 	if client.Config != nil && client.Config.Org == "" {
// 		return fmt.Errorf("failed to get organization. Please set 'org' in provider config")
// 	}

// 	// Create request
// 	req := steampipe.TypesInviteOrgUserRequest{
// 		Role: d.Get("role").(string),
// 	}

// 	if value, ok := d.GetOk("user_handle"); ok {
// 		req.Handle = types.String(value.(string))
// 	}

// 	if value, ok := d.GetOk("email"); ok {
// 		req.Email = types.String(value.(string))
// 	}

// 	if req.Handle == nil && req.Email == nil {
// 		return fmt.Errorf("either 'user_handle' or 'email' must be set in provider config")
// 	}

// 	resp, r, err := client.APIClient.OrgMembersApi.InviteOrgMember(context.Background(), client.Config.Org).Request(req).Execute()
// 	if err != nil {
// 		return fmt.Errorf("error inviting member: %s - %s", r.Status, r.Body)
// 	}
// 	log.Printf("\n[DEBUG] Member invited: %v", resp)

// 	// Set property values
// 	var id string
// 	if resp.UserHandle != "" {
// 		id = fmt.Sprintf("handle:%s", resp.UserHandle)
// 	} else {
// 		id = fmt.Sprintf("email:%s", resp.Email)
// 	}
// 	d.SetId(id)
// 	d.Set("user_handle", resp.UserHandle)
// 	d.Set("email", resp.Email)
// 	d.Set("role", resp.Role)

// 	return nil
// }

// func resourceSteampipeCloudOrganizationMemberRead(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*SteampipeClient)

// 	// Return if `org` is not given in config
// 	if client.Config != nil && client.Config.Org == "" {
// 		return fmt.Errorf("failed to get organization. Please set 'org' in provider config")
// 	}

// 	id := d.Id()
// 	idParts := strings.Split(id, ":")
// 	if len(idParts) < 2 {
// 		return fmt.Errorf("unexpected format of ID (%q), expected <input_type>:<value>; i.e. email:abc@gmail.com", id)
// 	}

// 	var userHandle, userEmail string
// 	if idParts[0] == "handle" {
// 		userHandle = idParts[1]
// 	} else {
// 		userEmail = idParts[1]
// 	}

// 	var resp steampipe.TypesOrgUser
// 	var err error
// 	var r *http.Response

// 	if userHandle != "" {
// 		resp, r, err = client.APIClient.OrgMembersApi.GetOrgMember(context.Background(), client.Config.Org, userHandle).Execute()
// 		if err != nil {
// 			if r.StatusCode == 404 {
				
// 			}
// 		}
// 	}

// 	resp, r, err := client.APIClient.OrgMembersApi.GetOrgMember(context.Background(), client.Config.Org, userHandle).Execute()

// 	resp, r, err := client.APIClient.OrgMembersApi.GetOrgMember(context.Background(), client.Config.Org, userHandle).Execute()
// 	if err != nil {
// 		if r.StatusCode == 404 {
// 			log.Printf("\n[WARN] Member (%s) not found", userHandle)
// 			d.SetId("")
// 			return nil
// 		}
// 		return fmt.Errorf("error reading %s:%s %s", client.Config.Org, userHandle, err)
// 	}
// 	log.Printf("\n[DEBUG] Organization Member received: %s", id)

// 	d.SetId(id)
// 	d.Set("user_handle", resp.UserHandle)
// 	d.Set("created_at", resp.CreatedAt)
// 	d.Set("email", resp.Email)
// 	d.Set("association_id", resp.Id)
// 	d.Set("organization_id", resp.OrgId)
// 	d.Set("role", resp.Role)
// 	d.Set("status", resp.Status)
// 	d.Set("updated_at", resp.UpdatedAt)
// 	d.Set("user_id", resp.UserId)
// 	d.Set("version_id", resp.VersionId)

// 	if resp.User != nil {
// 		d.Set("display_name", resp.User.DisplayName)
// 	}

// 	return nil
// }

// func resourceSteampipeCloudOrganizationMemberUpdate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*SteampipeClient)

// 	// Return if `org` is not given in config
// 	if client.Config != nil && client.Config.Org == "" {
// 		return fmt.Errorf("failed to get organization. Please set 'org' in provider config")
// 	}

// 	userHandle := d.Get("user_handle").(string)
// 	role := d.Get("role").(string)

// 	// Create request
// 	req := steampipe.TypesUpdateOrgUserRequest{
// 		Role: role,
// 	}

// 	log.Printf("\n[DEBUG] Updating membership: '%s:%s'", client.Config.Org, userHandle)

// 	resp, _, err := client.APIClient.OrgMembersApi.UpdateOrgMember(context.Background(), client.Config.Org, userHandle).Request(req).Execute()
// 	if err != nil {
// 		return fmt.Errorf("error updating membership: %s", err)
// 	}
// 	log.Printf("\n[DEBUG] Membership updated: %s:%s", client.Config.Org, resp.UserHandle)

// 	// Update state file
// 	id := fmt.Sprintf("%s:%s", client.Config.Org, resp.UserHandle)
// 	d.SetId(id)
// 	d.Set("user_handle", resp.UserHandle)
// 	d.Set("created_at", resp.CreatedAt)
// 	d.Set("email", resp.Email)
// 	d.Set("association_id", resp.Id)
// 	d.Set("organization_id", resp.OrgId)
// 	d.Set("role", resp.Role)
// 	d.Set("status", resp.Status)
// 	d.Set("updated_at", resp.UpdatedAt)
// 	d.Set("user_id", resp.UserId)
// 	d.Set("version_id", resp.VersionId)

// 	if resp.User != nil {
// 		d.Set("display_name", resp.User.DisplayName)
// 	}

// 	return nil
// }

// func resourceSteampipeCloudOrganizationMemberDelete(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*SteampipeClient)

// 	// Return if `org` is not given in config
// 	if client.Config != nil && client.Config.Org == "" {
// 		return fmt.Errorf("failed to get organization. Please set 'org' in provider config")
// 	}

// 	id := d.Id()
// 	idParts := strings.Split(id, ":")
// 	if len(idParts) < 2 {
// 		return fmt.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
// 	}
// 	log.Printf("\n[DEBUG] Removing membership: %s", id)

// 	_, _, err := client.APIClient.OrgMembersApi.DeleteOrgMember(context.Background(), client.Config.Org, idParts[1]).Execute()
// 	if err != nil {
// 		return fmt.Errorf("error removing membership %s: %s", id, err)
// 	}
// 	d.SetId("")

// 	return nil
// }
