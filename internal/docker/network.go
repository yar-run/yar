package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

// NetworkCreate creates a Docker network with the given name and options.
// Returns the network ID on success. If the network already exists, returns
// the existing network's ID (idempotent).
func (c *dockerClient) NetworkCreate(ctx context.Context, name string, opts NetworkCreateOptions) (string, error) {
	// Check if network already exists (idempotent)
	existing, err := c.findNetworkByName(ctx, name)
	if err != nil {
		return "", ErrNetworkCreate(name, err)
	}
	if existing != nil {
		return existing.ID, nil
	}

	// Build IPAM config if subnet specified
	var ipamConfig *network.IPAM
	if opts.Subnet != "" {
		ipamConfig = &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet:  opts.Subnet,
					Gateway: opts.Gateway,
				},
			},
		}
	}

	// Set default driver
	driver := opts.Driver
	if driver == "" {
		driver = "bridge"
	}

	// Create network
	createOpts := network.CreateOptions{
		Driver:     driver,
		IPAM:       ipamConfig,
		Labels:     opts.Labels,
		Internal:   opts.Internal,
		Attachable: opts.Attachable,
	}

	resp, err := c.cli.NetworkCreate(ctx, name, createOpts)
	if err != nil {
		// Check for "already exists" error (race condition)
		if strings.Contains(err.Error(), "already exists") {
			existing, findErr := c.findNetworkByName(ctx, name)
			if findErr != nil {
				return "", ErrNetworkCreate(name, err)
			}
			if existing != nil {
				return existing.ID, nil
			}
		}
		return "", ErrNetworkCreate(name, err)
	}

	return resp.ID, nil
}

// NetworkRemove removes a Docker network by name.
// Returns nil if the network doesn't exist (idempotent).
func (c *dockerClient) NetworkRemove(ctx context.Context, name string) error {
	err := c.cli.NetworkRemove(ctx, name)
	if err != nil {
		// Check if network not found (idempotent)
		if strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "No such network") {
			return nil
		}
		// Check if network is in use
		if strings.Contains(err.Error(), "has active endpoints") {
			// Get network details to find attached containers
			net, inspectErr := c.NetworkInspect(ctx, name)
			if inspectErr == nil && len(net.Containers) > 0 {
				return ErrNetworkInUse(name, net.Containers)
			}
		}
		return ErrNetworkRemove(name, err)
	}
	return nil
}

// NetworkList lists Docker networks with optional filters.
func (c *dockerClient) NetworkList(ctx context.Context, opts NetworkListOptions) ([]Network, error) {
	// Build filters
	filterArgs := filters.NewArgs()
	for key, values := range opts.Filters {
		for _, value := range values {
			filterArgs.Add(key, value)
		}
	}

	// List networks
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, ErrNetworkList(err)
	}

	// Convert to our Network type
	result := make([]Network, len(networks))
	for i, n := range networks {
		result[i] = networkFromDocker(n)
	}

	return result, nil
}

// NetworkInspect returns detailed information about a specific network.
func (c *dockerClient) NetworkInspect(ctx context.Context, name string) (*Network, error) {
	resp, err := c.cli.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "No such network") {
			return nil, ErrNetworkNotFound(name)
		}
		return nil, ErrNetworkInspect(name, err)
	}

	net := networkFromDockerResource(resp)
	return &net, nil
}

// findNetworkByName finds a network by exact name match.
func (c *dockerClient) findNetworkByName(ctx context.Context, name string) (*Network, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", name)

	networks, err := c.cli.NetworkList(ctx, network.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	// Find exact match (filter returns partial matches)
	for _, n := range networks {
		if n.Name == name {
			net := networkFromDocker(n)
			return &net, nil
		}
	}

	return nil, nil
}

// networkFromDocker converts a Docker network summary to our Network type.
func networkFromDocker(n network.Summary) Network {
	var ipam *IPAM
	if n.IPAM.Driver != "" || len(n.IPAM.Config) > 0 {
		ipam = &IPAM{
			Driver: n.IPAM.Driver,
			Config: make([]IPAMConfig, len(n.IPAM.Config)),
		}
		for i, cfg := range n.IPAM.Config {
			ipam.Config[i] = IPAMConfig{
				Subnet:  cfg.Subnet,
				Gateway: cfg.Gateway,
			}
		}
	}

	return Network{
		ID:     n.ID,
		Name:   n.Name,
		Driver: n.Driver,
		Scope:  n.Scope,
		IPAM:   ipam,
		Labels: n.Labels,
	}
}

// networkFromDockerResource converts a Docker network resource to our Network type.
func networkFromDockerResource(n network.Inspect) Network {
	var ipam *IPAM
	if n.IPAM.Driver != "" || len(n.IPAM.Config) > 0 {
		ipam = &IPAM{
			Driver: n.IPAM.Driver,
			Config: make([]IPAMConfig, len(n.IPAM.Config)),
		}
		for i, cfg := range n.IPAM.Config {
			ipam.Config[i] = IPAMConfig{
				Subnet:  cfg.Subnet,
				Gateway: cfg.Gateway,
			}
		}
	}

	// Extract container IDs
	containers := make([]string, 0, len(n.Containers))
	for id := range n.Containers {
		containers = append(containers, id)
	}

	return Network{
		ID:         n.ID,
		Name:       n.Name,
		Driver:     n.Driver,
		Scope:      n.Scope,
		IPAM:       ipam,
		Labels:     n.Labels,
		Containers: containers,
		Created:    n.Created,
	}
}
