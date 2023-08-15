package internal

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --target applicationmetric --clean ../internal/openapi-specs.bundled/application-metric-api.openapi.bundled.yaml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --debug.ignoreNotImplemented "mutualTLS security" --target custommetrics --clean ../internal/openapi-specs.bundled/custom-metrics-api.openapi.bundled.yaml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --target policy --clean ../internal/openapi-specs.bundled/policy-api.openapi.bundled.yaml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --target scalinghistory --clean ../internal/openapi-specs.bundled/scaling-history-api.openapi.bundled.yaml
