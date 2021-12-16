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
func TestAccOrganizationMember_Basic(t *testing.T) {
	orgHandle := "terraform" + randomString(3)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationMemberConfig(orgHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipOrganizationExists(orgHandle),
					testAccCheckOrganizationMemberExists("steampipecloud_organization_member.test"),
					resource.TestCheckResourceAttr(
						"steampipecloud_organization_member.test", "role", "member"),
				),
			},
			{
				Config: testAccOrganizationMemberUpdateConfig(orgHandle),
				Check: resource.TestCheckResourceAttr(
					"steampipecloud_organization_member.test", "role", "owner"),
			},
		},
	})
}

// configs
func testAccOrganizationMemberConfig(orgHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle = "%s"
}

# Please provide a valid email
resource "steampipecloud_organization_member" "test" {
	organization = steampipecloud_organization.test.handle
	email        = "user@domain.com"
	role         = "member"
}`, orgHandle)
}

func testAccOrganizationMemberUpdateConfig(orgHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle = "%s"
}

# Please provide a valid email
resource "steampipecloud_organization_member" "test" {
  organization = steampipecloud_organization.test.handle
  email        = "user@domain.com"
  role         = "owner"
}`, orgHandle)
}

// helper functions
func testAccCheckOrganizationMemberExists(resource string) resource.TestCheckFunc {
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
		if len(idParts) < 2 {
			return fmt.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
		}

		client := testAccProvider.Meta().(*SteampipeClient)
		_, _, err := client.APIClient.OrgMembers.Get(context.Background(), idParts[0], idParts[1]).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckOrganizationMemberDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*SteampipeClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "steampipecloud_organization_member" {
			// Extract organization handle and user handle from ID
			id := rs.Primary.ID
			idParts := strings.Split(id, ":")
			if len(idParts) < 2 {
				return fmt.Errorf("unexpected format of ID (%q), expected <organization_handle>:<user_handle>", id)
			}

			_, r, err := client.APIClient.OrgMembers.Get(context.Background(), idParts[0], idParts[1]).Execute()
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

func testAccCheckMembershipOrganizationExists(orgHandle string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)
		ctx := context.Background()
		var err error
		var r *http.Response

		// check if organization  is created
		_, r, err = client.APIClient.Orgs.Get(ctx, orgHandle).Execute()
		if err != nil {
			if r.StatusCode != 403 {
				return fmt.Errorf("error fetching organization with handle %s. %s", orgHandle, err)
			}
		}
		return nil
	}
}
