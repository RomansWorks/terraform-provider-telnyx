# Terraform Provider for Telnyx

A community [Terraform](https://www.terraform.io) and [OpenTofu](https://opentofu.org)
provider for the [Telnyx](https://telnyx.com) API: numbers, voice, messaging,
wireless, networking, storage, and AI assistants, avatars and missions. It is not
an official Telnyx product.

> **Status:** 0.x. Every Telnyx API v2 operation is either implemented or
> recorded as out of scope in [`catalog/coverage.yaml`](catalog/coverage.yaml):
> about 160 resources, 430 data sources, 110 actions and 4 ephemeral resources.
> Releases before 1.0 may still change attribute names between minor versions;
> each change is listed in the [changelog](CHANGELOG.md). See [the docs](docs/).

## Requirements

- Terraform 1.11 or later (write-only `_wo` attributes and ephemeral resources);
  Terraform 1.14 or later for actions (`action` blocks).
- OpenTofu 1.11 or later for resources, data sources and ephemeral resources.
  OpenTofu does not run provider actions; see the
  [OpenTofu guide](docs/guides/opentofu.md).
- A Telnyx API v2 key.

## Usage

```hcl
terraform {
  required_providers {
    telnyx = {
      source = "romansworks/telnyx"
    }
  }
}

provider "telnyx" {
  # Or set TELNYX_API_KEY.
  api_key = var.telnyx_api_key
}
```

| Argument | Environment variable | Default |
|---|---|---|
| `api_key` | `TELNYX_API_KEY` | (required) |
| `base_url` | `TELNYX_BASE_URL` | `https://api.telnyx.com/v2` |
| `max_retries` | `TELNYX_MAX_RETRIES` | `4` |

Objects the Telnyx API cannot delete are removed from state with a warning on
destroy. Phone numbers, SIM cards and hosted numbers stay in your account on
destroy unless `release_on_destroy = true`.

Telnyx Storage buckets are managed with the AWS provider pointed at Telnyx's S3
endpoint; see the [storage buckets guide](docs/guides/storage-buckets.md).

## How the provider stays current

Telnyx publishes its API as OpenAPI in
[team-telnyx/openapi](https://github.com/team-telnyx/openapi) and updates it most
working days. This repository pins that spec by commit, and a daily workflow opens a
pull request when it changes. Each update becomes an
[OpenSpec](https://github.com/Fission-AI/OpenSpec) change that a coding agent
implements. `catalog/coverage.yaml` records how every API operation is handled, and
CI fails if any operation is unaccounted for. See [AGENTS.md](AGENTS.md) for the
full loop.

## Development

Requirements: Go (see `go.mod`), Terraform 1.13 or later, Node.js 20.19 or later for
the OpenSpec CLI (`npm install -g @fission-ai/openspec`).

```sh
make check      # build, unit tests, API coverage, OpenSpec validation
make testacc    # acceptance tests; needs TELNYX_API_KEY and creates real objects
make generate   # resource descriptors from the spec and overrides, then docs
```

To try a local build, add a [dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
for `romansworks/telnyx` pointing at `$(go env GOPATH)/bin` and run `make install`.

## Roadmap

1. Foundations: provider core, pinned spec, coverage and sync loop. Done.
2. Resource generator and engine features. Done.
3. Coverage rollout by product area. Done.
4. Hardening: 0.1.0 on the Terraform and OpenTofu registries, nightly acceptance
   runs against a real account, then 1.0.

Releasing is described in [RELEASING.md](RELEASING.md).

## License

[Apache License 2.0](LICENSE)
