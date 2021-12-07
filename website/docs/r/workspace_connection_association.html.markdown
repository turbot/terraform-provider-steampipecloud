---
layout: "steampipecloud"
title: "steampipecloud"
template: Documentation
page_title: "Steampipe Cloud: steampipecloud_workspace_connection_association"
nav:
  title: steampipecloud_workspace_connection_association
---

# steampipecloud_workspace_connection_association

The `Steampipe Cloud Connection` represents a set of connections that are currently attached to the workspace. This resource can be used multiple times with the same connection for non-overlapping workspaces.

## Example Usage

**Creating Your First Workspace Connection Association**

```hcl
resource "steampipecloud_workspace" "dev_workspace" {
  handle = "dev"
}

resource "steampipecloud_connection" "dev_conn" {
  handle = "devconn"
  plugin = "bitbucket"
}

resource "steampipecloud_workspace_connection_association" "test" {
  workspace_handle  = steampipecloud_workspace.dev_workspace.handle
  connection_handle = steampipecloud_connection.dev_conn.handle
}
```

**Creating a Organization Workspace Connection Association**

```hcl
provider "steampipecloud" {
  org = "testorg"
}

resource "steampipecloud_workspace" "org_dev_workspace" {
  handle = "dev"
}

resource "steampipecloud_connection" "org_dev_conn" {
  handle = "devconn"
  plugin = "bitbucket"
}

resource "steampipecloud_workspace_connection_association" "org_test" {
  workspace_handle  = steampipecloud_workspace.org_dev_workspace.handle
  connection_handle = steampipecloud_connection.org_dev_conn.handle
}
```

## Argument Reference

The following arguments are supported:

- `workspace_handle` - (Required) The handle of the workspace to add the connection to.
- `connection_handle` - (Required) The handle of the connection to add to workspace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `association_id` - An unique identifier of the workspace connection association.
- `connection_config` - A map of connection configuration.
- `connection_created_at` - The creation time of the connection.
- `connection_id` - An unique identifier of the connection.
- `connection_identity_id` - An unique identifier of the entity, where the connection is created.
- `connection_plugin` - The name of the plugin.
- `connection_type` - The type of the resource.
- `connection_updated_at` - The time when the connection was last updated.
- `connection_version_id` - The version of the connection.
- `workspace_created_at` - The creation time of the workspace.
- `workspace_database_name` - The name of the Steampipe workspace database.
- `workspace_hive` - The Steampipe workspace hive.
- `workspace_host` - The workspace hostname.
- `workspace_id` - An unique identifier of the workspace.
- `workspace_identity_id` - An unique identifier of the entity, where the workspace is created.
- `workspace_public_key` - The workspace public key.
- `workspace_state` - The current state of the workspace.
- `workspace_updated_at` - The time when the workspace was last updated.
- `workspace_version_id` - The workspace version.

## Import

Connections can be imported using the `id`. For example,

```sh
terraform import steampipecloud_workspace_connection_association.test_import dev/devconn
```
