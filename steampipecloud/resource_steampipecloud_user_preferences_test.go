package steampipecloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccUserPreferences_Basic(t *testing.T) {
	resourceName := "steampipecloud_user_preferences.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserPreferencesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPreferencesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"steampipecloud_user_preferences.test", "communication_community_updates", "enabled"),
					resource.TestCheckResourceAttr(
						"steampipecloud_user_preferences.test", "communication_product_updates", "enabled"),
					resource.TestCheckResourceAttr(
						"steampipecloud_user_preferences.test", "communication_tips_and_tricks", "enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPreferencesModifyCommunityUpdatePreference(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"steampipecloud_user_preferences.test", "communication_community_updates", "disabled"),
					resource.TestCheckResourceAttr(
						"steampipecloud_user_preferences.test", "communication_product_updates", "enabled"),
					resource.TestCheckResourceAttr(
						"steampipecloud_user_preferences.test", "communication_tips_and_tricks", "enabled"),
				),
			},
		},
	})
}

// configs
func testAccUserPreferencesConfig() string {
	return `
		resource "steampipecloud_user_preferences" "test" {
			communication_community_updates       = "enabled"
		}
	`
}

func testAccUserPreferencesModifyCommunityUpdatePreference() string {
	return `
		resource "steampipecloud_user_preferences" "test" {
			communication_community_updates       = "disabled"
		}
	`
}

func testAccCheckUserPreferencesDestroy(s *terraform.State) error {
	return nil
}
