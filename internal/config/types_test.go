package config

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigYAMLMarshal(t *testing.T) {
	cfg := &Config{
		Container: "colima",
		VPN: &VPNConfig{
			Provider:   "openvpn",
			ConfigPath: "~/.config/yar/vpn/client.ovpn",
		},
		Hosts: &HostsConfig{
			Mode:   "etc",
			Suffix: ".local",
		},
		Network: &NetworkConfig{
			Name: "yar-net",
			CIDR: "172.16.34.0/23",
		},
		Secrets: &SecretsConfig{
			Local: &LocalSecretConfig{
				Provider: "pass",
				Store:    "~/.password-store",
				Fallback: true,
			},
		},
		Clusters: map[string]*ClusterConfig{
			"local": {
				Provider: "compose",
			},
			"dev": {
				Provider:  "k8s",
				Context:   "dev-cluster",
				Namespace: "default",
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config to YAML: %v", err)
	}

	// Unmarshal back
	var decoded Config
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal config from YAML: %v", err)
	}

	// Verify key fields
	if decoded.Container != cfg.Container {
		t.Errorf("Container: got %q, want %q", decoded.Container, cfg.Container)
	}
	if decoded.VPN.Provider != cfg.VPN.Provider {
		t.Errorf("VPN.Provider: got %q, want %q", decoded.VPN.Provider, cfg.VPN.Provider)
	}
	if decoded.Network.CIDR != cfg.Network.CIDR {
		t.Errorf("Network.CIDR: got %q, want %q", decoded.Network.CIDR, cfg.Network.CIDR)
	}
	if decoded.Secrets.Local.Fallback != cfg.Secrets.Local.Fallback {
		t.Errorf("Secrets.Local.Fallback: got %v, want %v", decoded.Secrets.Local.Fallback, cfg.Secrets.Local.Fallback)
	}
	if decoded.Clusters["dev"].Context != cfg.Clusters["dev"].Context {
		t.Errorf("Clusters[dev].Context: got %q, want %q", decoded.Clusters["dev"].Context, cfg.Clusters["dev"].Context)
	}
}

func TestConfigJSONMarshal(t *testing.T) {
	cfg := &Config{
		Container: "docker",
		Network: &NetworkConfig{
			Name: "test-net",
			CIDR: "10.0.0.0/16",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config to JSON: %v", err)
	}

	// Unmarshal back
	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal config from JSON: %v", err)
	}

	if decoded.Container != cfg.Container {
		t.Errorf("Container: got %q, want %q", decoded.Container, cfg.Container)
	}
	if decoded.Network.Name != cfg.Network.Name {
		t.Errorf("Network.Name: got %q, want %q", decoded.Network.Name, cfg.Network.Name)
	}
}

func TestConfigOmitEmpty(t *testing.T) {
	cfg := &Config{
		Container: "colima",
		// All optional fields left nil
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	yamlStr := string(data)

	// Optional fields should be omitted
	if contains(yamlStr, "vpn:") {
		t.Error("expected vpn to be omitted when nil")
	}
	if contains(yamlStr, "hosts:") {
		t.Error("expected hosts to be omitted when nil")
	}
	if contains(yamlStr, "clusters:") {
		t.Error("expected clusters to be omitted when nil")
	}
}

func TestProjectYAMLMarshal(t *testing.T) {
	proj := &Project{
		Project: "my-app",
		Environments: map[string]*Environment{
			"local": {
				Cluster: "local",
				Secrets: "local",
			},
			"prod": {
				Cluster: "prod-cluster",
				Secrets: "azure",
			},
		},
		Services: []*Service{
			{
				Name:     "redis",
				Pack:     "redis",
				Replicas: 1,
				Params: map[string]any{
					"passwordRef": "redis_pass",
					"port":        6379,
				},
			},
			{
				Name:     "api",
				Pack:     "node",
				Requires: []string{"redis"},
				Replicas: 3,
				Ingress: &IngressConfig{
					Host: "api.example.com",
					Path: "/",
					TLS:  true,
				},
				Env: map[string]string{
					"NODE_ENV": "production",
				},
				SecretRefs: map[string]string{
					"API_KEY": "api_key_secret",
				},
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(proj)
	if err != nil {
		t.Fatalf("failed to marshal project to YAML: %v", err)
	}

	// Unmarshal back
	var decoded Project
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal project from YAML: %v", err)
	}

	// Verify key fields
	if decoded.Project != proj.Project {
		t.Errorf("Project: got %q, want %q", decoded.Project, proj.Project)
	}
	if len(decoded.Services) != len(proj.Services) {
		t.Errorf("Services count: got %d, want %d", len(decoded.Services), len(proj.Services))
	}
	if decoded.Services[0].Name != proj.Services[0].Name {
		t.Errorf("Services[0].Name: got %q, want %q", decoded.Services[0].Name, proj.Services[0].Name)
	}
	if decoded.Services[1].Ingress.TLS != proj.Services[1].Ingress.TLS {
		t.Errorf("Services[1].Ingress.TLS: got %v, want %v", decoded.Services[1].Ingress.TLS, proj.Services[1].Ingress.TLS)
	}
	if decoded.Environments["prod"].Secrets != proj.Environments["prod"].Secrets {
		t.Errorf("Environments[prod].Secrets: got %q, want %q", decoded.Environments["prod"].Secrets, proj.Environments["prod"].Secrets)
	}
}

func TestProjectJSONMarshal(t *testing.T) {
	proj := &Project{
		Project: "test-project",
		Environments: map[string]*Environment{
			"local": {Cluster: "local", Secrets: "pass"},
		},
		Services: []*Service{
			{Name: "web", Pack: "nginx"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(proj)
	if err != nil {
		t.Fatalf("failed to marshal project to JSON: %v", err)
	}

	// Unmarshal back
	var decoded Project
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal project from JSON: %v", err)
	}

	if decoded.Project != proj.Project {
		t.Errorf("Project: got %q, want %q", decoded.Project, proj.Project)
	}
}

func TestServiceOmitEmpty(t *testing.T) {
	svc := &Service{
		Name: "minimal",
		Pack: "redis",
		// All optional fields left nil/empty
	}

	data, err := yaml.Marshal(svc)
	if err != nil {
		t.Fatalf("failed to marshal service: %v", err)
	}

	yamlStr := string(data)

	// Optional fields should be omitted
	if contains(yamlStr, "namespace:") {
		t.Error("expected namespace to be omitted when empty")
	}
	if contains(yamlStr, "requires:") {
		t.Error("expected requires to be omitted when empty")
	}
	if contains(yamlStr, "ingress:") {
		t.Error("expected ingress to be omitted when nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Verify default values per spec
	if cfg.Container != "colima" {
		t.Errorf("Container: got %q, want %q", cfg.Container, "colima")
	}
	if cfg.Hosts == nil || cfg.Hosts.Mode != "etc" {
		t.Errorf("Hosts.Mode: got %v, want %q", cfg.Hosts, "etc")
	}
	if cfg.Network == nil || cfg.Network.Name != "yar-net" {
		t.Errorf("Network.Name: got %v, want %q", cfg.Network, "yar-net")
	}
	if cfg.Network == nil || cfg.Network.CIDR != "172.16.34.0/23" {
		t.Errorf("Network.CIDR: got %v, want %q", cfg.Network, "172.16.34.0/23")
	}
	if cfg.Secrets == nil || cfg.Secrets.Local == nil {
		t.Fatal("Secrets.Local is nil")
	}
	if cfg.Secrets.Local.Provider != "pass" {
		t.Errorf("Secrets.Local.Provider: got %q, want %q", cfg.Secrets.Local.Provider, "pass")
	}
	if cfg.Secrets.Local.Fallback != true {
		t.Errorf("Secrets.Local.Fallback: got %v, want %v", cfg.Secrets.Local.Fallback, true)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
