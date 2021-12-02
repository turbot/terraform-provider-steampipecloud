#!/bin/bash

BIN_DIR=$PWD/bin
OVERRIDES_FILENAME=$HOME/.terraformrc

cat << EOF > $OVERRIDES_FILENAME
provider_installation {
  dev_overrides {
    "registry.terraform.io/hashicorp/steampipecloud" = "$BIN_DIR"
  }
  direct {}
}
EOF

# provider_installation {
#   dev_overrides {
#     "turbot/steampipecloud" = "/Users/lalitbhardwaj/go/src/github.com/turbot/terraform-provider-steampipecloud/bin"
#   }
#   direct {}
# }