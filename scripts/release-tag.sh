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

setup_gpg_tty() {
	# When using passphrase-protected keys, gpg will invoke pinentry.
	# Ensure it has a real TTY to talk to, otherwise signing fails with:
	# "Inappropriate ioctl for device".
	if [ ! -t 0 ] || [ ! -t 1 ]; then
		echo "error: no TTY available for interactive gpg signing" >&2
		echo "hint: run make release-tag in an interactive terminal" >&2
		echo "hint: export GPG_TTY=\"$(tty)\" && gpg-connect-agent updatestartuptty /bye" >&2
		exit 1
	fi

	if command -v tty >/dev/null 2>&1; then
		local tty_path
		tty_path="$(tty 2>/dev/null || true)"
		if [ -n "${tty_path}" ] && [ "${tty_path}" != "not a tty" ]; then
			export GPG_TTY="${tty_path}"
			gpg-connect-agent updatestartuptty /bye >/dev/null 2>&1 || true
		fi
	fi
}

require_release_identity() {
	local tagger_name="${GOFULMEN_TAGGER_NAME:-${EXPECTED_TAGGER_NAME}}"
	local tagger_email="${GOFULMEN_TAGGER_EMAIL:-${EXPECTED_TAGGER_EMAIL}}"
	local signing_key="${GOFULMEN_PGP_KEY_ID:-${EXPECTED_SIGNING_SUBKEY}}"

	if [ "${tagger_name}" != "${EXPECTED_TAGGER_NAME}" ]; then
		echo "error: GOFULMEN_TAGGER_NAME must be ${EXPECTED_TAGGER_NAME}" >&2
		exit 1
	fi
	if [ "${tagger_email}" != "${EXPECTED_TAGGER_EMAIL}" ]; then
		echo "error: GOFULMEN_TAGGER_EMAIL must be ${EXPECTED_TAGGER_EMAIL}" >&2
		exit 1
	fi
	if [ "${signing_key}" != "${EXPECTED_SIGNING_SUBKEY}" ]; then
		echo "error: GOFULMEN_PGP_KEY_ID must explicitly select ${EXPECTED_SIGNING_SUBKEY}" >&2
		exit 1
	fi

	GOFULMEN_TAGGER_NAME="${tagger_name}"
	GOFULMEN_TAGGER_EMAIL="${tagger_email}"
	GOFULMEN_PGP_KEY_ID="${signing_key}"
}

ensure_gpg_signing_ready() {
	if ! command -v gpg >/dev/null 2>&1; then
		echo "error: gpg not found in PATH (required for signed tags)" >&2
		echo "hint: see RELEASE_CHECKLIST.md (Tagging section)" >&2
		exit 1
	fi

	local listing
	listing="$(gpg --list-secret-keys --with-colons --fingerprint --fingerprint "${EXPECTED_PRIMARY_FINGERPRINT}" 2>/dev/null || true)"
	if ! printf '%s\n' "${listing}" | grep -q '^sec'; then
		echo "error: no usable Infosec secret key found for ${EXPECTED_PRIMARY_FINGERPRINT}" >&2
		echo "hint: set GOFULMEN_GPG_HOMEDIR to the dedicated Infosec signing keyring" >&2
		echo "hint: see RELEASE_CHECKLIST.md (Tagging section)" >&2
		exit 1
	fi
	if ! printf '%s\n' "${listing}" | grep -q "^fpr:::::::::${EXPECTED_PRIMARY_FINGERPRINT}:"; then
		echo "error: expected Infosec primary fingerprint is unavailable" >&2
		exit 1
	fi
	if ! printf '%s\n' "${listing}" | grep -q "^fpr:::::::::${EXPECTED_SIGNING_SUBKEY}:"; then
		echo "error: expected Infosec signing subkey is unavailable" >&2
		exit 1
	fi
}

require_origin_main_tip() {
	local origin_main head
	origin_main="$(git rev-parse --verify refs/remotes/origin/main 2>/dev/null || true)"
	if [ -z "${origin_main}" ]; then
		echo "error: origin/main is unavailable; fetch origin/main before tagging" >&2
		exit 1
	fi
	head="$(git rev-parse HEAD)"
	if [ "${head}" != "${origin_main}" ]; then
		echo "error: HEAD does not match origin/main; fetch and fast-forward before tagging" >&2
		exit 1
	fi
}

