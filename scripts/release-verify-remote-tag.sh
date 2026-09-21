#!/usr/bin/env bash
set -euo pipefail

readonly EXPECTED_TAGGER_NAME="FulmenHQ Infosec"
readonly EXPECTED_TAGGER_EMAIL="infosec@3leaps.net"
readonly GITHUB_REPOSITORY="fulmenhq/gofulmen"

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

require_value() {
	local value="$1"
	local description="$2"
	if [ -z "${value}" ] || [ "${value}" = "null" ]; then
		echo "error: GitHub did not return ${description}" >&2
		exit 1
	fi
}

main() {
	local root
	root="$(repo_root)"
	cd "${root}"

	if ! command -v gh >/dev/null 2>&1; then
		echo "error: gh is required to verify the remote tag object" >&2
		exit 1
	fi

	local version tag ref_sha tagger_name tagger_email target_sha verified reason local_target
	version="$(read_version)"
	tag="${GOFULMEN_RELEASE_TAG:-v${version}}"

	if ! [[ "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		echo "error: invalid release tag '${tag}' (expected vMAJOR.MINOR.PATCH)" >&2
		exit 1
	fi

	ref_sha="$(gh api "repos/${GITHUB_REPOSITORY}/git/ref/tags/${tag}" --jq '.object.sha')"
	require_value "${ref_sha}" "the annotated tag object SHA"
	tagger_name="$(gh api "repos/${GITHUB_REPOSITORY}/git/tags/${ref_sha}" --jq '.tagger.name')"
	tagger_email="$(gh api "repos/${GITHUB_REPOSITORY}/git/tags/${ref_sha}" --jq '.tagger.email')"
	target_sha="$(gh api "repos/${GITHUB_REPOSITORY}/git/tags/${ref_sha}" --jq '.object.sha')"
	verified="$(gh api "repos/${GITHUB_REPOSITORY}/git/tags/${ref_sha}" --jq '.verification.verified')"
	reason="$(gh api "repos/${GITHUB_REPOSITORY}/git/tags/${ref_sha}" --jq '.verification.reason')"

	if [ "${tagger_name}" != "${EXPECTED_TAGGER_NAME}" ] || [ "${tagger_email}" != "${EXPECTED_TAGGER_EMAIL}" ]; then
		echo "error: remote tag object does not record the required Infosec tagger identity" >&2
		exit 1
	fi
	if [ "${verified}" != "true" ] || [ "${reason}" != "valid" ]; then
		echo "error: GitHub did not verify the annotated tag object (reason: ${reason})" >&2
		exit 1
	fi

	local_target="$(git rev-parse "${tag}^{}" 2>/dev/null || true)"
	if [ -z "${local_target}" ]; then
		echo "error: local tag is unavailable; fetch the pushed tag before remote verification" >&2
		exit 1
	fi
	if [ "${target_sha}" != "${local_target}" ]; then
		echo "error: remote tag target does not match the local reviewed tag" >&2
		exit 1
	fi

	echo "✅ Remote tag verified: ${tag}"
	echo "   tag object: ${ref_sha}"
	echo "   target: ${target_sha}"
}

main "$@"
