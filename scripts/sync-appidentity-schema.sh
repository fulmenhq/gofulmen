#!/usr/bin/env bash
set -euo pipefail

# sync-appidentity-schema.sh
#
# gofulmen embeds an app identity schema for runtime validation:
#   appidentity/app-identity.schema.json
#
# Crucible SSOT also provides the schema under synced assets:
#   schemas/crucible-go/config/repository/app-identity/v1.0.0/app-identity.schema.json
#
# This script keeps the embedded copy in sync to avoid drift.

repo_root() {
	git rev-parse --show-toplevel 2>/dev/null || pwd
}

main() {
	local root
	root="$(repo_root)"

	local src="$root/schemas/crucible-go/config/repository/app-identity/v1.0.0/app-identity.schema.json"
	local dst="$root/appidentity/app-identity.schema.json"

	if [ ! -f "$src" ]; then
		echo "❌ Crucible app-identity schema not found: $src" >&2
		exit 1
	fi

	cp "$src" "$dst"
	echo "✅ Synced appidentity schema from Crucible"
}

main "$@"
