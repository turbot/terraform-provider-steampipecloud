WEBSITE_REPO=github.com/hashicorp/terraform-website
FULL_PKG_NAME=github.com/turbot/terraform-provider-steampipecloud
VERSION_PLACEHOLDER=version.ProviderVersion
PKG_NAME=steampipecloud
VERSION=0.0.1
DIR=~/.terraform.d/plugins
TEST?=$$(go list ./... |grep -v 'vendor')
RUN=TestAccOrganizationMember_Basic

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

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -parallel 1 -count 1 -timeout 120m

testaccfocus: fmtcheck
	TF_ACC=1 go test $(TEST) -run $(RUN) -parallel 1 -count 1 -timeout 120m
