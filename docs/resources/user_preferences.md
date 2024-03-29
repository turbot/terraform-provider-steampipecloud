---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "steampipecloud_user_preferences Resource - terraform-provider-steampipecloud"
subcategory: ""
description: |-
  The `Steampipe Cloud User Preferences` represents the preferences settings for a user.
---

# Resource: steampipecloud_user_preferences

Allows a user to manage various preferences related to their steampipe cloud profile e.g. email preferences.

## Example Usage

**Enable receiving all types of emails**

```hcl
resource "steampipecloud_user_preferences" "enable_all_emails" {
  communication_community_updates       = "enabled"
  communication_product_updates         = "enabled"
  communication_tips_and_tricks         = "enabled"
}
```

**Disable receiving community updates**

```hcl
resource "steampipecloud_user_preferences" "disable_community_updates" {
	communication_community_updates       = "disabled"
}
```

## Argument Reference

The following arguments are supported:

- `communication_community_updates` - (Optional) Whether the user is subscribed to receiving community update emails. Can either be `enabled` or `disabled`.
- `communication_product_updates` - (Optional) Whether the user is subscribed to receiving product update emails. Can either be `enabled` or `disabled`.
- `communication_tips_and_tricks` - (Optional) Whether the user is subscribed to receiving tips and tricks emails. Can either be `enabled` or `disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `created_at` - The ISO 8601 date & time the user preferences was created at.
- `updated_at` - The ISO 8601 date & time any of the user preferences was last updated at.
- `version_id` - The version ID of this user preferences.

## Import

### Import User Preferences

User workspace snapshots can be imported using an ID made up of `user_handle/preferences`, e.g.,

```sh
terraform import steampipecloud_user_preferences.example myhandle/preferences
```
