#!/usr/bin/env bash
set -euo pipefail

# verify-hooks-compat.sh
#
# Guards against known goneat hook template regressions:
# - Stray trailing brace (e.g. passed!"} / passed!})
# - Missing `set -f` (noglob), which can cause glob patterns to expand
#
# Reference: goneat/.plans/memos/fulmenhq/20251220-memo-on-hooks-compatibility.md

repo_root() {
	git rev-parse --show-toplevel 2>/dev/null || pwd
}

main() {
	local root
	root="$(repo_root)"

	local hooks_dir="$root/.goneat/hooks"
	local pre_commit="$hooks_dir/pre-commit"
	local pre_push="$hooks_dir/pre-push"

	if [ ! -f "$pre_commit" ] || [ ! -f "$pre_push" ]; then
		echo "❌ goneat hooks not found under $hooks_dir" >&2
		echo "👉 Run: goneat hooks generate --with-guardian && goneat hooks install" >&2
		exit 1
	fi

	local bad_pattern='passed!"}\|passed!}'
	if rg -n "$bad_pattern" "$pre_commit" "$pre_push" >/dev/null 2>&1; then
		echo "❌ Hook compatibility check failed: stray brace detected" >&2
		rg -n "$bad_pattern" "$pre_commit" "$pre_push" || true
		echo "" >&2
		echo "👉 Regenerate hooks with a fixed goneat version." >&2
		echo "   goneat hooks generate --with-guardian && goneat hooks install" >&2
		exit 1
	fi

	local missing_noglob=0
	for hook in "$pre_commit" "$pre_push"; do
		if ! rg -n '^set -f$' "$hook" >/dev/null 2>&1; then
			echo "❌ Hook compatibility check failed: missing 'set -f' (noglob) in $hook" >&2
			missing_noglob=1
		fi
	done

	if [ "$missing_noglob" -ne 0 ]; then
		echo "" >&2
		echo "👉 Temporary workaround until upstream templates guarantee noglob:" >&2
		echo "   ./scripts/fixup-hooks-noglob.sh" >&2
		exit 1
	fi

	echo "✅ Hook compatibility check passed"
}

main "$@"
