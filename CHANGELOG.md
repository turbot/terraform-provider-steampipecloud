## 0.5.0 (July 20, 2022)

FEATURES:

* **New Resource:** `steampipecloud_organization_workspace_member`
* `resources/steampipecloud_connection`: Add fields - `created_at`, `updated_at`, `created_by`, `updated_by`, `version_id`
* `resources/steampipecloud_organization`: Add fields - `created_by`, `updated_by`
* `resources/steampipecloud_organization_member`: Add fields - `created_by`, `updated_by`, `scope`
* `resources/steampipecloud_organization_member`: Modify the way organization members are listed i.e. use the `List` call instead of `Invited` and `Accepted` calls that was used previously
* `resources/steampipecloud_workspace`: Add fields - `created_by`, `updated_by`
* `resources/steampipecloud_workspace_connection`: Add fields - `created_at`, `updated_at`, `created_by`, `updated_by`, `version_id`, `identity_id`
* `resources/steampipecloud_workspace_mod`: Add fields - `created_by`, `updated_by`, `version_id`

## 0.4.0 (March 31, 2022)

FEATURES:

* **New Resource:** `steampipecloud_workspace_mod`
* **New Resource:** `steampipecloud_workspace_mod_variable`

## 0.3.0 (March 4, 2022)

ENHANCEMENTS:

* `resources/steampipecloud_connection`: Plugin connections are now defined in a `config` property and specific schemas are not required for new connection types. ([#33](https://github.com/turbot/terraform-provider-steampipecloud/issues/33))

BUG FIXES:

* `resources/steampipecloud_connection`: Fix import for connections in an organization. ([#32](https://github.com/turbot/terraform-provider-steampipecloud/issues/32))
* `resources/steampipecloud_workspace`: Fix import for workspaces in an organization. ([#32](https://github.com/turbot/terraform-provider-steampipecloud/issues/32))
* `resources/steampipecloud_workspace_connection`: Fix import for workspace connections in an organization. ([#32](https://github.com/turbot/terraform-provider-steampipecloud/issues/32))

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
