#!/usr/bin/env bash
# Unit paths for the fail-closed Go 1.26.9 runner toolchain gate.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
gate="${root}/scripts/verify-runner-toolchain.sh"

failures=0

fake_go_dir() {
	local version_line="$1"
	local dir
	dir="$(mktemp -d)"
	cat >"${dir}/go" <<EOF
#!/bin/sh
echo "${version_line}"
EOF
	chmod +x "${dir}/go"
	printf '%s\n' "${dir}"
}

run_gate() {
	local version_line="$1"
	local dir
	dir="$(fake_go_dir "${version_line}")"
	local rc=0
	PATH="${dir}:${PATH}" VERIFY_RUNNER_TOOLCHAIN_QUIET=1 "${gate}" || rc=$?
	rm -rf "${dir}"
	return "${rc}"
}

expect_fail() {
	local label="$1"
	local version_line="$2"
	if run_gate "${version_line}" >/dev/null 2>&1; then
		echo "FAIL: expected failure: ${label}" >&2
		failures=$((failures + 1))
	else
		echo "PASS: ${label}"
	fi
}

expect_pass() {
	local label="$1"
	local version_line="$2"
	if run_gate "${version_line}" >/dev/null 2>&1; then
		echo "PASS: ${label}"
	else
		echo "FAIL: expected success: ${label}" >&2
		failures=$((failures + 1))
	fi
}

expect_pass "exact go1.26.9" "go version go1.26.9 linux/amd64"
expect_fail "go1.26.8" "go version go1.26.8 linux/amd64"
expect_fail "go1.26.10" "go version go1.26.10 linux/amd64"
expect_fail "go1.27.2" "go version go1.27.2 linux/amd64"
expect_fail "go1.25.14" "go version go1.25.14 linux/amd64"
expect_fail "empty go version" ""

if [ "${failures}" -ne 0 ]; then
	echo "error: ${failures} runner-toolchain unit path(s) failed" >&2
	exit 1
fi

echo "✅ runner-toolchain unit paths passed"
