## 0.2.0 (December 17, 2021)

ENHANCEMENTS:

* `resources/steampipecloud_connection`: Add support for `turbot` plugin. ([#26](https://github.com/turbot/terraform-provider-steampipecloud/issues/26))

BUG FIXES:

* `resources/steampipecloud_workspace_connection`: Fix resource ID format when creating and deleting resources. ([#24](https://github.com/turbot/terraform-provider-steampipecloud/issues/24))

DOCUMENTATION:

* Update example usage in index doc to initialize plugin from `turbot/steampipecloud` instead of `hashicorp/steampipecloud`. ([#29](https://github.com/turbot/terraform-provider-steampipecloud/issues/29))

## 0.1.0 (December 16, 2021)

FEATURES:

* **New Resource:** `steampipecloud_connection`
* **New Resource:** `steampipecloud_organization`
* **New Resource:** `steampipecloud_organization_member`
* **New Resource:** `steampipecloud_workspace`
* **New Resource:** `steampipecloud_workspace_connection`
* **New Data Source:** `steampipecloud_organization`
* **New Data Source:** `steampipecloud_user`
