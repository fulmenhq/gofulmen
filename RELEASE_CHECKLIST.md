# Release Checklist (gofulmen)

gofulmen is a pure library module: releases are primarily **signed git tags** (`vX.Y.Z`) consumed via the Go
module proxy/VCS. We do not ship binaries.

This checklist is expected to become a shared ecosystem pattern (gofulmen/pyfulmen/tsfulmen/rsfulmen) once it’s
validated here and then codified into Crucible.

## Variables (Quick Reference)

- `GOFULMEN_RELEASE_TAG`: optional override tag (recommended for manual release commands; e.g. `v0.1.30`)
  - If unset, scripts default to `v$(cat VERSION)`.
- `GOFULMEN_GPG_HOMEDIR`: required dedicated Infosec signing keyring directory (separate from personal `~/.gnupg`)
- `GOFULMEN_PGP_KEY_ID`: defaults to, and if set must be, `DAB70DD758911B26BB45A08C3B75AC591449DEBE` (the pinned Infosec signing subkey)
- `GOFULMEN_TAGGER_NAME`: fixed as `FulmenHQ Infosec` by `make release-tag`
- `GOFULMEN_TAGGER_EMAIL`: fixed as `infosec@3leaps.net` by `make release-tag`; this email must be verified on the GitHub account holding the Infosec public key
- `GOFULMEN_MINISIGN_KEY`: optional minisign secret key path (creates a sidecar signature for the tag attestation)
- `GOFULMEN_MINISIGN_PUB`: optional minisign public key path (verifies the sidecar signature)
- `GOFULMEN_ALLOW_NON_MAIN=1`: optional override to tag from a non-`main` branch (not recommended)
- `GOFULMEN_REQUIRE_TAG=1`: enforce “must be on an exact tag” (intended for CI guard usage)

Note: `GOFULMEN_RELEASE_TAG` is not a secret and typically isn’t stored in encrypted env bundles.

## Pre-Release

- [ ] Optional: start from a clean slate (removes `dist/release`):
  ```bash
  make release-clean
  ```
- [ ] `git status` is clean
- [ ] `make sync` completed and provenance reviewed:
  - [ ] `.goneat/ssot/provenance.json` is present/current
  - [ ] `.crucible/metadata/metadata.yaml` is present/current
  - [ ] Run `make release-provenance-check`
- [ ] Quality gates pass: `make check-all`
- [ ] `CHANGELOG.md` updated (Unreleased → new section)
- [ ] `docs/releases/vX.Y.Z.md` created/updated
- [ ] `RELEASE_NOTES.md` updated (keep only latest 3 entries)
- [ ] `VERSION` matches the intended tag (`v$(cat VERSION)`)
- [ ] Guard: ensure tag/version match:
  ```bash
  make release-guard-tag-version
  ```

## Tagging (Signed Tag Required)

- [ ] Confirm the Infosec primary fingerprint `0CACA49B3119B6BC12B2CA11B9B485F294B9FE07` and signing subkey `DAB70DD758911B26BB45A08C3B75AC591449DEBE` are available in `GOFULMEN_GPG_HOMEDIR`. A key rotation requires updating these pinned release checks before it is used.
- [ ] Confirm `infosec@3leaps.net` is a verified email on the GitHub account that has the matching Infosec public key.
- [ ] Ensure interactive GPG signing can prompt for passphrase (recommended):
  ```bash
  export GPG_TTY="$(tty)"
  gpg-connect-agent updatestartuptty /bye
  ```
- [ ] Sanity checks (CI-friendly):
  ```bash
  make release-guard-tag-version
  make release-provenance-check
  ```
- [ ] Create the signed tag (this is _release process_ surface, not app code surface):
  - Signs an annotated git tag for the Go module version.
  - Does not upload release assets by default.
  - Does not publish public keys (GitHub verifies signatures using public keys attached to the signer account).
  ```bash
  make release-tag
  ```
- [ ] Verify the signed tag locally:
  ```bash
  make release-verify-tag
  # or:
  git tag -v v$(cat VERSION)
  ```
- [ ] Push:
  ```bash
  git push origin main
  git push origin v$(cat VERSION)
  make release-verify-remote-tag
  ```

## Post-Release

- [ ] Spot-check downstream consumption:
  ```bash
  go list -m github.com/fulmenhq/gofulmen@v$(cat VERSION)
  ```
- [ ] Optional: remove local release artifacts:
  ```bash
  make release-clean
  ```
- [ ] Optional: show consumers how to verify the tag signature:
  - [ ] **Local git** (most reliable):
    ```bash
    git fetch --tags origin
    git tag -v v$(cat VERSION)
    ```
  - [ ] **GitHub API (required after push)**:
    ```bash
    make release-verify-remote-tag
    ```
    This confirms the remote annotated tag object records the Infosec tagger identity, points to the reviewed local target, and reports GitHub verification as `valid`.
  - [ ] **GitHub Web UI (note)**: a green "Verified" badge requires the Infosec public key on the GitHub account and the `infosec@3leaps.net` tagger email verified there. The API gate above is the release record.
- [ ] Optional: publish minisign attestation (if enabled):
  - `make release-tag` can produce `dist/release/vX.Y.Z.tag.txt` + `.minisig` when `GOFULMEN_MINISIGN_KEY` and `GOFULMEN_MINISIGN_PUB` are set.
  - These files are **not uploaded automatically**; to distribute them, attach them to a GitHub Release (or another artifact channel):
    ```bash
    gh release create v$(cat VERSION) --notes-file docs/releases/v$(cat VERSION).md dist/release/v$(cat VERSION).tag.txt dist/release/v$(cat VERSION).tag.txt.minisig
    ```
- [ ] Announce / coordinate downstream upgrades as needed (templates, workhorses, CLIs).
