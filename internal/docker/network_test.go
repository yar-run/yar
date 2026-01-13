package docker

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNetwork_Struct(t *testing.T) {
	t.Parallel()

	now := time.Now()
	net := Network{
		ID:         "net-123",
		Name:       "test-net",
		Driver:     "bridge",
		Scope:      "local",
		Labels:     map[string]string{"env": "test", "yar.managed": "true"},
		Containers: []string{"container-1", "container-2"},
		Created:    now,
		IPAM: &IPAM{
			Driver: "default",
			Config: []IPAMConfig{
				{Subnet: "172.16.0.0/16", Gateway: "172.16.0.1"},
			},
		},
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"ID":              {got: net.ID, want: "net-123"},
		"Name":            {got: net.Name, want: "test-net"},
		"Driver":          {got: net.Driver, want: "bridge"},
		"Scope":           {got: net.Scope, want: "local"},
		"Labels count":    {got: len(net.Labels), want: 2},
		"Containers":      {got: len(net.Containers), want: 2},
		"Created":         {got: net.Created, want: now},
		"IPAM driver":     {got: net.IPAM.Driver, want: "default"},
		"IPAM config len": {got: len(net.IPAM.Config), want: 1},
		"Subnet":          {got: net.IPAM.Config[0].Subnet, want: "172.16.0.0/16"},
		"Gateway":         {got: net.IPAM.Config[0].Gateway, want: "172.16.0.1"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("Network.%s = %v, want %v", name, tc.got, tc.want)
			}
		})
	}
}

func TestNetwork_NilIPAM(t *testing.T) {
	t.Parallel()

	net := Network{
		ID:     "net-456",
		Name:   "no-ipam-net",
		Driver: "bridge",
		IPAM:   nil,
	}

	if net.IPAM != nil {
		t.Errorf("Network.IPAM = %v, want nil", net.IPAM)
	}
}

func TestIPAM_EmptyConfig(t *testing.T) {
	t.Parallel()

	ipam := IPAM{
		Driver: "default",
		Config: []IPAMConfig{},
	}

	if ipam.Driver != "default" {
		t.Errorf("IPAM.Driver = %q, want %q", ipam.Driver, "default")
	}
	if len(ipam.Config) != 0 {
		t.Errorf("IPAM.Config length = %d, want 0", len(ipam.Config))
	}
}

func TestIPAMConfig_MultipleSubnets(t *testing.T) {
	t.Parallel()

	ipam := IPAM{
		Driver: "default",
		Config: []IPAMConfig{
			{Subnet: "172.16.0.0/16", Gateway: "172.16.0.1"},
			{Subnet: "192.168.0.0/24", Gateway: "192.168.0.1"},
		},
	}

	if len(ipam.Config) != 2 {
		t.Fatalf("IPAM.Config length = %d, want 2", len(ipam.Config))
	}

	tests := map[string]struct {
		idx     int
		subnet  string
		gateway string
	}{
		"first subnet": {
			idx:     0,
			subnet:  "172.16.0.0/16",
			gateway: "172.16.0.1",
		},
		"second subnet": {
			idx:     1,
			subnet:  "192.168.0.0/24",
			gateway: "192.168.0.1",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := ipam.Config[tc.idx]
			if cfg.Subnet != tc.subnet {
				t.Errorf("IPAMConfig[%d].Subnet = %q, want %q", tc.idx, cfg.Subnet, tc.subnet)
			}
			if cfg.Gateway != tc.gateway {
				t.Errorf("IPAMConfig[%d].Gateway = %q, want %q", tc.idx, cfg.Gateway, tc.gateway)
			}
		})
	}
}

func TestNetworkCreateOptions_AllFields(t *testing.T) {
	t.Parallel()

	opts := NetworkCreateOptions{
		Driver:     "overlay",
		Subnet:     "10.0.0.0/8",
		Gateway:    "10.0.0.1",
		Labels:     map[string]string{"yar.managed": "true", "env": "prod"},
		Internal:   true,
		Attachable: true,
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"Driver":     {got: opts.Driver, want: "overlay"},
		"Subnet":     {got: opts.Subnet, want: "10.0.0.0/8"},
		"Gateway":    {got: opts.Gateway, want: "10.0.0.1"},
		"Internal":   {got: opts.Internal, want: true},
		"Attachable": {got: opts.Attachable, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("NetworkCreateOptions.%s = %v, want %v", name, tc.got, tc.want)
			}
		})
	}

	// Labels are a map, test separately
	if diff := cmp.Diff(map[string]string{"yar.managed": "true", "env": "prod"}, opts.Labels); diff != "" {
		t.Errorf("NetworkCreateOptions.Labels mismatch (-want +got):\n%s", diff)
	}
}

