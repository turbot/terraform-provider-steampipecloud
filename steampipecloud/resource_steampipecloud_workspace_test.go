package steampipecloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// test suites
func TestAccUserWorkspace_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserWorkspaceExists("steampipecloud_workspace.test"),
					resource.TestCheckResourceAttr(
						"steampipecloud_workspace.test", "handle", "terraformtest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserWorkspaceUpdateHandleConfig(),
				Check: resource.TestCheckResourceAttr(
					"steampipecloud_workspace.test", "handle", "terraformtestworkspace"),
			},
		},
	})
}

// configs
func testAccUserWorkspaceConfig() string {
	return `
resource "steampipecloud_workspace" "test" {
	handle = "terraformtest"
}
`
}

func testAccUserWorkspaceUpdateHandleConfig() string {
	return `
resource "steampipecloud_workspace" "test" {
	handle = "terraformtestworkspace"
}
`
}

// helper functions
func testAccCheckUserWorkspaceExists(resource string) resource.TestCheckFunc {
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
		userData, _, userErr := client.APIClient.Actors.Get(context.Background()).Execute()
		if userErr != nil {
			return fmt.Errorf("error fetching user handle. %s", userErr)
		}

		_, _, err := client.APIClient.UserWorkspaces.Get(context.Background(), userData.Handle, rs.Primary.ID).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckUserWorkspaceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*SteampipeClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "steampipecloud_workspace" {
			// Get user handle
			userData, _, userErr := client.APIClient.Actors.Get(context.Background()).Execute()
			if userErr != nil {
				return fmt.Errorf("error fetching user handle. %s", userErr)
			}

			_, r, err := client.APIClient.UserWorkspaces.Get(context.Background(), userData.Handle, rs.Primary.ID).Execute()
			if err == nil {
				return fmt.Errorf("alert still exists")
			}

			if r.StatusCode != 404 {
				return fmt.Errorf("expected 'no content' error, got %s", err)
			}
		}
	}

	return nil
}
