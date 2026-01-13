package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yar-run/yar/internal/errors"
)

// ValidateConfig validates a Config struct against business rules.
// Note: Full JSON Schema validation would require embedding schemas,
// so we perform key validations programmatically here.
func ValidateConfig(cfg *Config) error {
	var errs []string

	// container is required
	if cfg.Container == "" {
		errs = append(errs, "container is required")
	} else {
		// Validate container value
		validContainers := map[string]bool{
			"colima": true, "docker": true, "nerdctl": true, "podman": true,
		}
		if !validContainers[cfg.Container] {
			errs = append(errs, fmt.Sprintf("container must be one of: colima, docker, nerdctl, podman (got %q)", cfg.Container))
		}
	}

	// Validate VPN if present
	if cfg.VPN != nil {
		validVPN := map[string]bool{
			"openvpn": true, "wireguard": true, "tailscale": true,
		}
		if cfg.VPN.Provider != "" && !validVPN[cfg.VPN.Provider] {
			errs = append(errs, fmt.Sprintf("vpn.provider must be one of: openvpn, wireguard, tailscale (got %q)", cfg.VPN.Provider))
		}
	}

	// Validate hosts if present
	if cfg.Hosts != nil {
		validModes := map[string]bool{"etc": true, "kubedns": true}
		if cfg.Hosts.Mode != "" && !validModes[cfg.Hosts.Mode] {
			errs = append(errs, fmt.Sprintf("hosts.mode must be one of: etc, kubedns (got %q)", cfg.Hosts.Mode))
		}
	}

	// Validate clusters if present
	for name, cluster := range cfg.Clusters {
		if cluster.Provider == "" {
			errs = append(errs, fmt.Sprintf("clusters.%s.provider is required", name))
		} else {
			validProviders := map[string]bool{"compose": true, "k8s": true}
			if !validProviders[cluster.Provider] {
				errs = append(errs, fmt.Sprintf("clusters.%s.provider must be one of: compose, k8s (got %q)", name, cluster.Provider))
			}
		}
	}

	// Validate secrets.local if present
	if cfg.Secrets != nil && cfg.Secrets.Local != nil {
		validLocal := map[string]bool{
			"pass": true, "keychain": true, "credential-manager": true, "auto": true,
		}
		if cfg.Secrets.Local.Provider != "" && !validLocal[cfg.Secrets.Local.Provider] {
			errs = append(errs, fmt.Sprintf("secrets.local.provider must be one of: pass, keychain, credential-manager, auto (got %q)", cfg.Secrets.Local.Provider))
		}
	}

	if len(errs) > 0 {
		return &errors.ValidationError{
			Field:   "config",
			Message: "configuration validation failed",
			Errors:  errs,
		}
	}

	return nil
}

// ValidateProject validates a Project struct against business rules.
func ValidateProject(proj *Project) error {
	var errs []string

	// project name is required
	if proj.Project == "" {
		errs = append(errs, "project is required")
	} else {
		// Validate project name format
		if !isValidName(proj.Project) {
			errs = append(errs, fmt.Sprintf("project name must match pattern ^[a-z][a-z0-9-]*$ (got %q)", proj.Project))
		}
	}

	// environments is required and must have at least one
	if len(proj.Environments) == 0 {
		errs = append(errs, "at least one environment is required")
	}
	for name, env := range proj.Environments {
		if env.Cluster == "" {
			errs = append(errs, fmt.Sprintf("environments.%s.cluster is required", name))
		}
		if env.Secrets == "" {
			errs = append(errs, fmt.Sprintf("environments.%s.secrets is required", name))
		}
	}

	// services is required and must have at least one
	if len(proj.Services) == 0 {
		errs = append(errs, "at least one service is required")
	}

	// Check for duplicate service names
	serviceNames := make(map[string]bool)
	for i, svc := range proj.Services {
		if svc.Name == "" {
			errs = append(errs, fmt.Sprintf("services[%d].name is required", i))
		} else {
			if serviceNames[svc.Name] {
				errs = append(errs, fmt.Sprintf("duplicate service name: %q", svc.Name))
			}
			serviceNames[svc.Name] = true

			if !isValidName(svc.Name) {
				errs = append(errs, fmt.Sprintf("services[%d].name must match pattern ^[a-z][a-z0-9-]*$ (got %q)", i, svc.Name))
			}
		}
		if svc.Pack == "" {
			errs = append(errs, fmt.Sprintf("services[%d].pack is required", i))
		}
	}

	if len(errs) > 0 {
		return &errors.ValidationError{
			Field:   "project",
			Message: "project validation failed",
			Errors:  errs,
		}
	}

	return nil
}

// isValidName checks if a name matches the pattern ^[a-z][a-z0-9-]*$
func isValidName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// First char must be lowercase letter
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}
	// Remaining chars must be lowercase letter, digit, or hyphen
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// ConfigToJSON converts a Config to JSON for display
func ConfigToJSON(cfg *Config) ([]byte, error) {
	return json.MarshalIndent(cfg, "", "  ")
}

// ProjectToJSON converts a Project to JSON for display
func ProjectToJSON(proj *Project) ([]byte, error) {
	return json.MarshalIndent(proj, "", "  ")
}

// ConfigToYAML converts a Config to YAML for display
func ConfigToYAML(cfg *Config) (string, error) {
	// Use a custom approach to avoid yaml.v3 producing non-standard output
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("container: %s\n", cfg.Container))

	if cfg.VPN != nil {
		sb.WriteString("vpn:\n")
		sb.WriteString(fmt.Sprintf("  provider: %s\n", cfg.VPN.Provider))
		if cfg.VPN.ConfigPath != "" {
			sb.WriteString(fmt.Sprintf("  configPath: %s\n", cfg.VPN.ConfigPath))
		}
	}

	if cfg.Hosts != nil {
		sb.WriteString("hosts:\n")
		sb.WriteString(fmt.Sprintf("  mode: %s\n", cfg.Hosts.Mode))
		if cfg.Hosts.Suffix != "" {
			sb.WriteString(fmt.Sprintf("  suffix: %s\n", cfg.Hosts.Suffix))
		}
	}

	if cfg.Network != nil {
		sb.WriteString("network:\n")
		sb.WriteString(fmt.Sprintf("  name: %s\n", cfg.Network.Name))
		sb.WriteString(fmt.Sprintf("  cidr: %s\n", cfg.Network.CIDR))
	}

	if cfg.Secrets != nil && cfg.Secrets.Local != nil {
		sb.WriteString("secrets:\n")
		sb.WriteString("  local:\n")
		sb.WriteString(fmt.Sprintf("    provider: %s\n", cfg.Secrets.Local.Provider))
		if cfg.Secrets.Local.Store != "" {
			sb.WriteString(fmt.Sprintf("    store: %s\n", cfg.Secrets.Local.Store))
		}
		sb.WriteString(fmt.Sprintf("    fallback: %t\n", cfg.Secrets.Local.Fallback))
	}

	return sb.String(), nil
}
