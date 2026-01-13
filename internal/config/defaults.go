package config

// DefaultConfig returns sensible defaults for global configuration.
// Used when no config file exists.
func DefaultConfig() *Config {
	return &Config{
		Container: "colima",
		Hosts: &HostsConfig{
			Mode: "etc",
		},
		Network: &NetworkConfig{
			Name: "yar-net",
			CIDR: "172.16.34.0/23",
		},
		Secrets: &SecretsConfig{
			Local: &LocalSecretConfig{
				Provider: "pass",
				Fallback: true,
			},
		},
	}
}
