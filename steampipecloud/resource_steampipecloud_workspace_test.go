package steampipecloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccUserWorkspace_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace.test"
	workspaceHandle := "workspace" + randomString(3)
	newWorkspaceHandle := "workspace" + randomString(4)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceConfig(workspaceHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserWorkspaceExists("steampipecloud_workspace.test"),
					resource.TestCheckResourceAttr(
						"steampipecloud_workspace.test", "handle", workspaceHandle),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceUpdateHandleConfig(newWorkspaceHandle),
				Check: resource.TestCheckResourceAttr(
					"steampipecloud_workspace.test", "handle", newWorkspaceHandle),
			},
		},
	})
}

// configs
func testAccUserWorkspaceConfig(workspaceHandle string) string {
	return fmt.Sprintf(`
resource "steampipecloud_workspace" "test" {
	handle = "%s"
}`, workspaceHandle)
}

func testAccUserWorkspaceUpdateHandleConfig(newWorkspaceHandle string) string {
	return fmt.Sprintf(`
resource "steampipecloud_workspace" "test" {
	handle = "%s"
}`, newWorkspaceHandle)
}

// helper functions
func testAccCheckUserWorkspaceExists(resource string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		client := testAccProvider.Meta().(*SteampipeClient)

		// Get user handle
		userData, _, userErr := client.APIClient.Actors.Get(ctx).Execute()
		if userErr != nil {
			return fmt.Errorf("error fetching user handle. %s", userErr)
		}

		_, _, err := client.APIClient.UserWorkspaces.Get(ctx, userData.Handle, rs.Primary.ID).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckUserWorkspaceDestroy(s *terraform.State) error {
	ctx := context.Background()
	client := testAccProvider.Meta().(*SteampipeClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "steampipecloud_workspace" {
			// Get user handle
			userData, _, userErr := client.APIClient.Actors.Get(ctx).Execute()
			if userErr != nil {
				return fmt.Errorf("error fetching user handle. %s", userErr)
			}

			_, r, err := client.APIClient.UserWorkspaces.Get(ctx, userData.Handle, rs.Primary.ID).Execute()
			if err == nil {
				return fmt.Errorf("Workspace still exists")
			}

			if r.StatusCode != 404 {
				return fmt.Errorf("expected 'no content' error, got %s", err)
			}
		}
	}

	return nil
}
