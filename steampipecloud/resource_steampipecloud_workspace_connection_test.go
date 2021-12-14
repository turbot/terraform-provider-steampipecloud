package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// test suites

func TestAccWorkspaceConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_connection.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConnectionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestWorkspaceExists("testworkspaceconnection"),
					testAccCheckTestConnectionExists("aws_connection_test"),
					testAccCheckWorkspaceConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", "testworkspaceconnection"),
					resource.TestCheckResourceAttr(resourceName, "connection_handle", "aws_connection_test"),
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgWorkspaceConnectionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionOrganizationExists("terraformtestorganization"),
					testAccCheckTestWorkspaceExists("testworkspaceconnection"),
					testAccCheckTestConnectionExists("aws_connection_test"),
					testAccCheckWorkspaceConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", "testworkspaceconnection"),
					resource.TestCheckResourceAttr(resourceName, "connection_handle", "aws_connection_test"),
				),
			},
		},
	})
}

// User Workspace Connection association config
func testAccWorkspaceConnectionConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_workspace" "test" {
  handle = "testworkspaceconnection"
}

resource "steampipecloud_connection" "test" {
	handle     = "aws_connection_test"
	plugin     = "aws"
	regions    = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}

resource "steampipecloud_workspace_connection" "test" {
  workspace_handle  = steampipecloud_workspace.test.handle
  connection_handle = steampipecloud_connection.test.handle
}
`
}

// Organization Workspace Connection association config
func testAccOrgWorkspaceConnectionConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_organization" "test_org" {
	handle       = "terraformtestorganization"
	display_name = "Terraform Test Org"
}

provider "steampipecloud" {
	alias = "turbie"
	org 	= steampipecloud_organization.test_org.handle
}

resource "steampipecloud_workspace" "test_org" {
	provider = steampipecloud.turbie
  handle   = "testworkspaceconnection"
}

resource "steampipecloud_connection" "test_org" {
	provider   = steampipecloud.turbie
	handle     = "aws_connection_test"
	plugin     = "aws"
	regions    = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}

resource "steampipecloud_workspace_connection" "test_org" {
	provider          = steampipecloud.turbie
  workspace_handle  = steampipecloud_workspace.test_org.handle
  connection_handle = steampipecloud_connection.test_org.handle
}
`
}

// testAccCheckWorkspaceConnectionDestroy verifies the workspace connection association has been destroyed
func testAccCheckWorkspaceConnectionDestroy(s *terraform.State) error {
	isUser := true
	var r *http.Response
	var err error
	var actorHandle, orgHandle string

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)
	if client.Config != nil {
		if client.Config.Org != "" {
			orgHandle = client.Config.Org
			isUser = false
		}
	}

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace" && rs.Type != "steampipecloud_connection" && rs.Type != "steampipecloud_workspace_connection" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		connectionHandle := rs.Primary.Attributes["connection_handle"]
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]

		if isUser {
			actorHandle, _, err = getUserHandler(client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(context.Background(), actorHandle, workspaceHandle, connectionHandle).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(context.Background(), orgHandle, workspaceHandle, connectionHandle).Execute()
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
	return func(s *terraform.State) error {
		isUser := true
		var orgHandle, actorHandle string
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

		if client.Config != nil {
			if client.Config.Org != "" {
				orgHandle = client.Config.Org
				isUser = false
			}
		}
		var err error

		if isUser {
			actorHandle, _, err = getUserHandler(client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(context.Background(), actorHandle, workspaceHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("error reading user workspace connection: %s/%s.\nerr: %s", workspaceHandle, connectionHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(context.Background(), orgHandle, workspaceHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("error reading organization workspace connection: %s/%s.\nerr: %s", workspaceHandle, connectionHandle, err)
			}
		}

		return nil
	}
}

func testAccCheckTestWorkspaceExists(workspaceHandle string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		isUser := true
		var orgHandle, actorHandle string
		client := testAccProvider.Meta().(*SteampipeClient)

		if client.Config != nil {
			if client.Config.Org != "" {
				orgHandle = client.Config.Org
				isUser = false
			}
		}
		var err error

		if isUser {
			actorHandle, _, err = getUserHandler(client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserWorkspaces.Get(context.Background(), actorHandle, workspaceHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching user workspace with handle %s. %s", workspaceHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgWorkspaces.Get(context.Background(), orgHandle, workspaceHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching org workspace with handle %s. %s", workspaceHandle, err)
			}
		}
		return nil
	}
}

func testAccCheckTestConnectionExists(connHandle string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		isUser := true
		var orgHandle, actorHandle string
		client := testAccProvider.Meta().(*SteampipeClient)

		if client.Config != nil {
			if client.Config.Org != "" {
				orgHandle = client.Config.Org
				isUser = false
			}
		}
		var err error

		if isUser {
			actorHandle, _, err = getUserHandler(client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserConnections.Get(context.Background(), actorHandle, connHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching user connection with handle %s. %s", connHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgConnections.Get(context.Background(), orgHandle, connHandle).Execute()
			if err != nil {
				return fmt.Errorf("error fetching org connection with handle %s. %s", connHandle, err)
			}
		}
		return nil
	}
}
