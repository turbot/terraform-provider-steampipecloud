package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccUserWorkspaceAggregator_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_aggregator.aggregator_1"
	workspaceHandle := "workspace" + randomString(3)
	aggregatorHandle := "aws_all"
	plugin := "aws"
	connections := `["aws1", "aws2"]`
	updatedAggregatorHandle := "aws_all_updated"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceAggregatorConfig(workspaceHandle, aggregatorHandle, plugin, connections),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceAggregatorExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "handle", aggregatorHandle),
					resource.TestCheckResourceAttr(resourceName, "plugin", plugin),
					TestJSONFieldEqual(t, resourceName, "connections", connections),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "connections"},
			},
			{
				Config: testAccUserWorkspaceAggregatorConfig(workspaceHandle, updatedAggregatorHandle, plugin, connections),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceAggregatorExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "handle", updatedAggregatorHandle),
					resource.TestCheckResourceAttr(resourceName, "plugin", plugin),
					TestJSONFieldEqual(t, resourceName, "connections", connections),
				),
			},
		},
	})
}

func testAccUserWorkspaceAggregatorConfig(workspaceHandle, aggregatorHandle, plugin string, connections string) string {
	return fmt.Sprintf(`
	provider "steampipecloud" {}

	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_aggregator" "aggregator_1" {
		workspace = steampipecloud_workspace.test_workspace.handle
		handle             = "%s"
		plugin             = "%s"
		connections        = jsonencode(%s)
	}`, workspaceHandle, aggregatorHandle, plugin, connections)
}

func testAccUserWorkspaceAggregatorUpdateConfig(workspaceHandle, aggregatorHandle, plugin string, connections string) string {
	return fmt.Sprintf(`
	provider "steampipecloud" {}

	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_aggregator" "aggregator_1" {
		workspace = steampipecloud_workspace.test_workspace.handle
		handle             = "%s"
		plugin             = "%s"
		connections        = jsonencode(%s)
	}`, workspaceHandle, aggregatorHandle, plugin, connections)
}

func testAccCheckWorkspaceAggregatorExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "steampipecloud_workspace_aggregator" {
				continue
			}

			aggregatorHandle := rs.Primary.Attributes["handle"]
			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var userHandle string
				userHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspaceAggregators.Get(ctx, userHandle, workspaceHandle, aggregatorHandle).Execute()
				if err != nil {
					return fmt.Errorf("error fetching aggregator %s in user workspace with handle %s. %s", aggregatorHandle, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceAggregators.Get(ctx, org, workspaceHandle, aggregatorHandle).Execute()
				if err != nil {
					return fmt.Errorf("error fetching aggregator %s in org workspace with handle %s. %s", aggregatorHandle, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceAggregatorDestroy verifies the aggregator has been deleted from the workspace
func testAccCheckWorkspaceAggregatorDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace_aggregator" {
			continue
		}

		// Retrieve workspace handle and aggregator handle by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace"]
		aggregatorHandle := rs.Primary.Attributes["handle"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceAggregators.Get(ctx, userHandle, workspaceHandle, aggregatorHandle).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceAggregators.Get(ctx, org, workspaceHandle, aggregatorHandle).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Aggregator %s/%s still exists", workspaceHandle, aggregatorHandle)
		}

		if isUser {
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspaceAggregatorDestroy testAccCheckUserWorkspaceAggregatorDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspaceAggregatorDestroy testAccCheckUserWorkspaceAggregatorDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
