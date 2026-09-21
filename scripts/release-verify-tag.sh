#!/usr/bin/env bash
set -euo pipefail

readonly EXPECTED_TAGGER_NAME="FulmenHQ Infosec"
readonly EXPECTED_TAGGER_EMAIL="infosec@3leaps.net"
readonly EXPECTED_PRIMARY_FINGERPRINT="0CACA49B3119B6BC12B2CA11B9B485F294B9FE07"
readonly EXPECTED_SIGNING_SUBKEY="DAB70DD758911B26BB45A08C3B75AC591449DEBE"

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

verify_tag_identity() {
	local tag="$1"
	local tagger_name tagger_email
	tagger_name="$(git for-each-ref --format='%(taggername)' "refs/tags/${tag}")"
	tagger_email="$(git for-each-ref --format='%(taggeremail)' "refs/tags/${tag}")"
	tagger_email="${tagger_email#<}"
	tagger_email="${tagger_email%>}"
	if [ "${tagger_name}" != "${EXPECTED_TAGGER_NAME}" ] || [ "${tagger_email}" != "${EXPECTED_TAGGER_EMAIL}" ]; then
		echo "error: tag object does not record the required Infosec tagger identity" >&2
		exit 1
	fi
}

verify_tag_signature() {
	local tag="$1"
	local verification
	if ! verification="$(git verify-tag --raw "${tag}" 2>&1)"; then
		printf '%s\n' "${verification}" >&2
		exit 1
	fi
	if ! printf '%s\n' "${verification}" | grep -q "\[GNUPG:\] VALIDSIG ${EXPECTED_SIGNING_SUBKEY} "; then
		echo "error: tag signature does not use expected Infosec signing subkey ${EXPECTED_SIGNING_SUBKEY}" >&2
		exit 1
	fi
	if ! printf '%s\n' "${verification}" | grep -q "\[GNUPG:\] VALIDSIG ${EXPECTED_SIGNING_SUBKEY} .* ${EXPECTED_PRIMARY_FINGERPRINT}$"; then
		echo "error: tag signature does not identify expected Infosec primary fingerprint ${EXPECTED_PRIMARY_FINGERPRINT}" >&2
		exit 1
	fi
}

main() {
	local root
	root="$(repo_root)"
	cd "${root}"

	local version
	version="$(read_version)"

	local tag="${GOFULMEN_RELEASE_TAG:-v${version}}"

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

	echo "→ Verifying tag signature: ${tag}"
	verify_tag_signature "${tag}"
	verify_tag_identity "${tag}"
	echo "✅ Tag verified: ${tag}"

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
