#!/usr/bin/env bash
# Unit paths for required tagger-env SSOT. Uses synthetic identities only.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=scripts/release-tagger-env.sh
source "${root}/scripts/release-tagger-env.sh"

failures=0

expect_fail() {
	local label="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		echo "FAIL: expected failure: ${label}" >&2
		failures=$((failures + 1))
	else
		echo "PASS: ${label}"
	fi
}

expect_pass() {
	local label="$1"
	shift
	if "$@"; then
		echo "PASS: ${label}"
	else
		echo "FAIL: expected success: ${label}" >&2
		failures=$((failures + 1))
	fi
}

unset GOFULMEN_TAGGER_NAME GOFULMEN_TAGGER_EMAIL
expect_fail "unset name and email" require_tagger_env

GOFULMEN_TAGGER_NAME="Unit Test Tagger"
unset GOFULMEN_TAGGER_EMAIL
expect_fail "unset email" require_tagger_env

unset GOFULMEN_TAGGER_NAME
GOFULMEN_TAGGER_EMAIL="tagger@example.test"
expect_fail "unset name" require_tagger_env

GOFULMEN_TAGGER_NAME="Unit Test Tagger"
GOFULMEN_TAGGER_EMAIL="tagger@example.test"
expect_pass "valid env" require_tagger_env
expect_pass "matching identity" assert_tagger_identity "Unit Test Tagger" "tagger@example.test"
expect_fail "mismatched name" assert_tagger_identity "Other Tagger" "tagger@example.test"
expect_fail "mismatched email" assert_tagger_identity "Unit Test Tagger" "other@example.test"

GOFULMEN_TAGGER_NAME=$'bad\nname'
GOFULMEN_TAGGER_EMAIL="tagger@example.test"
expect_fail "newline in name" require_tagger_env

GOFULMEN_TAGGER_NAME="Unit Test Tagger"
GOFULMEN_TAGGER_EMAIL=$'bad\nemail@example.test'
expect_fail "newline in email" require_tagger_env

GOFULMEN_TAGGER_NAME="Unit <Test>"
GOFULMEN_TAGGER_EMAIL="tagger@example.test"
expect_fail "angle brackets in name" require_tagger_env

GOFULMEN_TAGGER_NAME="Unit Test Tagger"
GOFULMEN_TAGGER_EMAIL="tagger @example.test"
expect_fail "space in email" require_tagger_env

# Script entry points fail closed before ceremony work.
script_fail() {
	local label="$1"
	shift
	local output
	if output="$("$@" 2>&1)"; then
		echo "FAIL: expected failure: ${label}" >&2
		failures=$((failures + 1))
		return
	fi
	if ! printf '%s\n' "${output}" | grep -q 'GOFULMEN_TAGGER_.* is required'; then
		echo "FAIL: ${label}: missing required-env error" >&2
		printf '%s\n' "${output}" >&2
		failures=$((failures + 1))
		return
	fi
	echo "PASS: ${label}"
}

unset GOFULMEN_TAGGER_NAME GOFULMEN_TAGGER_EMAIL
script_fail "release-tag.sh unset env" "${root}/scripts/release-tag.sh"
script_fail "release-verify-tag.sh unset env" "${root}/scripts/release-verify-tag.sh"
script_fail "release-verify-remote-tag.sh unset env" "${root}/scripts/release-verify-remote-tag.sh"

if [ "${failures}" -ne 0 ]; then
	echo "error: ${failures} tagger-env unit path(s) failed" >&2
	exit 1
fi

echo "✅ release-tagger-env unit paths passed"
