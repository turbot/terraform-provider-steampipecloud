package steampipecloud

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserDataSource_basic(t *testing.T) {
	dataSourceName := "data.steampipecloud_user.caller"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "user_id", regexp.MustCompile(`^u_[a-z0-9]{20}`)),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^u_[a-z0-9]{20}`)),
				),
			},
		},
	})
}

const testAccUserDataSourceConfig_empty = `
data "steampipecloud_user" "caller" {}
`
