package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
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
					resource.TestCheckResourceAttr(resourceName, "handle", "awstest"),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "access_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "secret_key", "redacted"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectionHandleUpdateConfig(),
				Check:  resource.TestCheckResourceAttr("steampipecloud_organization.test", "handle", "awstest"),
			},
			// {
			// 	Config: testAccOrganizationUpdateHandleConfig(),
			// 	Check: resource.TestCheckResourceAttr(
			// 		"steampipecloud_organization.test", "handle", "terraformtestorg"),
			// },
		},
	})
}

// configs
func testAccConnectionConfig() string {
	return `
provider "steampipecloud" {}

resource "steampipecloud_connection" "test" {
	handle = "awstest1"
	plugin = "aws"
	regions = ["us-east-1"]
	access_key = "redacted"
	secret_key = "redacted"
}`
}

func testAccConnectionHandleUpdateConfig() string {
	return `
resource "steampipecloud_organization" "test" {
	handle       = "terraformtest"
	display_name = "Terraform Test Org"
}
`
}

// func testAccConnectionUpdateHandleConfig() string {
// 	return `
// resource "steampipecloud_organization" "test" {
// 	handle       = "terraformtestorg"
// 	display_name = "Terraform Test Org"
// }
// `
// }

// helper functions

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
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. \nGetConnection.error:\n	status_code: %d\n	body: %v", r.StatusCode, r.Body)
			}
		} else {
			_, r, err = client.APIClient.OrgConnectionsApi.GetOrgConnection(ctx, orgHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead.\nGetConnection.error in organization %s:	\n	status_code: %d\n	body: %v", orgHandle, r.StatusCode, r.Body)
			}
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
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead. \nGetConnection.error:\n	status_code: %d\n	body: %v", r.StatusCode, r.Body)
			}
		} else {
			_, r, err = client.APIClient.OrgConnectionsApi.GetOrgConnection(context.Background(), orgHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("inside resourceSteampipeCloudConnectionRead.\nGetConnection.error in organization %s:	\n	status_code: %d\n	body: %v", orgHandle, r.StatusCode, r.Body)
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
		_, _, err := client.APIClient.OrgsApi.GetOrg(context.Background(), orgHandle).Execute()
		if err != nil {
			return fmt.Errorf("error fetching organization with handle %s. %s", orgHandle, err)
		}
		return nil
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
