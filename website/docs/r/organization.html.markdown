---
layout: "steampipecloud"
title: "steampipecloud"
template: Documentation
page_title: "Steampipe Cloud: steampipecloud_organization"
nav:
  title: steampipecloud_organization
---

# steampipecloud_organization

The `Steampipe Cloud Organization` include multiple users and are intended for organizations to collaborate and share workspaces and connections.

## Example Usage

**Creating Your First Organization**

```hcl
resource "steampipecloud_organization" "test_org" {
  handle       = "testorg"
  display_name = "Test Org"
}
```

## Argument Reference

The following arguments are supported:

- `avatar_url` - (Optional) A publicly accessible URL for the organization's logo.
- `display_name` - (Optional) A friendly name for your organization.
- `handle` - (Required) A friendly identifier for your workspace, and must be unique across your workspaces.
- `url` - (Optional) A publicly accessible URL for the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `created_at` - The creation time of the organization.
- `organization_id` - An unique identifier of the organization.
- `updated_at` - The time when the organization was last updated.
- `version_id` - The organization version.

## Import

Workspaces can be imported using the `handle`. For example,

```sh
terraform import steampipecloud_organization.test_org testorg
```
