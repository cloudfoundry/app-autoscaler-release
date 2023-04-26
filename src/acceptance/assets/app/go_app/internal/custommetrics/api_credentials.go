package api

type CustomMetricsCredentials struct {
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
	URL      string `mapstructure:"url" json:"url"`
	MtlsURL  string `mapstructure:"mtls_url" json:"mtls_url"`
}
