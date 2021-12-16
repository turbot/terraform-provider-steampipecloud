---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "steampipecloud_organization_member Resource - terraform-provider-steampipecloud"
subcategory: ""
description: |-
  The `Steampipe Cloud Organization Member` provides the members of an organization who can collaborate and share workspaces and connections.

  This resource allows you to add/remove users from your organization. When applied, an invitation will be sent to the user to become part of the organization. When destroyed, either the invitation will be cancelled or the user will be removed.
---

# Resource: steampipecloud_organization_member

Manages an organization membership.

This resource allows you to add/remove users from your organization. When
applied, an invitation will be sent to the user to become part of the
organization. When destroyed, either the invitation will be cancelled or the
user will be removed.

## Example Usage

**Invite a user using user handle**

```hcl
resource "steampipecloud_organization_member" "example" {
  organization = "myorg"
  user_handle  = "someuser"
  role         = "member"
}
```

**Invite a user using an email address**

```hcl
resource "steampipecloud_organization_member" "example" {
  organization = "myorg"
  email        = "user@domain.com"
  role         = "member"
}
```

## Argument Reference

The following arguments are supported:

- `organization` - (Required) The organization ID or handle to invite the user to.
- `role` - (Required) The role of the user within the organization. Must be one of `member` or `owner`.

~> **Note:** An member can be invited either using an email address or a user handle. Providing both at the same time will result in an error.

- `email` - (Optional) The email address of the user to add to the organization.
- `user_handle` - (Optional) The handle of the user to add to the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `created_at` - The time when the invitation has been sent.
- `display_name` - The display name of the user to add to the organization.
- `organization_id` - An unique identifier of the organization.
- `organization_member_id` - An unique identifier of the organization membership.
- `status` - The current membership status. Can be either `invited`, or `accepted`.
- `updated_at` - The time when the membership was last updated.
- `user_id` - An unique identifier of the user to add to the organization.
- `version_id` - The membership version.

## Import

Organization memberships can be imported using an ID made up of `organization_handle:user_handle`, e.g.,

```sh
terraform import steampipecloud_organization_member.example hashicorp:someuser
```