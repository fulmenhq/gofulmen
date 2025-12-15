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

	local gpg_homedir="${GOFULMEN_GPG_HOMEDIR:-${GOFULMEN_GPG_HOME:-}}"
	if [ -n "${GOFULMEN_GPG_HOME:-}" ] && [ -z "${GOFULMEN_GPG_HOMEDIR:-}" ]; then
		echo "warning: GOFULMEN_GPG_HOME is deprecated; use GOFULMEN_GPG_HOMEDIR" >&2
	fi

	if [ -n "${gpg_homedir}" ]; then
		if [ ! -d "${gpg_homedir}" ]; then
			echo "error: GOFULMEN_GPG_HOMEDIR=${gpg_homedir} is not a directory" >&2
			exit 1
		fi
		export GNUPGHOME="${gpg_homedir}"
	fi

	echo "→ Verifying tag signature: $tag"
	git verify-tag "$tag" >/dev/null
	echo "✅ Tag verified: $tag"

	# Optional: verify minisign sidecar signature for the tag attestation.
	if [ -n "${GOFULMEN_MINISIGN_PUB:-}" ]; then
		if ! command -v minisign >/dev/null 2>&1; then
			echo "error: GOFULMEN_MINISIGN_PUB is set but minisign is not found in PATH" >&2
			exit 1
		fi

		local sig_dir="dist/release"
		local payload="${sig_dir}/${tag}.tag.txt"
		local sig="${payload}.minisig"
		if [ -f "${payload}" ] && [ -f "${sig}" ]; then
			minisign -Vm "${payload}" -p "${GOFULMEN_MINISIGN_PUB}" >/dev/null
			echo "✅ Minisign tag attestation verified: ${sig}"
		else
			echo "note: minisign pubkey set but no attestation found at ${sig}" >&2
		fi
	fi
}

main "$@"
