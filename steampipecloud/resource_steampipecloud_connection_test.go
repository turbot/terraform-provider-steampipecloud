package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites

func TestAccConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_connection.test"
	connHandle := "aws_" + randomString(5)
	newHandle := "aws_" + randomString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig(connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", connHandle),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
					testCheckJSONString(resourceName, "config", `{"access_key":"redacted","regions":["us-east-1"],"secret_key":"redacted"}`),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccConnectionHandleUpdateConfig(newHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("steampipecloud_connection.test", "handle", newHandle),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-2\",\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
					testCheckJSONString(resourceName, "config", `{"access_key":"redacted","regions":[""us-east-2","us-east-1"],"secret_key":"redacted"}`),
				),
			},
		},
	})
}

func TestAccOrgConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_connection.test_org"
	orgHandle := "terraform" + randomString(9)
	connHandle := "aws_" + randomString(7)
	newHandle := "aws_" + randomString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConnectionConfig(connHandle, orgHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionOrganizationExists(orgHandle),
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", connHandle),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
				),
			},
			{
				Config: testAccOrgConnectionUpdateConfig(newHandle, orgHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("steampipecloud_connection.test_org", "handle", newHandle),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-2\",\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
				),
			},
		},
	})
}

// configs
func testAccConnectionConfig(connHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_connection" "test" {
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}`, connHandle)
}

func testAccConnectionHandleUpdateConfig(newHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_connection" "test" {
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-2", "us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}`, newHandle)
}

func testAccOrgConnectionConfig(connHandle string, orgHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "steampipecloud_connection" "test_org" {
	organization = steampipecloud_organization.test.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions      = ["us-east-1"]
		access_key   = "redacted"
		secret_key   = "redacted"
	})
}`, orgHandle, connHandle)
}

func testAccOrgConnectionUpdateConfig(newHandle string, orgHandle string) string {
	return fmt.Sprintf(`
provider "steampipecloud" {}

resource "steampipecloud_organization" "test" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "steampipecloud_connection" "test_org" {
	organization = steampipecloud_organization.test.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions      = ["us-east-2", "us-east-1"]
		access_key   = "redacted"
		secret_key   = "redacted"
	})
}`, orgHandle, newHandle)
}

// testAccCheckConnectionDestroy verifies the connection has been destroyed
func testAccCheckConnectionDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_connection" {
			continue
		}

		// Retrieve connection by referencing it's state handle for API lookup
		connectionHandle := rs.Primary.Attributes["handle"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var actorHandle string
			actorHandle, r, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. getUserHandler Error: \n%v", r)
			}
			_, r, err = client.APIClient.UserConnections.Get(ctx, actorHandle, connectionHandle).Execute()
		} else {
			_, r, err = client.APIClient.OrgConnections.Get(ctx, org, connectionHandle).Execute()
		}
		if err == nil {
			return fmt.Errorf("Connection %s still exists in organization %s", connectionHandle, org)
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
	ctx := context.Background()

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

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		var r *http.Response
		var err error

		if isUser {
			var actorHandle string
			actorHandle, r, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("testAccCheckConnectionExists. getUserHandler error: %v", decodeResponse(r))
			}
			_, r, err = client.APIClient.UserConnections.Get(context.Background(), actorHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("testAccCheckConnectionExists. Get user connection error: %v", decodeResponse(r))
			}
		} else {
			_, r, err = client.APIClient.OrgConnections.Get(context.Background(), org, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("testAccCheckConnectionExists.\n Get organization connection error: %v", decodeResponse(r))
			}
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection %s in organization %s not found.\nstatus: %d \nerr: %v", connectionHandle, org, r.StatusCode, r.Body)
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
		_, _, err = client.APIClient.Orgs.Get(ctx, orgHandle).Execute()
		if err != nil {
			return fmt.Errorf("error fetching organization with handle %s. %s", orgHandle, err)
		}
		return nil
	}
}

func testCheckJSONString(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, s.RootModule().Path)
		}

		emptyCheck := false
		if value == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%")) {
			emptyCheck = true
		}

		if v, ok := rs.Primary.Attributes[key]; !ok {
			configString, _ := formatJson(v)
			compareValue, _ := formatJson(value)
			if configString != compareValue {
				if emptyCheck && !ok {
					return nil
				}

				if !ok {
					return fmt.Errorf("%s: Attribute '%s' not found", name, key)
				}

				return fmt.Errorf(
					"%s: Attribute '%s' expected %#v, got %#v",
					name,
					key,
					value,
					v)

			}
		}
		return nil
	}
}
