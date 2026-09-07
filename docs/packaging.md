# Packaging and Releases

Garterscopic publishes one binary per supported platform and native packages from GitHub Actions.

## Release a Version

1. Make sure `go test ./...`, `go vet ./...`, and `mkdocs build --strict` pass on `main`.
2. Create and push an annotated tag:

   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. The `Release` workflow runs the CI workflow, builds binaries, creates a source archive, generates SHA-256 checksums, and creates the GitHub Release.
4. The distro package jobs build DEB, RPM, Arch, and Nix artifacts. A follow-up job attaches those artifacts and a separate package `SHA256SUMS` file to the same GitHub Release.

The workflow also publishes build provenance attestations for the release binaries. Package repositories are notified only after the release and package artifacts are available.

## GitHub Actions Artifacts

Pull-request and branch builds upload short-lived binary artifacts for verification. Release artifacts are retained by GitHub as release assets, so users should download from the release page rather than from the Actions run.

Release assets include:

- Linux, macOS, and FreeBSD binaries for amd64 and arm64
- A source archive
- DEB, RPM, Arch, and Nix package artifacts when their jobs succeed
- `SHA256SUMS` for binaries and source archive
- A second package `SHA256SUMS` for native package files

Verify a downloaded release asset with:

```bash
sha256sum -c SHA256SUMS
```

The checksum file must be in the same directory as the downloaded assets. On macOS, use `shasum -a 256 -c SHA256SUMS`.

## Package Jobs

The package workflow can be started manually from the Actions tab with a tag such as `v1.0.0`, or it can be called automatically by the release workflow. It downloads the Linux amd64 release binary, builds each package, validates package metadata, and uploads the result as a GitHub Actions artifact.

The release workflow then attaches those artifacts to the matching GitHub Release. This keeps package creation separate from source compilation while giving users one canonical download location.

## Required Repository Settings

For the complete release flow, configure:

- `GITHUB_TOKEN` permissions allowing contents write and artifact attestations
- `PACKAGING_DISPATCH_TOKEN` with permission to send repository dispatch events to the external packaging repositories
- `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` for documentation deployment

If external packaging repositories are not available yet, the dispatch step should be disabled or the secret should be configured before pushing a release tag.

## Local Package Checks

The package scripts expect a prebuilt binary named `garterscopic` in their output directory:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o garterscopic ./cmd/garterscopic

VERSION=0.1.0 OUTPUT=. bash packaging/deb/build.sh
VERSION=0.1.0 OUTPUT=. bash packaging/rpm/build.sh
```

The DEB and RPM commands require their platform package tooling (`dpkg-deb` or `rpmbuild`). Arch and Nix packages are best checked in their matching container or Nix environment, as the GitHub workflow does.
