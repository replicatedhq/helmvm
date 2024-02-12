VERSION ?= $(shell git describe --tags --dirty)
UNAME := $(shell uname)
ARCH := $(shell uname -m)
APP_NAME = embedded-cluster
ADMIN_CONSOLE_CHART_URL = oci://registry.replicated.com/library
ADMIN_CONSOLE_CHART_NAME = admin-console
ADMIN_CONSOLE_CHART_VERSION = 1.107.3
ADMIN_CONSOLE_IMAGE_OVERRIDE =
ADMIN_CONSOLE_MIGRATIONS_IMAGE_OVERRIDE =
EMBEDDED_OPERATOR_CHART_URL = oci://registry.replicated.com/library
EMBEDDED_OPERATOR_CHART_NAME = embedded-cluster-operator
EMBEDDED_OPERATOR_CHART_VERSION = 0.22.5
OPENEBS_CHART_URL = https://openebs.github.io/charts
OPENEBS_CHART_NAME = openebs/openebs
OPENEBS_CHART_VERSION = 3.10.0
KUBECTL_VERSION = v1.29.1
K0S_VERSION = v1.29.1+k0s.1
K0S_BINARY_SOURCE_OVERRIDE =
TROUBLESHOOT_VERSION = v0.83.0
LD_FLAGS = -X github.com/replicatedhq/embedded-cluster/pkg/defaults.K0sVersion=$(K0S_VERSION) \
	-X github.com/replicatedhq/embedded-cluster/pkg/defaults.Version=$(VERSION) \
	-X github.com/replicatedhq/embedded-cluster/pkg/defaults.K0sBinaryURL=$(K0S_BINARY_SOURCE_OVERRIDE) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/adminconsole.ChartURL=$(ADMIN_CONSOLE_CHART_URL) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/adminconsole.ChartName=$(ADMIN_CONSOLE_CHART_NAME) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/adminconsole.Version=$(ADMIN_CONSOLE_CHART_VERSION) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/adminconsole.ImageOverride=$(ADMIN_CONSOLE_IMAGE_OVERRIDE) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/adminconsole.MigrationsImageOverride=$(ADMIN_CONSOLE_MIGRATIONS_IMAGE_OVERRIDE) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/embeddedclusteroperator.ChartURL=$(EMBEDDED_OPERATOR_CHART_URL) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/embeddedclusteroperator.ChartName=$(EMBEDDED_OPERATOR_CHART_NAME) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/embeddedclusteroperator.Version=$(EMBEDDED_OPERATOR_CHART_VERSION) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/openebs.ChartURL=$(OPENEBS_CHART_URL) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/openebs.ChartName=$(OPENEBS_CHART_NAME) \
	-X github.com/replicatedhq/embedded-cluster/pkg/addons/openebs.Version=$(OPENEBS_CHART_VERSION)

.DEFAULT_GOAL := default
default: embedded-cluster

pkg/goods/bins/k0s: Makefile
	mkdir -p pkg/goods/bins
	if [ "$(K0S_BINARY_SOURCE_OVERRIDE)" != "" ]; then \
	    curl -L -o pkg/goods/bins/k0s "$(K0S_BINARY_SOURCE_OVERRIDE)" ; \
	else \
	    curl -L -o pkg/goods/bins/k0s "https://github.com/k0sproject/k0s/releases/download/$(K0S_VERSION)/k0s-$(K0S_VERSION)-amd64" ; \
	fi
	chmod +x pkg/goods/bins/k0s
	touch pkg/goods/bins/k0s

pkg/goods/bins/kubectl: Makefile
	mkdir -p pkg/goods/bins
	curl -L -o pkg/goods/bins/kubectl "https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/linux/amd64/kubectl"
	chmod +x pkg/goods/bins/kubectl
	touch pkg/goods/bins/kubectl

pkg/goods/bins/kubectl-support_bundle: Makefile
	mkdir -p pkg/goods/bins
	mkdir -p output/tmp/support-bundle
	curl -L -o output/tmp/support-bundle/support-bundle.tar.gz https://github.com/replicatedhq/troubleshoot/releases/download/$(TROUBLESHOOT_VERSION)/support-bundle_linux_amd64.tar.gz
	tar -xzf output/tmp/support-bundle/support-bundle.tar.gz -C output/tmp/support-bundle
	mv output/tmp/support-bundle/support-bundle pkg/goods/bins/kubectl-support_bundle
	touch pkg/goods/bins/kubectl-support_bundle

pkg/goods/bins/kubectl-preflight: Makefile
	mkdir -p pkg/goods/bins
	mkdir -p output/tmp/preflight
	curl -L -o output/tmp/preflight/preflight.tar.gz https://github.com/replicatedhq/troubleshoot/releases/download/$(TROUBLESHOOT_VERSION)/preflight_linux_amd64.tar.gz
	tar -xzf output/tmp/preflight/preflight.tar.gz -C output/tmp/preflight
	mv output/tmp/preflight/preflight pkg/goods/bins/kubectl-preflight
	touch pkg/goods/bins/kubectl-preflight

output/tmp/release-onmerge.tar.gz: e2e/kots-release-onmerge/*
	mkdir -p output/tmp
	tar -czf output/tmp/release-onmerge.tar.gz -C e2e/kots-release-onmerge .

output/tmp/release-onpr.tar.gz: e2e/kots-release-onpr/*
	mkdir -p output/tmp
	tar -czf output/tmp/release-onpr.tar.gz -C e2e/kots-release-onpr .

output/bin/embedded-cluster-release-builder:
	mkdir -p output/bin
	go build -o output/bin/embedded-cluster-release-builder e2e/embedded-cluster-release-builder/main.go

.PHONY: embedded-release-onmerge
embedded-release-onmerge: embedded-cluster output/tmp/release-onmerge.tar.gz output/bin/embedded-cluster-release-builder
	./output/bin/embedded-cluster-release-builder output/bin/embedded-cluster output/tmp/release-onmerge.tar.gz output/bin/embedded-cluster

.PHONY: embedded-release-onpr
embedded-release-onpr: embedded-cluster output/tmp/release-onpr.tar.gz output/bin/embedded-cluster-release-builder
	./output/bin/embedded-cluster-release-builder output/bin/embedded-cluster output/tmp/release-onpr.tar.gz output/bin/embedded-cluster

.PHONY: static
static: pkg/goods/bins/k0s \
	pkg/goods/bins/kubectl-preflight \
	pkg/goods/bins/kubectl \
	pkg/goods/bins/kubectl-support_bundle
	
.PHONY: embedded-cluster
embedded-cluster: static
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o ./output/bin/$(APP_NAME) ./cmd/embedded-cluster

.PHONY: unit-tests
unit-tests:
	go test -v $(shell go list ./... | grep -v /e2e)

.PHONY: vet
vet: static
	go vet ./...

.PHONY: e2e-tests
e2e-tests: embedded-release-onmerge
	go test -timeout 45m -parallel 1 -failfast -v ./e2e

.PHONY: e2e-test-onpr
e2e-test-onpr:
	go test -timeout 45m -v ./e2e -run $(TEST_NAME)$

.PHONY: e2e-test-onmerge
e2e-test-onmerge:
	go test -timeout 45m -v ./e2e -run $(TEST_NAME)$

.PHONY: clean
clean:
	rm -rf output
	rm -rf pkg/goods/bins

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: lint-and-fix
lint-and-fix:
	golangci-lint run --fix -c .golangci.yml ./...

.PHONY: scan
scan:
	trivy fs \
		--scanners vuln \
		--exit-code=1 \
		--severity="HIGH,CRITICAL" \
		--ignore-unfixed \
		./
