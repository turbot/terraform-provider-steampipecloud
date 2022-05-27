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

// Test case for user workspace only -
// Test case assumes that a user workspace already exists in the env of the handle dev
// TODO - Add the workspace creation and destruction logic as part of this test case.

// test suites
func TestAccUserWorkspaceModVariable_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_mod_variable.mandatory_tags"
	// workspaceHandle := "workspace" + randomString(3)
	workspaceHandle := "dev"
	modPath := "github.com/turbot/steampipe-mod-aws-tags"
	modAlias := "aws_tags"
	variableName := "mandatory_tags"
	defaultValue := fmt.Sprintf(`["Environment","Owner"]`)
	setting := fmt.Sprintf(`["Environment","Owner","Foo"]`)
	updatedSetting := fmt.Sprintf(`["Environment","Owner","Foo","Bar"]`)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceModVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceModVariableConfig(workspaceHandle, modPath, variableName, setting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceModVariableExists(workspaceHandle, modAlias, variableName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "mod_alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "name", variableName),
					resource.TestCheckResourceAttr(resourceName, "default_value", defaultValue),
					resource.TestCheckResourceAttr(resourceName, "setting_value", setting),
					resource.TestCheckResourceAttr(resourceName, "value", setting),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceModVariableUpdateConfig(workspaceHandle, modPath, variableName, updatedSetting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceModVariableExists(workspaceHandle, modAlias, variableName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "mod_alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "name", variableName),
					resource.TestCheckResourceAttr(resourceName, "default_value", defaultValue),
					resource.TestCheckResourceAttr(resourceName, "setting_value", updatedSetting),
					resource.TestCheckResourceAttr(resourceName, "value", updatedSetting),
				),
			},
		},
	})
}

func testAccUserWorkspaceModVariableConfig(workspaceHandle, modPath, variableName, setting string) string {
	return fmt.Sprintf(`
	resource "steampipecloud_workspace_mod" "aws_tags" {
		workspace_handle = "dev"
		path = "%s"
	}
	
	resource "steampipecloud_workspace_mod_variable" "mandatory_tags" {
		workspace_handle = "dev"
		mod_alias = steampipecloud_workspace_mod.aws_tags.alias
		name = "%s"
		setting_value = jsonencode(%s)
	}`, modPath, variableName, setting)
}

func testAccUserWorkspaceModVariableUpdateConfig(workspaceHandle, modPath, variableName, setting string) string {
	return fmt.Sprintf(`
	resource "steampipecloud_workspace_mod" "aws_tags" {
		workspace_handle = "dev"
		path = "%s"
	}
	
	resource "steampipecloud_workspace_mod_variable" "mandatory_tags" {
		workspace_handle = "dev"
		mod_alias = steampipecloud_workspace_mod.aws_tags.alias
		name = "%s"
		setting_value = jsonencode(%s)
	}`, modPath, variableName, setting)
}

func testAccCheckWorkspaceModVariableExists(workspaceHandle, modAlias, variableName string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "steampipecloud_workspace" {
				continue
			}

			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var actorHandle string
				actorHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspaceModVariables.Get(ctx, actorHandle, workspaceHandle, modAlias, variableName).Execute()
				if err != nil {
					return fmt.Errorf("error fetching variable %s in mod %s for user workspace with handle %s. %s", variableName, modAlias, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceModVariables.Get(ctx, org, workspaceHandle, modAlias, variableName).Execute()
				if err != nil {
					return fmt.Errorf("error fetching variable %s in mod %s for org workspace with handle %s. %s", variableName, modAlias, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceModVariableDestroy verifies the mod has been destroyed in the workspace
func testAccCheckWorkspaceModVariableDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace_mod_variable" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]
		modAlias := rs.Primary.Attributes["mod_alias"]
		variableName := rs.Primary.Attributes["name"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceModVariables.Get(ctx, actorHandle, workspaceHandle, modAlias, variableName).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceModVariables.Get(ctx, org, workspaceHandle, modAlias, variableName).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Mod Variable %s:%s:%s still exists", workspaceHandle, modAlias, variableName)
		}

		if isUser {
			// If the request is being made from the context of a user
			// the status code should be equivalent to 404 not found which means the workspace mod has been uninstalled
			// Otherwise return the error
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspaceModDestroy testAccCheckUserWorkspaceModDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			// If the request is made from the context of an organization, we will have status code as 403
			// i.e. Forbidden
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspaceModDestroy testAccCheckUserWorkspaceModDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
