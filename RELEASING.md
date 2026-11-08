# Releasing

Releases are built by [GoReleaser](https://goreleaser.com) in
`.github/workflows/release.yml` when a `v*` tag is pushed. The workflow first runs
the release gate (`make api-check-release`), the unit tests and the generated-docs
check, and refuses a tag without a collated changelog section. It then builds
every platform, signs the SHA256 checksums with the project GPG key, attaches the
registry manifest and creates the GitHub release with the changelog section as
its notes. The registries read that GitHub release.

## One-time setup (maintainer)

These need the maintainer's own accounts and keys, so they are not automated.

1. **GPG signing key.** Create an RSA or DSA key (the Terraform Registry does not
   accept ECC keys), for example `gpg --full-generate-key` with RSA 4096. Export
   the private key with `gpg --armor --export-secret-keys <fingerprint>` and the
   public key with `gpg --armor --export <fingerprint>`.
2. **GitHub secrets** on `RomansWorks/terraform-provider-telnyx`
   (Settings → Secrets and variables → Actions):
   - `GPG_PRIVATE_KEY`: the armored private key.
   - `PASSPHRASE`: its passphrase.
   The workflow uses the built-in `GITHUB_TOKEN`; no other token is needed.
   `TELNYX_API_KEY` is only for the acceptance workflow, not for releases.
3. **Terraform Registry.** Sign in to <https://registry.terraform.io> with the
   GitHub account that owns the `RomansWorks` namespace, add the public GPG key
   under User Settings → Signing Keys, then Publish → Provider and pick
   `terraform-provider-telnyx`. The registry installs a webhook and picks up each
   new GitHub release.
4. **OpenTofu registry.** Open the "Submit new provider" issue at
   <https://github.com/opentofu/registry/issues/new/choose> for
   `romansworks/terraform-provider-telnyx`, and the "Submit new provider signing
   key" issue with the public GPG key. The OpenTofu registry picks up later
   releases automatically.

## Each release

1. Make sure `main` is green and every in-flight pull request for the release has
   merged.
2. Optionally check the build locally: `make release-snapshot` (needs GoReleaser
   v2; nothing is signed or uploaded, output goes to `dist/`).
3. Collate the changelog on a branch and merge it:

   ```sh
   changie batch <version>   # e.g. 0.2.0, or `changie batch auto`
   changie merge
   ```

   This writes `.changes/<version>.md` and the `## <version>` section in
   `CHANGELOG.md`.
4. Tag the merge commit and push the tag:

   ```sh
   git tag v<version>
   git push origin v<version>
   ```

5. Watch the Release workflow. If the gate fails, nothing is published: fix it on
   `main`, delete the tag (`git push --delete origin v<version>`) and tag again.

Versioning follows the `release` spec in `openspec/specs/release/spec.md`:
additions are minor releases, fixes are patches, and removals or renames wait for
the next major after a deprecation.
