---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "steampipecloud Provider"
description: "Terraform provider for interacting with Steampipe Cloud API."
---

# Steampipe Cloud Provider

~> **Note:** The Steampipe Cloud provider has been deprecated. Please use the [Turbot Pipes provider](https://registry.terraform.io/providers/turbot/pipes) instead. This was part of our [renaming](https://turbot.com/blog/2023/07/introducing-turbot-guardrails-and-pipes) of Steampipe Cloud to Turbot Pipes.

<!-- Steampipe Cloud provides a hosted platform for Steampipe, simplifying setup and operation, accelerating integration, and providing solutions for collaborating and sharing insights. -->

The [Steampipe Cloud](https://cloud.steampipe.io/) provider is used to interact
with the resources supported by Steampipe Cloud.  The provider needs to be
configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
terraform {
  required_providers {
    steampipecloud = {
      source = "turbot/steampipecloud"
    }
  }
}

# Configure the Steampipe Cloud provider
provider "steampipecloud" {
  token = "spt_q9boaa6gutha5g3rgexample"
}

# Create a user workspace
resource "steampipecloud_workspace" "my_user_workspace" {
  # ...
}

# Create an organization workspace
resource "steampipecloud_workspace" "my_org_workspace" {
  organization = 'myorg'
  # ...
}
```

## Argument Reference

- **token** (Required) Token used to authenticate to Steampipe Cloud API. You can manage your API tokens from the Settings page for your user account in Steampipe Cloud. This can also be set via the `STEAMPIPE_CLOUD_TOKEN` environment variable.
- **host** (Optional) The Steampipe Cloud Host URL. This defaults to `https://cloud.steampipe.io/`. You only need to set this if you are connecting to a remote Steampipe Cloud database that is NOT hosted in `https://cloud.steampipe.io/`. This can also be set via the `STEAMPIPE_CLOUD_HOST` environment variable.