verify_local_tag_object() {
	local tag="$1"
	local tagger_name tagger_email tag_target head
	tagger_name="$(git for-each-ref --format='%(taggername)' "refs/tags/${tag}")"
	tagger_email="$(git for-each-ref --format='%(taggeremail)' "refs/tags/${tag}")"
	tagger_email="${tagger_email#<}"
	tagger_email="${tagger_email%>}"
	tag_target="$(git rev-parse "${tag}^{}")"
	head="$(git rev-parse HEAD)"

	if [ "${tagger_name}" != "${EXPECTED_TAGGER_NAME}" ] || [ "${tagger_email}" != "${EXPECTED_TAGGER_EMAIL}" ]; then
		echo "error: tag object does not record the required Infosec tagger identity" >&2
		exit 1
	fi
	if [ "${tag_target}" != "${head}" ]; then
		echo "error: tag object does not point to the intended main commit" >&2
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

	if ! [[ "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		echo "error: invalid release tag '${tag}' (expected vMAJOR.MINOR.PATCH)" >&2
		exit 1
	fi

	if [ -n "$(git status --porcelain)" ]; then
		echo "error: working tree is not clean (commit or stash changes before tagging)" >&2
		git status --porcelain >&2
		exit 1
	fi

	local branch
	branch="$(git branch --show-current 2>/dev/null || true)"
	if [ "${branch}" != "main" ] && [ "${GOFULMEN_ALLOW_NON_MAIN:-}" != "1" ]; then
		echo "error: refusing to tag from branch '${branch}' (set GOFULMEN_ALLOW_NON_MAIN=1 to override)" >&2
		exit 1
	fi

	if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
		echo "error: tag ${tag} already exists" >&2
		exit 1
	fi

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

	if [ -z "${gpg_homedir}" ]; then
		echo "error: GOFULMEN_GPG_HOMEDIR is required; set the dedicated Infosec signing homedir" >&2
		exit 1
	fi

	require_release_identity
	require_origin_main_tip
	setup_gpg_tty
	ensure_gpg_signing_ready

	echo "→ Creating signed tag: ${tag}"

	local tag_err
	tag_err="$(mktemp)"

	if ! git -c user.name="${GOFULMEN_TAGGER_NAME}" -c user.email="${GOFULMEN_TAGGER_EMAIL}" tag -s -a "${tag}" -u "${GOFULMEN_PGP_KEY_ID}" -m "Release ${tag}" 2>"${tag_err}"; then
		cat "${tag_err}" >&2
		if grep -qi "no secret key" "${tag_err}"; then
			echo "hint: no Infosec secret key available (check GOFULMEN_GPG_HOMEDIR/GOFULMEN_PGP_KEY_ID)" >&2
			echo "hint: see RELEASE_CHECKLIST.md (Tagging section)" >&2
		fi
		rm -f "${tag_err}"
		exit 1
	fi

	rm -f "${tag_err}"

	echo "→ Verifying tag signature: ${tag}"
	git verify-tag "${tag}" >/dev/null
	verify_local_tag_object "${tag}"

	echo "✅ Created and verified signed tag: ${tag}"

	# Optional: produce a minisign signature for a deterministic tag attestation.
	# This does NOT modify the git tag object; it creates a sidecar signature
	# that can be uploaded as a release asset.
	if [ -n "${GOFULMEN_MINISIGN_KEY:-}" ] || [ -n "${GOFULMEN_MINISIGN_PUB:-}" ]; then
		if ! command -v minisign >/dev/null 2>&1; then
			echo "error: minisign requested but not found in PATH" >&2
			exit 1
		fi
		if [ -z "${GOFULMEN_MINISIGN_KEY:-}" ] || [ -z "${GOFULMEN_MINISIGN_PUB:-}" ]; then
			echo "error: minisign requires both GOFULMEN_MINISIGN_KEY and GOFULMEN_MINISIGN_PUB" >&2
			exit 1
		fi

		local out_dir="dist/release"
		mkdir -p "${out_dir}"
		local payload="${out_dir}/${tag}.tag.txt"

		local tag_object
		tag_object="$(git rev-parse "${tag}^{tag}")"
		local tag_target
		tag_target="$(git rev-parse "${tag}^{}")"

		cat >"${payload}" <<EOF2
tag: ${tag}
tag_object: ${tag_object}
tag_target: ${tag_target}
EOF2

		echo "→ Minisign tag attestation: ${payload}"
		minisign -Sm "${payload}" -s "${GOFULMEN_MINISIGN_KEY}"
		minisign -Vm "${payload}" -p "${GOFULMEN_MINISIGN_PUB}" >/dev/null
		echo "✅ Minisign signature verified: ${payload}.minisig"
	fi

	echo "Next:"
	echo "  git push origin main"
	echo "  git push origin ${tag}"
	echo "  make release-verify-remote-tag"
	if [ -f "dist/release/${tag}.tag.txt.minisig" ]; then
		echo "  # Optional: upload dist/release/${tag}.tag.txt* as release assets"
	fi
}

main "$@"
