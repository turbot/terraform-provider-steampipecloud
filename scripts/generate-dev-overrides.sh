#!/bin/bash

BIN_DIR=$PWD/bin
OVERRIDES_FILENAME=$HOME/.terraformrc

cat << EOF > $OVERRIDES_FILENAME
provider_installation {
  dev_overrides {
    "registry.terraform.io/turbot/steampipecloud" = "$BIN_DIR"
  }
  direct {}
}
EOF

