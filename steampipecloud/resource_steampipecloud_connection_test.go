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

func TestAccConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_connection.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", "aws_conn_test"),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "access_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "secret_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-1"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccConnectionHandleUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("steampipecloud_connection.test", "handle", "aws_conn_update"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-2"),
					resource.TestCheckResourceAttr(resourceName, "regions.1", "us-east-1"),
				),
			},
		},
	})
}

func TestAccOrgConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_connection.test_org"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConnectionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionOrganizationExists("terraformtestorg"),
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", "aws_conn_test"),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "access_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "secret_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-1"),
				),
			},
			{
				Config: testAccOrgConnectionUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("steampipecloud_connection.test_org", "handle", "aws_conn_update"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-2"),
					resource.TestCheckResourceAttr(resourceName, "regions.1", "us-east-1"),
				),
			},
		},
	})
}

// configs
func testAccConnectionConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_connection" "test" {
	handle     = "aws_conn_test"
	plugin     = "aws"
	regions    = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}`
}

func testAccConnectionHandleUpdateConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_connection" "test" {
	handle     = "aws_conn_update"
	plugin     = "aws"
	regions    = ["us-east-2", "us-east-1"]
	access_key = "redacted"
  secret_key = "redacted"
}`
}

func testAccOrgConnectionConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle       = "terraformtestorg"
	display_name = "Terraform Test Org"
}

provider "steampipecloud" {
	alias = "turbie"
	org 	= steampipecloud_organization.test.handle
}

resource "steampipecloud_connection" "test_org" {
	provider   = steampipecloud.turbie
	handle     = "aws_conn_test"
	plugin     = "aws"
	regions    = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}`
}
func testAccOrgConnectionUpdateConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle       = "terraformtestorg"
	display_name = "Terraform Test Org"
}

provider "steampipecloud" {
	alias = "turbie"
	org 	= steampipecloud_organization.test.handle
}

resource "steampipecloud_connection" "test_org" {
	provider   = steampipecloud.turbie
	handle     = "aws_conn_update"
	plugin     = "aws"
	regions    = ["us-east-2", "us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}`
}

// testAccCheckConnectionDestroy verifies the connection has been destroyed
func testAccCheckConnectionDestroy(s *terraform.State) error {
	isUser := true
	var r *http.Response
	var err error
	var actorHandle, orgHandle string
	ctx := context.Background()
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)
	if client.Config != nil {
		if client.Config.Org != "" {
			orgHandle = client.Config.Org
			isUser = false
		}
	}

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_connection" {
			continue
		}

		// Retrieve connection by referencing it's state handle for API lookup
		connectionHandle := rs.Primary.Attributes["handle"]

		if isUser {
			actorHandle, r, err = getUserHandler(client)
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. getUserHandler Error: \n%v", r)
			}
			_, r, err = client.APIClient.UserConnectionsApi.GetUserConnection(ctx, actorHandle, connectionHandle).Execute()
		} else {
			_, r, err = client.APIClient.OrgConnectionsApi.GetOrgConnection(ctx, orgHandle, connectionHandle).Execute()
		}
		if err == nil {
			return fmt.Errorf("Connection %s still exists in organization %s", connectionHandle, client.Config.Org)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccOrgConnection_Basic testAccCheckConnectionExists %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckConnectionExists(n string) resource.TestCheckFunc {
	isUser := true
	var orgHandle, actorHandle string
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connectionHandle := rs.Primary.Attributes["handle"]

		client := testAccProvider.Meta().(*SteampipeClient)

		if client.Config != nil {
			if client.Config.Org != "" {
				orgHandle = client.Config.Org
				isUser = false
			}
		}

		var r *http.Response
		var err error

		log.Printf("[DEBUG -------]: \n	IS_USER: %t \n	orgHandle: %s", isUser, orgHandle)

		if isUser {
			actorHandle, r, err = getUserHandler(client)
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. getUserHandler Error:\n	status_code: %d\n	body: %v", r.StatusCode, r.Body)
			}
			_, r, err = client.APIClient.UserConnectionsApi.GetUserConnection(context.Background(), actorHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. \nGetUserConnection.error:\n	status_code: %d\n	body: %v", r.StatusCode, r.Body)
			}
		} else {
			_, r, err = client.APIClient.OrgConnectionsApi.GetOrgConnection(context.Background(), orgHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead.\n GetOrgConnection.error in organization %s:	\n	status_code: %d\n	body: %v", orgHandle, r.StatusCode, r.Body)
			}
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection %s in organization %s not found.\nstatus: %d \nerr: %v", connectionHandle, client.Config.Org, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccOrgConnection_Basic testAccCheckConnectionExists %v", err)
			return err
		}
		return nil
	}
}

func testAccCheckConnectionOrganizationExists(orgHandle string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)
		ctx := context.Background()
		var err error

		// check if organization  is created
		_, _, err = client.APIClient.OrgsApi.GetOrg(ctx, orgHandle).Execute()
		if err != nil {
			return fmt.Errorf("error fetching organization with handle %s. %s", orgHandle, err)
		}
		return nil
	}
}
