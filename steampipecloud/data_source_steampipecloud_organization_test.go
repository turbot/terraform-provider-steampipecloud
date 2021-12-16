package steampipecloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccOrganizationDataSource_basic(t *testing.T) {
	dataSourceName := "data.steampipecloud_organization.org_aaa"
	orgHandle := "terraform-" + randomString(14)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig(orgHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "organization_id", regexp.MustCompile(`^o_[a-z0-9]{20}`)),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^o_[a-z0-9]{20}`)),
					resource.TestCheckResourceAttr(dataSourceName, "handle", orgHandle),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", "Terraform Test Org Data Source"),
				),
			},
		},
	})
}

func testAccOrganizationDataSourceConfig(orgHandle string) string {
	return fmt.Sprintf(`
resource "steampipecloud_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test Org Data Source"
}

data "steampipecloud_organization" "org_aaa" {
	handle = "${steampipecloud_organization.test_org.handle}"
}`, orgHandle)
}
