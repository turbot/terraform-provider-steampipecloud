---
layout: "steampipecloud"
title: Provider
template: Documentation
nav:
  order: 20
---

# Provider: steampipecloud

The Steampipe Cloud provider is used to interact with the resources supported by Steampipe Cloud.
The provider needs to be configured with the proper credentials before it can
be used.

## Authentication

The Steampipe provider offers a flexible means of providing credentials for authentication. The following methods are supported:

- Static credentials
- Environment variables

### Static Credentials

Static credentials can be provided by adding `token` arguments in-line in the Steampipe Cloud provider block. This information must be present in your configuration file.

**Example Usage**

```hcl
# Configure the Steampipe Cloud provider
provider "steampipecloud" {
  token = "spt_example"
}

# Configure the Steampipe Cloud provider with `org`, to target the organization resources
provider "steampipecloud" {
  alias = "test_dev"
  org   = "testorg"
  token = "spt_example"
}

# Create a user workspace
resource "steampipecloud_workspace" "my_user_workspace" {
  # ...
}

# Create an organization workspace
resource "steampipecloud_workspace" "my_org_workspace" {
  provider = steampipecloud.test_dev
  # ...
}
```

### Environment Variables

You can provide your credentials via `STEAMPIPE_CLOUD_TOKEN` environment variable, representing your Steampipe Cloud API token.

**Example Usage**

```shell
export STEAMPIPE_CLOUD_TOKEN=spt_xxxxxxxxxxxxxxxxxxxx_xxxxxxxxxxxxxxxxxxxxxxxxx
```

## Argument Reference

The following arguments are used:

- `token` - Steampipe Cloud API token. May also be set via the `STEAMPIPE_CLOUD_TOKEN` environment variable.
- `org` - The handle of the Steampipe Cloud organization.
- `hostname` - The hostname.
- `insecure_skip_verify` - If false, SSL verification will be skipped. Default value is false.
