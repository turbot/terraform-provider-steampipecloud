---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "steampipecloud Provider"
subcategory: ""
description: "Terraform provider for interacting with SteampipeCloud API."
---

# SteampipeCloud Provider

<!-- Steampipe Cloud provides a hosted platform for Steampipe, simplifying setup and operation, accelerating integration, and providing solutions for collaborating and sharing insights. -->

The [Steampipe Cloud](https://cloud.steampipe.io/) provider is used to interact with the resources supported by Steampipe Cloud.
The provider needs to be configured with the proper credentials before it can
be used.

Use the navigation to the left to read about the available resources.
**Example Usage**

```hcl
# Configure the Steampipe Cloud provider
provider "steampipecloud" {
  token = "spt_example"
}

# Configure the Steampipe Cloud provider with `org`, to target the organization resources
provider "steampipecloud" {
  alias         = "test_dev"
  organization  = "testorg"
  token         = "spt_example"
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

## Schema

### Optional

- **token** (String, Required) Token used to authenticate to Steampipe Cloud API. This can also be set via the `STEAMPIPE_CLOUD_TOKEN` environment variable.
- **host** (String, Optional) The Steampipe Cloud Host URL. This default is to `https://cloud.steampipe.io/`, you only need to set this if you are connecting to a remote Steampipe Cloud database that is NOT hosted in `https://cloud.steampipe.io/`. This can also be set via the `STEAMPIPE_CLOUD_HOST` environment variable.
- **organization** (String, Optional) Steampipe Organizations, include multiple users and are intended for organizations to collaborate and share workspaces and connections.