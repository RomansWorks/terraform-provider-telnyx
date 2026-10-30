# Terraform Provider for Telnyx

A community [Terraform](https://www.terraform.io) and [OpenTofu](https://opentofu.org)
provider for the [Telnyx](https://telnyx.com) API: numbers, voice, messaging,
wireless, networking, storage, and AI assistants, avatars and missions. It is not
an official Telnyx product.

> **Status:** early. Seven pilot resources are available: AI assistants, avatars,
> tools and MCP servers, outbound voice profiles, phone number voice settings and
> number orders. See [the docs](docs/) and [the roadmap](#roadmap).

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

1. Foundations: provider core, pinned spec, coverage and sync loop.
2. Resource generator and a pilot set: AI assistant, AI avatar, AI tool, MCP server,
   outbound voice profile, phone number voice settings, number order (this release).
3. Coverage rollout by product area.
4. Hardening, nightly acceptance runs, and 1.0 on the Terraform and OpenTofu registries.

## License

[Apache License 2.0](LICENSE)
