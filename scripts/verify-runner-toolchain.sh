#!/usr/bin/env bash
# Fail-closed runner toolchain gate.
#
# CI and local gated targets require exactly Go 1.26.10 (CVE-2026-111943 /
# GO-2026-6809 in crypto/x509). Go 1.27 is still refused because gosec
# v2.29.0 cannot read Go 1.27 package data. A floor check would accept
# 1.27, so this gate matches the exact patch: `go version` must contain
# `go1.26.10 `.
set -euo pipefail

required_go_token="go1.26.10 "

require_runner_go() {
	local actual
	actual="$(go version 2>/dev/null || true)"
	if ! printf '%s\n' "${actual}" | grep -F -q "${required_go_token}"; then
		echo "ERROR: Need Go 1.26.10; got ${actual:-<empty>}" >&2
		return 1
	fi
}

print_runner_toolchain() {
	echo "PATH=${PATH}"
	echo "go: $(go version)"
	echo "goneat: $(goneat --version 2>/dev/null | head -n1 || echo 'not found')"
	echo "sfetch: $(sfetch --version 2>/dev/null | head -n1 || echo 'not found')"
	echo "actionlint: $(actionlint -version 2>/dev/null || echo 'not found')"
	echo "golangci-lint: $(golangci-lint version 2>/dev/null | head -n1 || echo 'not found')"
	echo "make: $(make --version | head -n1)"
	echo "bash: $(bash --version | head -n1)"
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
	require_runner_go
	if [ "${VERIFY_RUNNER_TOOLCHAIN_QUIET:-}" != "1" ]; then
		print_runner_toolchain
	fi
fi
