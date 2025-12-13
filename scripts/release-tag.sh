#!/usr/bin/env bash
set -euo pipefail

repo_root() {
	git rev-parse --show-toplevel
}

read_version() {
	if [ ! -f VERSION ]; then
		echo "error: VERSION file not found" >&2
		exit 1
	fi
	tr -d ' \t\r\n' <VERSION
}

main() {
	local root
	root="$(repo_root)"
	cd "$root"

	local version
	version="$(read_version)"

	local tag="${GOFULMEN_RELEASE_TAG:-${RELEASE_TAG:-v${version}}}"

	if ! [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		echo "error: invalid release tag '$tag' (expected vMAJOR.MINOR.PATCH)" >&2
		exit 1
	fi

	if [ -n "$(git status --porcelain)" ]; then
		echo "error: working tree is not clean (commit or stash changes before tagging)" >&2
		git status --porcelain >&2
		exit 1
	fi

	local branch
	branch="$(git branch --show-current 2>/dev/null || true)"
	if [ "$branch" != "main" ] && [ "${GOFULMEN_ALLOW_NON_MAIN:-}" != "1" ]; then
		echo "error: refusing to tag from branch '$branch' (set GOFULMEN_ALLOW_NON_MAIN=1 to override)" >&2
		exit 1
	fi

	if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
		echo "error: tag $tag already exists" >&2
		exit 1
	fi

	if [ -n "${GOFULMEN_GPG_HOME:-}" ]; then
		if [ ! -d "${GOFULMEN_GPG_HOME}" ]; then
			echo "error: GOFULMEN_GPG_HOME=${GOFULMEN_GPG_HOME} is not a directory" >&2
			exit 1
		fi
		export GNUPGHOME="${GOFULMEN_GPG_HOME}"
	fi

	if [ -n "${GOFULMEN_PGP_KEY_ID:-}" ] && [ -z "${GOFULMEN_GPG_HOME:-}" ]; then
		echo "error: GOFULMEN_PGP_KEY_ID is set but GOFULMEN_GPG_HOME is not; set a dedicated signing homedir" >&2
		exit 1
	fi

	echo "→ Creating signed tag: $tag"

	if [ -n "${GOFULMEN_PGP_KEY_ID:-}" ]; then
		git tag -s -a "$tag" -u "${GOFULMEN_PGP_KEY_ID}" -m "Release $tag"
	else
		git tag -s -a "$tag" -m "Release $tag"
	fi

	echo "→ Verifying tag signature: $tag"
	git verify-tag "$tag" >/dev/null

	echo "✅ Created and verified signed tag: $tag"
	echo "Next:"
	echo "  git push origin main"
	echo "  git push origin $tag"
}

main "$@"