func TestNetworkCreateOptions_Defaults(t *testing.T) {
	t.Parallel()

	// Zero-value options
	opts := NetworkCreateOptions{}

	if opts.Driver != "" {
		t.Errorf("NetworkCreateOptions.Driver default = %q, want empty", opts.Driver)
	}
	if opts.Subnet != "" {
		t.Errorf("NetworkCreateOptions.Subnet default = %q, want empty", opts.Subnet)
	}
	if opts.Internal {
		t.Error("NetworkCreateOptions.Internal default = true, want false")
	}
	if opts.Attachable {
		t.Error("NetworkCreateOptions.Attachable default = true, want false")
	}
	if opts.Labels != nil {
		t.Errorf("NetworkCreateOptions.Labels default = %v, want nil", opts.Labels)
	}
}

func TestNetworkListOptions_Filters(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filters map[string][]string
	}{
		"empty filters": {
			filters: map[string][]string{},
		},
		"single label filter": {
			filters: map[string][]string{
				"label": {"yar.managed=true"},
			},
		},
		"multiple label filters": {
			filters: map[string][]string{
				"label": {"yar.managed=true", "env=prod"},
			},
		},
		"mixed filters": {
			filters: map[string][]string{
				"label":  {"yar.managed=true"},
				"driver": {"bridge"},
				"name":   {"my-network"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := NetworkListOptions{Filters: tc.filters}

			if diff := cmp.Diff(tc.filters, opts.Filters); diff != "" {
				t.Errorf("NetworkListOptions.Filters mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNetworkListOptions_NilFilters(t *testing.T) {
	t.Parallel()

	opts := NetworkListOptions{}

	if opts.Filters != nil {
		t.Errorf("NetworkListOptions.Filters default = %v, want nil", opts.Filters)
	}
}

// Note: The following functions (NetworkCreate, NetworkRemove, NetworkList,
// NetworkInspect) are implemented on dockerClient which requires a Docker
// daemon. Their core logic is tested via MockClient in mock_test.go.
// Integration tests with a real Docker daemon would go in a separate
// docker_integration_test.go file with a build tag.

func TestNetworkComparison_UsingCmpDiff(t *testing.T) {
	t.Parallel()

	// Demonstrate cmp.Diff usage for comparing Network structs
	want := Network{
		ID:     "net-123",
		Name:   "test-net",
		Driver: "bridge",
		Scope:  "local",
		IPAM: &IPAM{
			Driver: "default",
			Config: []IPAMConfig{
				{Subnet: "172.16.0.0/16", Gateway: "172.16.0.1"},
			},
		},
		Labels:     map[string]string{"env": "test"},
		Containers: []string{"c1", "c2"},
	}

	got := Network{
		ID:     "net-123",
		Name:   "test-net",
		Driver: "bridge",
		Scope:  "local",
		IPAM: &IPAM{
			Driver: "default",
			Config: []IPAMConfig{
				{Subnet: "172.16.0.0/16", Gateway: "172.16.0.1"},
			},
		},
		Labels:     map[string]string{"env": "test"},
		Containers: []string{"c1", "c2"},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Network mismatch (-want +got):\n%s", diff)
	}
}

func TestNetworkComparison_DetectsDifferences(t *testing.T) {
	t.Parallel()

	base := Network{
		ID:     "net-123",
		Name:   "test-net",
		Driver: "bridge",
	}

	different := Network{
		ID:     "net-456", // Different ID
		Name:   "test-net",
		Driver: "bridge",
	}

	diff := cmp.Diff(base, different)
	if diff == "" {
		t.Error("cmp.Diff should detect difference between networks with different IDs")
	}
}
