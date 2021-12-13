package steampipecloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// test suites

func TestAccOrgConnection_Basic(t *testing.T) {
	resourceName := "steampipecloud_connection.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConnectionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionOrganizationExists("netaji"),
					testAccCheckOrgConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", "awstest"),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "access_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "secret_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-1"),
				),
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
			// {
			// 	Config: testAccOrganizationUpdateDisplayNameConfig(),
			// 	Check: resource.TestCheckResourceAttr(
			// 		"steampipecloud_organization.test", "display_name", "Terraform Test Org"),
			// },
			// {
			// 	Config: testAccOrganizationUpdateHandleConfig(),
			// 	Check: resource.TestCheckResourceAttr(
			// 		"steampipecloud_organization.test", "handle", "terraformtestorg"),
			// },
		},
	})
}

// configs
func testAccOrgConnectionConfig() string {
	return `
provider "steampipecloud" {
  org   = "netaji"
}

resource "steampipecloud_connection" "test" {
	handle = "awstest"
	plugin = "aws"
	regions = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}`
}

func testAccConnectionUpdateDisplayNameConfig() string {
	return `
resource "steampipecloud_organization" "test" {
	handle       = "terraformtest"
	display_name = "Terraform Test Org"
}
`
}

func testAccConnectionUpdateHandleConfig() string {
	return `
resource "steampipecloud_organization" "test" {
	handle       = "terraformtestorg"
	display_name = "Terraform Test Org"
}
`
}

// helper functions

// testAccCheckConnectionDestroy verifies the connection has been destroyed
func testAccCheckOrgConnectionDestroy(s *terraform.State) error {
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

		// Check
		_, r, err := client.APIClient.OrgConnectionsApi.GetOrgConnection(ctx, client.Config.Org, connectionHandle).Execute()
		if err == nil {
			return fmt.Errorf("Connection %s still exists in organization %s", connectionHandle, client.Config.Org)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}
	}

	return nil
}

func testAccEndpointsConfig(endpoints string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  endpoints {
    %[1]s
  }
}
`, endpoints))
}

func ConfigCompose(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}

const testAccProviderConfigBase = `
data "aws_partition" "provider_test" {}
# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`

func testAccCheckOrgConnectionExists(n string) resource.TestCheckFunc {
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
		_, r, err := client.APIClient.OrgConnectionsApi.GetOrgConnection(context.Background(), client.Config.Org, connectionHandle).Execute()

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection %s in organization %s not found.\nstatus: %d \nerr: %v", connectionHandle, client.Config.Org, r.StatusCode, r.Body)
			}
			return err
		}
		return nil
	}
}

func testAccCheckConnectionOrganizationExists(orgHandle string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)
		_, _, err := client.APIClient.OrgsApi.GetOrg(context.Background(), orgHandle).Execute()
		if err != nil {
			return fmt.Errorf("error fetching organization with handle %s. %s", orgHandle, err)
		}
		return nil
	}
}
