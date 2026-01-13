package config

// Config represents global yar configuration (~/.config/yar/config.yaml)
type Config struct {
	Container string                    `yaml:"container" json:"container"`
	VPN       *VPNConfig                `yaml:"vpn,omitempty" json:"vpn,omitempty"`
	Hosts     *HostsConfig              `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	Network   *NetworkConfig            `yaml:"network,omitempty" json:"network,omitempty"`
	Secrets   *SecretsConfig            `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Clusters  map[string]*ClusterConfig `yaml:"clusters,omitempty" json:"clusters,omitempty"`
}

// VPNConfig configures VPN connectivity
type VPNConfig struct {
	Provider   string `yaml:"provider" json:"provider"`
	ConfigPath string `yaml:"configPath" json:"configPath"`
}

// HostsConfig configures host resolution
type HostsConfig struct {
	Mode   string `yaml:"mode" json:"mode"`
	Suffix string `yaml:"suffix,omitempty" json:"suffix,omitempty"`
}

// NetworkConfig configures Docker network settings
type NetworkConfig struct {
	Name string `yaml:"name" json:"name"`
	CIDR string `yaml:"cidr" json:"cidr"`
}

// SecretsConfig configures secret providers
type SecretsConfig struct {
	Local     *LocalSecretConfig               `yaml:"local" json:"local"`
	Providers map[string]*SecretProviderConfig `yaml:"providers,omitempty" json:"providers,omitempty"`
}

// LocalSecretConfig configures the local secret store
type LocalSecretConfig struct {
	Provider string `yaml:"provider" json:"provider"`
	Store    string `yaml:"store,omitempty" json:"store,omitempty"`
	Fallback bool   `yaml:"fallback" json:"fallback"`
}

// SecretProviderConfig configures a remote secret provider
type SecretProviderConfig struct {
	Type   string         `yaml:"type" json:"type"`
	Config map[string]any `yaml:",inline" json:"-"`
}

// ClusterConfig configures a deployment cluster
type ClusterConfig struct {
	Provider  string `yaml:"provider" json:"provider"`
	Context   string `yaml:"context,omitempty" json:"context,omitempty"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

// Project represents project configuration (yar.yaml)
type Project struct {
	Project      string                  `yaml:"project" json:"project"`
	Environments map[string]*Environment `yaml:"environments" json:"environments"`
	Services     []*Service              `yaml:"services" json:"services"`
}

// Environment defines a deployment environment
type Environment struct {
	Cluster string `yaml:"cluster" json:"cluster"`
	Secrets string `yaml:"secrets" json:"secrets"`
}

// Service defines a service in the project
type Service struct {
	Name       string            `yaml:"name" json:"name"`
	Namespace  string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Pack       string            `yaml:"pack" json:"pack"`
	Requires   []string          `yaml:"requires,omitempty" json:"requires,omitempty"`
	Replicas   int               `yaml:"replicas,omitempty" json:"replicas,omitempty"`
	Params     map[string]any    `yaml:"params,omitempty" json:"params,omitempty"`
	Ingress    *IngressConfig    `yaml:"ingress,omitempty" json:"ingress,omitempty"`
	Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	SecretRefs map[string]string `yaml:"secretRefs,omitempty" json:"secretRefs,omitempty"`
}

// IngressConfig configures ingress for a service
type IngressConfig struct {
	Host string `yaml:"host" json:"host"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	TLS  bool   `yaml:"tls,omitempty" json:"tls,omitempty"`
}
