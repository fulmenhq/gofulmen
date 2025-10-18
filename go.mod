module github.com/fulmenhq/gofulmen

go 1.25.1

require (
	github.com/fulmenhq/crucible v0.0.0-00010101000000-000000000000
	github.com/mattn/go-runewidth v0.0.19
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

// Development only: use local crucible
replace github.com/fulmenhq/crucible => ../crucible/lang/go
