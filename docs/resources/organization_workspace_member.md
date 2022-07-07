---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "steampipecloud_organization_workspace_member Resource - terraform-provider-steampipecloud"
subcategory: ""
description: |-
  The `Steampipe Cloud Organization Workspace Member` provides the members of a workspace belonging to an organization who can collaborate, run queries and snapshots.

  This resource allows you to add/remove users from a workspace of your organization. When applied, an invitation will be sent to the user to become part of the workspace. When destroyed, either the invitation will be cancelled or the user will be removed.
---

# Resource: steampipecloud_organization_workspace_member

Manages the membership of a workspace in an organization.

This resource allows you to add/remove users from a workspace of your 
organization. When applied, an invitation will be sent to the user to become 
part of the workspace. When destroyed, either the invitation will be 
cancelled or the user will be removed.

## Example Usage

**Invite a user using user handle**

```hcl
resource "steampipecloud_organization" "myorg" {
  handle       = "myorg"
  display_name = "Test Org"
}

resource "steampipecloud_workspace" "myworkspace" {
  organization = steampipecloud_organization.myorg.handle
  handle       = "myworkspace"
}

resource "steampipecloud_organization_member" "example" {
  organization     = steampipecloud_organization.myorg.handle
  workspace_handle = steampipecloud_workspace.myworkspace.handle
  user_handle      = "someuser"
  role             = "owner"
}
```

**Invite a user using an email address**

```hcl
resource "steampipecloud_organization" "myorg" {
  handle       = "myorg"
  display_name = "Test Org"
}

resource "steampipecloud_workspace" "myworkspace" {
  organization = steampipecloud_organization.myorg.handle
  handle       = "myworkspace"
}

resource "steampipecloud_organization_member" "example" {
  organization     = steampipecloud_organization.myorg.handle
  workspace_handle = steampipecloud_workspace.myworkspace.handle
  email            = "user@domain.com"
  role             = "owner"
}
```

## Argument Reference

The following arguments are supported:

- `organization` - (Required) The organization ID or handle to which the workspace belongs to.
- `workspace_handle` - (Required) The workspace handle to which the user will be invited to.
- `role` - (Required) The role of the user in the workspace of the organization. Must be one of `reader`, `admin` or `owner`.

~> **Note:** An member can be invited either using an email address or a user handle. Providing both at the same time will result in an error.

- `email` - (Optional) The email address of the user to add to the workspace.
- `user_handle` - (Optional) The handle of the user to add to the workspace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `organization_workspace_member_id` - A unique identifier of the organization workspace membership.
- `organization_id` - A unique identifier of the organization.
- `workspace_id` - A unique identifier of the workspace.
- `user_id` - A unique identifier of the user to add to the workspace.
- `display_name` - The display name of the user to add to the workspace.
- `status` - The current membership status. Can be either `invited`, or `accepted`.
- `created_at` - The time when the invitation has been sent.
- `updated_at` - The time when the membership was last updated.
- `created_by` - The handle of the user who sent the invitation.
- `updated_by` - The handle of the user who last updated the membership.
- `version_id` - The membership version.

## Import

Organization workspace memberships can be imported using an ID made up of `organization_handle:workspace_handle:user_handle`, e.g.,

```sh
terraform import steampipecloud_organization_workspace_member.example hashicorp:dev:someuser
```