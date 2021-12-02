---
layout: "steampipecloud"
title: "steampipecloud"
template: Documentation
page_title: "Steampipe Cloud: steampipecloud_connection"
nav:
  title: steampipecloud_connection
---

# steampipecloud_connection

The `Steampipe Cloud Connection` represents a set of tables for a single data source. Each connection is represented as a distinct Postgres schema. In order to query data, you'll need at least one connection.

Connections are defined at the user account or organization level, and they can be shared by multiple workspaces within the account or organization.

## Example Usage

**Creating Your First Connection**

```hcl
resource "steampipecloud_connection" "test" {
  plugin = "aws"
  handle = "test"
}
```

**Creating a Organization Connection**

```hcl
provider "steampipecloud" {
  org = "testorg"
}

resource "steampipecloud_connection" "test_org_connection" {
  resource "steampipecloud_connection" "test" {
    plugin = "aws"
    handle = "test"
    config   = <<EOT
      {
        "regions": [ "us-east-1" ],
        "access_key": "redacted",
        "secret_key": "redacted"
      }
    EOT
  }
}
```

## Argument Reference

The following arguments are supported:

- `handle` - (Required) A friendly identifier for your connection, and must be unique across your connections.
- `plugin` - (Required) The name of the plugin.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `connection_id` - An unique identifier of the connection.
- `identity_id` - A unique identifier of the entity, where the connection is created.
- `type` - The type of the resource.

## Import

Connections can be imported using the `handle`. For example,

```sh
terraform import steampipecloud_connection.test test
```
