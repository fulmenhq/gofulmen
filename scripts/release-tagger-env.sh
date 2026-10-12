#!/usr/bin/env bash
# Shared tagger-identity SSOT for signed-tag create/verify.
# GOFULMEN_TAGGER_NAME and GOFULMEN_TAGGER_EMAIL are required; fail closed if unset.
# Sourced by release-tag.sh, release-verify-tag.sh, and release-verify-remote-tag.sh.

require_tagger_env() {
	if [ -z "${GOFULMEN_TAGGER_NAME:-}" ]; then
		echo "error: GOFULMEN_TAGGER_NAME is required" >&2
		return 1
	fi
	if [ -z "${GOFULMEN_TAGGER_EMAIL:-}" ]; then
		echo "error: GOFULMEN_TAGGER_EMAIL is required" >&2
		return 1
	fi
	case "${GOFULMEN_TAGGER_NAME}" in
	*$'\n'* | *$'\r'* | *'<'* | *'>'*)
		echo "error: GOFULMEN_TAGGER_NAME contains an invalid character" >&2
		return 1
		;;
	esac
	case "${GOFULMEN_TAGGER_EMAIL}" in
	*$'\n'* | *$'\r'* | *'<'* | *'>'* | *' '*)
		echo "error: GOFULMEN_TAGGER_EMAIL contains an invalid character" >&2
		return 1
		;;
	esac
}

assert_tagger_identity() {
	local tagger_name="$1"
	local tagger_email="$2"
	require_tagger_env || return 1
	if [ "${tagger_name}" != "${GOFULMEN_TAGGER_NAME}" ] || [ "${tagger_email}" != "${GOFULMEN_TAGGER_EMAIL}" ]; then
		echo "error: tag object does not record GOFULMEN_TAGGER_NAME/GOFULMEN_TAGGER_EMAIL" >&2
		return 1
	fi
}
