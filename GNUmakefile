# Common tasks. `make check` is what CI runs; run it before every push.

BASE ?= origin/main
CHANGE ?=
APISYNC = go run ./cmd/apisync

default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w -e .

# Resource descriptors from the spec and generator/overrides, then provider
# documentation and example formatting.
generate:
	go run ./cmd/codegen
	cd tools; go generate ./...

# Validate every example against a fresh build; no credentials needed.
# CLI=tofu KINDS=resources,data-sources,ephemeral-resources for OpenTofu.
CLI ?= terraform
KINDS ?= resources,data-sources,ephemeral-resources,actions
examples-check:
	@mkdir -p build
	go build -o build/terraform-provider-telnyx .
	go run ./cmd/examplecheck -cli $(CLI) -provider build/terraform-provider-telnyx -kinds $(KINDS)

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

# Acceptance tests create real Telnyx objects. Needs TELNYX_API_KEY.
testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./internal/...

## Telnyx API sync ##########################################################

# Download the latest upstream spec if its commit changed.
upstream-fetch:
	$(APISYNC) fetch

# Rebuild upstream/telnyx/by-tag, catalog/resources.yaml and catalog/coverage.yaml.
catalog:
	$(APISYNC) catalog

# Fetch and rebuild in one step.
sync: upstream-fetch catalog

# Compare the spec pinned on $(BASE) with the working tree. With CHANGE set,
# writes openspec/changes/$(CHANGE)/api-diff.md; otherwise build/api-diff.md.
api-diff:
	@mkdir -p build
	git show $(BASE):upstream/telnyx/spec3.json > build/old-spec3.json
	git show $(BASE):upstream/telnyx/lock.json > build/old-lock.json
	$(APISYNC) diff -old build/old-spec3.json -old-lock build/old-lock.json \
		-out $(if $(CHANGE),openspec/changes/$(CHANGE)/api-diff.md,build/api-diff.md)

# Lock matches, every operation is classified, and generated views are current.
api-check:
	$(APISYNC) check
	$(APISYNC) catalog
	git diff --exit-code -- upstream catalog || \
		(echo "catalog is out of date: run make catalog and commit"; exit 1)

# Release gate: also requires every coverage entry to be reviewed.
api-check-release:
	$(APISYNC) check -require-reviewed

# All release archives, checksums and the registry manifest in dist/, without
# signing or publishing. Needs goreleaser v2.
release-snapshot:
	goreleaser release --snapshot --clean --skip=sign,publish

check: build test api-check
	openspec validate --all --strict

.PHONY: examples-check default build install lint fmt generate test testacc upstream-fetch catalog sync api-diff api-check api-check-release release-snapshot check
