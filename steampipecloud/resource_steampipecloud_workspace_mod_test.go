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
func TestAccUserWorkspaceMod_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_mod.aws_insights"
	workspaceHandle := "workspace" + randomString(3)
	modPath := "github.com/turbot/steampipe-mod-aws-insights"
	modAlias := "aws_insights"
	constraint := "*"
	newConstraint := ">v0.2.0"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceModDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceModConfig(workspaceHandle, modPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", constraint),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceModUpdateConfig(workspaceHandle, modPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", newConstraint),
				),
			},
		},
	})
}

func TestAccOrgWorkspaceMod_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_mod.aws_insights"
	orgHandle := "terraformtest"
	workspaceHandle := "workspace" + randomString(3)
	modPath := "github.com/turbot/steampipe-mod-aws-insights"
	modAlias := "aws_insights"
	constraint := "*"
	newConstraint := ">v0.2.0"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceModDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgWorkspaceModConfig(orgHandle, workspaceHandle, modPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "organization", orgHandle),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", constraint),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccOrgWorkspaceModUpdateConfig(orgHandle, workspaceHandle, modPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "organization", orgHandle),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", newConstraint),
				),
			},
		},
	})
}

func testAccUserWorkspaceModConfig(workspaceHandle, modPath string) string {
	return fmt.Sprintf(`
	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_mod" "aws_insights" {
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		path = "%s"
	}`, workspaceHandle, modPath)
}

func testAccUserWorkspaceModUpdateConfig(workspaceHandle, modPath string) string {
	return fmt.Sprintf(`
	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}

	resource "steampipecloud_workspace_mod" "aws_insights" {
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		path = "%s"
		constraint = ">v0.2.0"
	}`, workspaceHandle, modPath)
}

func testAccOrgWorkspaceModConfig(orgHandle, workspaceHandle, modPath string) string {
	return fmt.Sprintf(`
	resource "steampipecloud_organization" "test_org" {
		handle       = "%s"
		display_name = "Terraform Test"
	}

	resource "steampipecloud_workspace" "test_workspace" {
		organization = steampipecloud_organization.test_org.handle
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_mod" "aws_insights" {
		organization = steampipecloud_organization.test_org.handle
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		path = "%s"
	}`, orgHandle, workspaceHandle, modPath)
}

func testAccOrgWorkspaceModUpdateConfig(orgHandle, workspaceHandle, modPath string) string {
	return fmt.Sprintf(`
	resource "steampipecloud_organization" "test_org" {
		handle       = "%s"
		display_name = "Terraform Test"
	}

	resource "steampipecloud_workspace" "test_workspace" {
		organization = steampipecloud_organization.test_org.handle
		handle = "%s"
	}

	resource "steampipecloud_workspace_mod" "aws_insights" {
		organization = steampipecloud_organization.test_org.handle
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		path = "%s"
		constraint = ">v0.2.0"
	}`, orgHandle, workspaceHandle, modPath)
}

func testAccCheckWorkspaceModExists(workspaceHandle, modAlias string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "steampipecloud_workspace_mod" {
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
				_, _, err = client.APIClient.UserWorkspaceMods.Get(ctx, actorHandle, workspaceHandle, modAlias).Execute()
				if err != nil {
					return fmt.Errorf("error fetching mod %s in user workspace with handle %s. %s", modAlias, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceMods.Get(ctx, org, workspaceHandle, modAlias).Execute()
				if err != nil {
					return fmt.Errorf("error fetching mod %s in org workspace with handle %s. %s", modAlias, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceModDestroy verifies the mod has been destroyed in the workspace
func testAccCheckWorkspaceModDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace_mod" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]
		modAlias := rs.Primary.Attributes["alias"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceMods.Get(ctx, actorHandle, workspaceHandle, modAlias).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceMods.Get(ctx, org, workspaceHandle, modAlias).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Mod %s:%s still exists", workspaceHandle, modAlias)
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
