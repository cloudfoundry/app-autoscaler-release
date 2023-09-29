package internal

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --target applicationmetric --clean ../../../../../../api/application-metric-api.openapi.yaml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --debug.ignoreNotImplemented "mutualTLS security" --target custommetrics --clean ../../../../../../api/custom-metrics-api.openapi.yaml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --no-server --allow-remote --target policy --clean ../../../../../../api/policy-api.openapi.yaml
