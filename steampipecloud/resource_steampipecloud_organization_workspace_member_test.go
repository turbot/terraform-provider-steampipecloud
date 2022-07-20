// NOTE: Please provide a valid email in the config before performing the test

package steampipecloud

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccOrganizationWorkspaceMember_Basic(t *testing.T) {
	orgHandle := "terraform" + randomString(3)
	workspaceHandle := "dev" + randomString(3)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationWorkspaceMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationWorkspaceMemberConfig(orgHandle, workspaceHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipOrganizationExists(orgHandle),
					testAccCheckMembershipWorkspaceExists(orgHandle, workspaceHandle),
					testAccCheckOrganizationWorkspaceMemberExists("steampipecloud_organization_workspace_member.test"),
					resource.TestCheckResourceAttr(
						"steampipecloud_organization_workspace_member.test", "role", "reader"),
				),
			},
			{
				Config: testAccOrganizationWorkspaceMemberUpdateConfig(orgHandle, workspaceHandle),
				Check: resource.TestCheckResourceAttr(
					"steampipecloud_organization_workspace_member.test", "role", "owner"),
			},
		},
	})
}

// configs
func testAccOrganizationWorkspaceMemberConfig(orgHandle, workspaceHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle = "%s"
}

resource "steampipecloud_workspace" "test" {
	organization = steampipecloud_organization.test.handle
	handle = "%s"
}

# Invite the user to the organization
resource "steampipecloud_organization_member" "test" {
	organization = steampipecloud_organization.test.handle
	email        = "user@domain.com"
	role         = "member"
}

# Add the user to the workspace
resource "steampipecloud_organization_workspace_member" "test" {
	organization = steampipecloud_organization.test.handle
	workspace_handle = steampipecloud_workspace.test.handle
	user_handle        = steampipecloud_organization_member.test.user_handle
	role         = "reader"
}`, orgHandle, workspaceHandle)
}

func testAccOrganizationWorkspaceMemberUpdateConfig(orgHandle, workspaceHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle = "%s"
}

resource "steampipecloud_workspace" "test" {
	organization = steampipecloud_organization.test.handle
	handle = "%s"
}

# Invite the user to the organization
resource "steampipecloud_organization_member" "test" {
	organization = steampipecloud_organization.test.handle
	email        = "user@domain.com"
	role         = "member"
}

# Please provide a valid email
resource "steampipecloud_organization_workspace_member" "test" {
	organization = steampipecloud_organization.test.handle
	workspace_handle = steampipecloud_workspace.test.handle
	user_handle        = steampipecloud_organization_member.test.user_handle
	role         = "owner"
}`, orgHandle, workspaceHandle)
}

// helper functions
func testAccCheckOrganizationWorkspaceMemberExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		// Extract organization handle and user handle from ID
		id := rs.Primary.ID
		idParts := strings.Split(id, ":")
		if len(idParts) < 3 {
			return fmt.Errorf("unexpected format of ID (%q), expected <organization_handle>:<workspace_handle>:<user_handle>", id)
		}

		client := testAccProvider.Meta().(*SteampipeClient)
		_, _, err := client.APIClient.OrgWorkspaceMembers.Get(context.Background(), idParts[0], idParts[1], idParts[2]).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckOrganizationWorkspaceMemberDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*SteampipeClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "steampipecloud_organization_workspace_member" {
			// Extract organization handle and user handle from ID
			id := rs.Primary.ID
			idParts := strings.Split(id, ":")
			if len(idParts) < 3 {
				return fmt.Errorf("unexpected format of ID (%q), expected <organization_handle>:<workspace_handle>:<user_handle>", id)
			}

			_, r, err := client.APIClient.OrgWorkspaceMembers.Get(context.Background(), idParts[0], idParts[1], idParts[2]).Execute()
			if err == nil {
				return fmt.Errorf("organization member still exists")
			}

			// If a organization is deleted, all the members will lost access to that organization
			// If anyone try to get that deleted resource, it will always return `403 Forbidden` error
			if r.StatusCode != 403 {
				return fmt.Errorf("expected 'forbidden' error, got %s", err)
			}
		}
	}

	return nil
}

func testAccCheckMembershipWorkspaceExists(orgHandle, workspaceHandle string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)
		ctx := context.Background()
		var err error
		var r *http.Response

		// check if organization  is created
		_, r, err = client.APIClient.OrgWorkspaces.Get(ctx, orgHandle, workspaceHandle).Execute()
		if err != nil {
			if r.StatusCode != 403 {
				return fmt.Errorf("error fetching workspace with handle %s. %s", workspaceHandle, err)
			}
		}
		return nil
	}
}
