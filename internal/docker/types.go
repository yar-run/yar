package docker

import "time"

// Network represents a Docker network.
type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	IPAM       *IPAM             `json:"ipam,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Containers []string          `json:"containers,omitempty"` // container IDs
	Created    time.Time         `json:"created"`
}

// IPAM represents IP Address Management configuration.
type IPAM struct {
	Driver string       `json:"driver"`
	Config []IPAMConfig `json:"config,omitempty"`
}

// IPAMConfig represents IPAM pool configuration.
type IPAMConfig struct {
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

// NetworkCreateOptions configures network creation.
type NetworkCreateOptions struct {
	Driver     string            // Network driver (default: "bridge")
	Subnet     string            // CIDR notation (e.g., "172.16.34.0/23")
	Gateway    string            // Gateway IP (optional, derived from subnet)
	Labels     map[string]string // Network labels
	Internal   bool              // Restrict external access
	Attachable bool              // Allow manual container attachment
}

// NetworkListOptions configures network listing.
type NetworkListOptions struct {
	Filters map[string][]string // Filter by name, id, driver, label, etc.
}
