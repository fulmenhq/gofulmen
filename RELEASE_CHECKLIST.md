# Release Checklist (gofulmen)

gofulmen is a pure library module: releases are primarily **signed git tags** (`vX.Y.Z`) consumed via the Go
module proxy/VCS. We do not ship binaries.

This checklist is expected to become a shared ecosystem pattern (gofulmen/pyfulmen/tsfulmen/rsfulmen) once it’s
validated here and then codified into Crucible.

## Variables (Quick Reference)

- `RELEASE_TAG`: optional override tag (recommended for manual release commands; e.g. `v0.1.21`)
- `GOFULMEN_RELEASE_TAG`: optional override tag (alias; scripts accept either)
- `GOFULMEN_GPG_HOME`: recommended dedicated signing keyring directory (separate from personal `~/.gnupg`)
- `GOFULMEN_PGP_KEY_ID`: optional key id/email/fingerprint for signing
- `GOFULMEN_ALLOW_NON_MAIN=1`: optional override to tag from a non-`main` branch (not recommended)
- `GOFULMEN_REQUIRE_TAG=1`: enforce “must be on an exact tag” (intended for CI guard usage)

Note: `RELEASE_TAG`/`GOFULMEN_RELEASE_TAG` are not secrets and typically aren’t stored in encrypted env bundles.

## Pre-Release

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

## Tagging (Signed Tag Required)

- [ ] Create and verify the signed tag:
  ```bash
  make release-tag
  ```
- [ ] Push:
  ```bash
  git push origin main
  git push origin v$(cat VERSION)
  ```

## Post-Release

- [ ] Spot-check downstream consumption:
  ```bash
  go list -m github.com/fulmenhq/gofulmen@v$(cat VERSION)
  ```
- [ ] Announce / coordinate downstream upgrades as needed (templates, workhorses, CLIs).
