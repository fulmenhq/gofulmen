#!/usr/bin/env bash
set -euo pipefail

usage() {
	echo "Usage: make export-schema SCHEMA_ID=... OUT=output.json" >&2
}

export_schema() {
	local schema_id="${SCHEMA_ID:-}"
	local out="${OUT:-}"

	if [ -z "$schema_id" ]; then
		echo "❌ SCHEMA_ID not specified. Usage: make export-schema SCHEMA_ID=observability/logging/v1.0.0/log-event.schema.json OUT=output.json" >&2
		exit 1
	fi
	if [ -z "$out" ]; then
		usage
		exit 1
	fi

	echo "Exporting schema $schema_id to $out..."
	go run ./cmd/gofulmen-export-schema --schema-id="$schema_id" --out="$out" --no-validate
	echo "✅ Schema exported successfully"
}

export_example() {
	echo "Exporting example logging schema..."
	mkdir -p vendor/crucible/schemas
	go run ./cmd/gofulmen-export-schema \
		--schema-id=observability/logging/v1.0.0/log-event.schema.json \
		--out=vendor/crucible/schemas/logging-event.schema.json \
		--no-validate
	echo "✅ Example schema exported to vendor/crucible/schemas/logging-event.schema.json"
}

case "${1:-}" in
export)
	export_schema
	;;
example)
	export_example
	;;
*)
	usage
	exit 1
	;;
esac
