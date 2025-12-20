#!/usr/bin/env bash
set -euo pipefail

# fixup-hooks-noglob.sh
#
# Inserts `set -f` (noglob) into generated bash hooks after `set -euo pipefail`.
# This is a temporary downstream mitigation until goneat templates consistently
# emit noglob.

repo_root() {
	git rev-parse --show-toplevel 2>/dev/null || pwd
}

add_noglob() {
	local hook_path="$1"

	if [ ! -f "$hook_path" ]; then
		echo "❌ Hook not found: $hook_path" >&2
		return 1
	fi

	if rg -n '^set -f$' "$hook_path" >/dev/null 2>&1; then
		return 0
	fi

	python3 - <<'PY'
import os
import pathlib

hook_path_value = os.environ.get("HOOK_PATH")
if not hook_path_value:
    raise SystemExit("HOOK_PATH env var is required")

hook_path = pathlib.Path(hook_path_value)
content = hook_path.read_text(encoding="utf-8").splitlines(keepends=True)

out = []
injected = False
for line in content:
    out.append(line)
    if not injected and line.strip() == "set -euo pipefail":
        out.append("set -f\n")
        injected = True

if not injected:
    raise SystemExit("did not find 'set -euo pipefail' line")

hook_path.write_text("".join(out), encoding="utf-8")
PY
}

main() {
	local root
	root="$(repo_root)"

	local hooks_dir="$root/.goneat/hooks"
	local pre_commit="$hooks_dir/pre-commit"
	local pre_push="$hooks_dir/pre-push"

	HOOK_PATH="$pre_commit" add_noglob "$pre_commit"
	HOOK_PATH="$pre_push" add_noglob "$pre_push"

	echo "✅ Inserted 'set -f' into goneat hooks (if missing)"
}

main "$@"
