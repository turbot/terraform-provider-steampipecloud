WEBSITE_REPO=github.com/hashicorp/terraform-website
FULL_PKG_NAME=github.com/turbot/terraform-provider-steampipecloud
VERSION_PLACEHOLDER=version.ProviderVersion
PKG_NAME=steampipecloud
VERSION=0.0.1
DIR=~/.terraform.d/plugins

default: build

build:
	go build -o bin/terraform-provider-$(PKG_NAME)_$(VERSION) -ldflags="-X $(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=$(VERSION)"
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

install: fmtcheck
	mkdir -vp $(DIR)
	go build -o bin/terraform-provider-$(PKG_NAME)_$(VERSION) -ldflags="-X $(FULL_PKG_NAME)/$(VERSION_PLACEHOLDER)=$(VERSION)"
	@sh -c "'$(CURDIR)/scripts/generate-dev-overrides.sh'"

uninstall:
	@rm -vf $(DIR)/terraform-provider-$(PKG_NAME)

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

