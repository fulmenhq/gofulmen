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

	if [ -n "${GOFULMEN_GPG_HOME:-}" ]; then
		if [ ! -d "${GOFULMEN_GPG_HOME}" ]; then
			echo "error: GOFULMEN_GPG_HOME=${GOFULMEN_GPG_HOME} is not a directory" >&2
			exit 1
		fi
		export GNUPGHOME="${GOFULMEN_GPG_HOME}"
	fi

	echo "→ Verifying tag signature: $tag"
	git verify-tag "$tag" >/dev/null
	echo "✅ Tag verified: $tag"
}

main "$@"
