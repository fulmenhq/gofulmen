#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-${1:-}}"

if [ -z "$version" ]; then
	echo "❌ VERSION not specified. Usage: make crucible-update VERSION=v0.2.19" >&2
	exit 1
fi

echo "Updating Crucible to $version..."
echo ""
echo "Step 1: Updating .goneat/ssot-consumer.yaml..."
sed -i.bak "s|ref: v[0-9]*\.[0-9]*\.[0-9]*|ref: $version|" .goneat/ssot-consumer.yaml
rm .goneat/ssot-consumer.yaml.bak
echo "✅ Updated ssot-consumer.yaml ref to $version"

echo ""
echo "Step 2: Running make sync to update provenance..."
make sync

echo ""
echo "Step 3: Updating go.mod..."
go get "github.com/fulmenhq/crucible@$version"
go mod tidy
echo "✅ Updated go.mod to $version"

echo ""
echo "Step 4: Running tests to verify compatibility..."
go test ./crucible -run TestCrucibleVersionMatchesMetadata -v

echo ""
echo "✅ Crucible updated successfully to $version"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Run full checks: make check-all"
echo "  3. Commit changes with proper attribution"
