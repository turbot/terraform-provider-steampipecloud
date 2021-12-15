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

func TestAccWorkspaceConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_connection.test_conn"
	workspaceHandle := "workspace" + randomString(6)
	connHandle := "aws_" + randomString(4)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConnectionConfig(workspaceHandle, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestWorkspaceExists(workspaceHandle),
					testAccCheckTestConnectionExists(connHandle),
					testAccCheckWorkspaceConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "connection_handle", connHandle),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
		},
	})
}

func TestAccOrgWorkspaceConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_connection.test_org"
	orgName := "terraform-" + randomString(11)
	workspaceHandle := "workspace" + randomString(5)
	connHandle := "aws_" + randomString(3)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgWorkspaceConnectionConfig(orgName, workspaceHandle, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionOrganizationExists(orgName),
					testAccCheckTestWorkspaceExists(workspaceHandle),
					testAccCheckTestConnectionExists(connHandle),
					testAccCheckWorkspaceConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "connection_handle", connHandle),
				),
			},
		},
	})
}

// User Workspace Connection association config
func testAccWorkspaceConnectionConfig(workspace string, conn string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_workspace" "test_conn" {
  handle = "%s"
}

resource "steampipecloud_connection" "test_conn" {
	handle     = "%s"
	plugin     = "aws"
	regions    = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}

resource "steampipecloud_workspace_connection" "test_conn" {
  workspace_handle  = steampipecloud_workspace.test_conn.handle
  connection_handle = steampipecloud_connection.test_conn.handle
}`, workspace, conn)
}

// Organization Workspace Connection association config
func testAccOrgWorkspaceConnectionConfig(org string, workspace string, conn string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

provider "steampipecloud" {
	alias = "turbie"
	organization 	= steampipecloud_organization.test_org.handle
}

resource "steampipecloud_workspace" "test_org" {
	provider = steampipecloud.turbie
  handle   = "%s"
}

resource "steampipecloud_connection" "test_org" {
	provider   = steampipecloud.turbie
	handle     = "%s"
	plugin     = "aws"
	regions    = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}

resource "steampipecloud_workspace_connection" "test_org" {
	provider          = steampipecloud.turbie
  workspace_handle  = steampipecloud_workspace.test_org.handle
  connection_handle = steampipecloud_connection.test_org.handle
}`, org, workspace, conn)
}

// testAccCheckWorkspaceConnectionDestroy verifies the workspace connection association has been destroyed
func testAccCheckWorkspaceConnectionDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)
	isUser, orgHandle := isUserConnection(client)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace" && rs.Type != "steampipecloud_connection" && rs.Type != "steampipecloud_workspace_connection" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		connectionHandle := rs.Primary.Attributes["connection_handle"]
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(ctx, orgHandle, workspaceHandle, connectionHandle).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Connection association %s/%s still exists", workspaceHandle, connectionHandle)
		}

		// If the error is equivalent to 404 not found, the workspace connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccWorkspaceConnection_Basic testAccCheckWorkspaceConnectionDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckWorkspaceConnectionExists(n string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		var err error

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connectionHandle := rs.Primary.Attributes["connection_handle"]
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]

		client := testAccProvider.Meta().(*SteampipeClient)
		isUser, orgHandle := isUserConnection(client)

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("error reading user workspace connection: %s/%s.\nerr: %s", workspaceHandle, connectionHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(ctx, orgHandle, workspaceHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("error reading organization workspace connection: %s/%s.\nerr: %s", workspaceHandle, connectionHandle, err)
			}
		}

		return nil
	}
}

func testAccCheckTestWorkspaceExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)
		isUser, orgHandle := isUserConnection(client)

		var err error
		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserWorkspaces.Get(ctx, actorHandle, workspaceHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching user workspace with handle %s. %s", workspaceHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgWorkspaces.Get(ctx, orgHandle, workspaceHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching org workspace with handle %s. %s", workspaceHandle, err)
			}
		}
		return nil
	}
}

func testAccCheckTestConnectionExists(connHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)

		isUser, orgHandle := isUserConnection(client)
		var err error

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserConnections.Get(ctx, actorHandle, connHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching user connection with handle %s. %s", connHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgConnections.Get(ctx, orgHandle, connHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching org connection with handle %s. %s", connHandle, err)
			}
		}
		return nil
	}
}
