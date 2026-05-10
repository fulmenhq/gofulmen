#!/usr/bin/env bash
set -euo pipefail

ensure_go_licenses() {
	if ! command -v go-licenses >/dev/null 2>&1; then
		echo "Installing go-licenses..."
		go install github.com/google/go-licenses@latest
	fi
}

inventory() {
	echo "🔎 Generating license inventory (CSV)..."
	mkdir -p docs/licenses dist/reports
	ensure_go_licenses
	go-licenses csv ./... >docs/licenses/inventory.csv
	echo "✅ Wrote docs/licenses/inventory.csv"
}

save_texts() {
	echo "📄 Saving third-party license texts..."
	rm -rf docs/licenses/third-party
	ensure_go_licenses
	go-licenses save ./... --save_path=docs/licenses/third-party
	echo "✅ Saved third-party licenses to docs/licenses/third-party"
}

audit() {
	local forbidden='GPL|LGPL|AGPL|MPL|CDDL'
	local out

	echo "🧪 Auditing dependency licenses..."
	mkdir -p dist/reports
	ensure_go_licenses
	out="$(go-licenses csv ./...)"
	echo "$out" >dist/reports/license-inventory.csv

	if echo "$out" | grep -E "$forbidden" >/dev/null; then
		echo "❌ Forbidden license detected. See dist/reports/license-inventory.csv"
		exit 1
	fi
	echo "✅ No forbidden licenses detected"
}

case "${1:-}" in
inventory)
	inventory
	;;
save)
	save_texts
	;;
audit)
	audit
	;;
*)
	echo "Usage: $0 {inventory|save|audit}" >&2
	exit 1
	;;
esac
