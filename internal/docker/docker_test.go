package docker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockClient_ImplementsClient(t *testing.T) {
	// This test verifies at compile time that MockClient implements Client
	var _ Client = (*MockClient)(nil)
}

func TestMockClient_Ping(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Test successful ping
	if err := mock.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
	if mock.PingCalls != 1 {
		t.Errorf("PingCalls = %d, want 1", mock.PingCalls)
	}

	// Test ping with error
	mock.Reset()
	mock.PingError = errors.New("connection refused")
	if err := mock.Ping(ctx); err == nil {
		t.Error("Ping() error = nil, want error")
	}
}

func TestMockClient_NetworkCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("default ID", func(t *testing.T) {
		mock := NewMockClient()
		id, err := mock.NetworkCreate(ctx, "test-net", NetworkCreateOptions{})
		if err != nil {
			t.Fatalf("NetworkCreate() error = %v", err)
		}
		if id != "mock-network-id-test-net" {
			t.Errorf("NetworkCreate() = %q, want %q", id, "mock-network-id-test-net")
		}
		if len(mock.NetworkCreateCalls) != 1 {
			t.Errorf("NetworkCreateCalls = %d, want 1", len(mock.NetworkCreateCalls))
		}
		if mock.NetworkCreateCalls[0].Name != "test-net" {
			t.Errorf("NetworkCreateCalls[0].Name = %q, want %q", mock.NetworkCreateCalls[0].Name, "test-net")
		}
	})

	t.Run("custom ID", func(t *testing.T) {
		mock := NewMockClient()
		mock.NetworkCreateID = "custom-id-123"
		id, err := mock.NetworkCreate(ctx, "test-net", NetworkCreateOptions{})
		if err != nil {
			t.Fatalf("NetworkCreate() error = %v", err)
		}
		if id != "custom-id-123" {
			t.Errorf("NetworkCreate() = %q, want %q", id, "custom-id-123")
		}
	})

	t.Run("with error", func(t *testing.T) {
		mock := NewMockClient()
		mock.NetworkCreateError = errors.New("network error")
		_, err := mock.NetworkCreate(ctx, "test-net", NetworkCreateOptions{})
		if err == nil {
			t.Error("NetworkCreate() error = nil, want error")
		}
	})

	t.Run("with options", func(t *testing.T) {
		mock := NewMockClient()
		opts := NetworkCreateOptions{
			Driver:     "bridge",
			Subnet:     "172.16.34.0/23",
			Gateway:    "172.16.34.1",
			Labels:     map[string]string{"yar.managed": "true"},
			Internal:   true,
			Attachable: true,
		}
		_, err := mock.NetworkCreate(ctx, "test-net", opts)
		if err != nil {
			t.Fatalf("NetworkCreate() error = %v", err)
		}
		if mock.NetworkCreateCalls[0].Opts.Subnet != "172.16.34.0/23" {
			t.Errorf("Subnet = %q, want %q", mock.NetworkCreateCalls[0].Opts.Subnet, "172.16.34.0/23")
		}
		if mock.NetworkCreateCalls[0].Opts.Labels["yar.managed"] != "true" {
			t.Errorf("Labels = %v, want yar.managed=true", mock.NetworkCreateCalls[0].Opts.Labels)
		}
	})

	t.Run("with callback", func(t *testing.T) {
		mock := NewMockClient()
		mock.OnNetworkCreate = func(ctx context.Context, name string, opts NetworkCreateOptions) (string, error) {
			return "callback-id-" + name, nil
		}
		id, err := mock.NetworkCreate(ctx, "test-net", NetworkCreateOptions{})
		if err != nil {
			t.Fatalf("NetworkCreate() error = %v", err)
		}
		if id != "callback-id-test-net" {
			t.Errorf("NetworkCreate() = %q, want %q", id, "callback-id-test-net")
		}
	})
}

func TestMockClient_NetworkRemove(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock := NewMockClient()
		err := mock.NetworkRemove(ctx, "test-net")
		if err != nil {
			t.Errorf("NetworkRemove() error = %v, want nil", err)
		}
		if len(mock.NetworkRemoveCalls) != 1 {
			t.Errorf("NetworkRemoveCalls = %d, want 1", len(mock.NetworkRemoveCalls))
		}
		if mock.NetworkRemoveCalls[0] != "test-net" {
			t.Errorf("NetworkRemoveCalls[0] = %q, want %q", mock.NetworkRemoveCalls[0], "test-net")
		}
	})

	t.Run("with error", func(t *testing.T) {
		mock := NewMockClient()
		mock.NetworkRemoveError = errors.New("network in use")
		err := mock.NetworkRemove(ctx, "test-net")
		if err == nil {
			t.Error("NetworkRemove() error = nil, want error")
		}
	})
}

func TestMockClient_NetworkList(t *testing.T) {
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		mock := NewMockClient()
		networks, err := mock.NetworkList(ctx, NetworkListOptions{})
		if err != nil {
			t.Fatalf("NetworkList() error = %v", err)
		}
		if len(networks) != 0 {
			t.Errorf("NetworkList() = %d networks, want 0", len(networks))
		}
	})

	t.Run("with results", func(t *testing.T) {
		mock := NewMockClient()
		mock.NetworkListResult = []Network{
			{ID: "id1", Name: "net1", Driver: "bridge"},
			{ID: "id2", Name: "net2", Driver: "bridge"},
		}
		networks, err := mock.NetworkList(ctx, NetworkListOptions{})
		if err != nil {
			t.Fatalf("NetworkList() error = %v", err)
		}
		if len(networks) != 2 {
			t.Errorf("NetworkList() = %d networks, want 2", len(networks))
		}
		if networks[0].Name != "net1" {
			t.Errorf("networks[0].Name = %q, want %q", networks[0].Name, "net1")
		}
	})

	t.Run("with filters", func(t *testing.T) {
		mock := NewMockClient()
		opts := NetworkListOptions{
			Filters: map[string][]string{
				"label": {"yar.managed=true"},
			},
		}
		_, err := mock.NetworkList(ctx, opts)
		if err != nil {
			t.Fatalf("NetworkList() error = %v", err)
		}
		if len(mock.NetworkListCalls) != 1 {
			t.Errorf("NetworkListCalls = %d, want 1", len(mock.NetworkListCalls))
		}
		if mock.NetworkListCalls[0].Filters["label"][0] != "yar.managed=true" {
			t.Errorf("Filters = %v, want label=yar.managed=true", mock.NetworkListCalls[0].Filters)
		}
	})
}

func TestMockClient_NetworkInspect(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		mock := NewMockClient()
		mock.NetworkInspectResult = &Network{
			ID:     "net-123",
			Name:   "test-net",
			Driver: "bridge",
			Scope:  "local",
			IPAM: &IPAM{
				Driver: "default",
				Config: []IPAMConfig{
					{Subnet: "172.16.34.0/23", Gateway: "172.16.34.1"},
				},
			},
		}
		net, err := mock.NetworkInspect(ctx, "test-net")
		if err != nil {
			t.Fatalf("NetworkInspect() error = %v", err)
		}
		if net.ID != "net-123" {
			t.Errorf("net.ID = %q, want %q", net.ID, "net-123")
		}
		if net.IPAM == nil {
			t.Fatal("net.IPAM = nil, want IPAM config")
		}
		if net.IPAM.Config[0].Subnet != "172.16.34.0/23" {
			t.Errorf("Subnet = %q, want %q", net.IPAM.Config[0].Subnet, "172.16.34.0/23")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock := NewMockClient()
		mock.NetworkInspectError = ErrNetworkNotFound("test-net")
		_, err := mock.NetworkInspect(ctx, "test-net")
		if err == nil {
			t.Error("NetworkInspect() error = nil, want error")
		}
		var dockerErr *DockerError
		if !errors.As(err, &dockerErr) {
			t.Errorf("error type = %T, want *DockerError", err)
		}
	})
}

func TestMockClient_Reset(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Make some calls
	mock.Ping(ctx)
	mock.NetworkCreate(ctx, "net1", NetworkCreateOptions{})
	mock.NetworkRemove(ctx, "net1")
	mock.NetworkList(ctx, NetworkListOptions{})
	mock.NetworkInspect(ctx, "net1")

	// Verify calls were recorded
	if mock.PingCalls != 1 {
		t.Errorf("PingCalls = %d, want 1", mock.PingCalls)
	}

	// Reset and verify
	mock.Reset()
	if mock.PingCalls != 0 {
		t.Errorf("After reset: PingCalls = %d, want 0", mock.PingCalls)
	}
	if len(mock.NetworkCreateCalls) != 0 {
		t.Errorf("After reset: NetworkCreateCalls = %d, want 0", len(mock.NetworkCreateCalls))
	}
}

func TestDockerError(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		underlying := errors.New("connection refused")
		err := ErrDaemonConnection(underlying)

		if err.Op != "connect" {
			t.Errorf("Op = %q, want %q", err.Op, "connect")
		}

		errStr := err.Error()
		if errStr == "" {
			t.Error("Error() returned empty string")
		}

		// Test Unwrap
		if !errors.Is(err, underlying) {
			t.Error("errors.Is failed to match underlying error")
		}
	})

	t.Run("without underlying error", func(t *testing.T) {
		err := ErrNetworkNotFound("test-net")
		errStr := err.Error()
		if errStr == "" {
			t.Error("Error() returned empty string")
		}
		if err.Unwrap() != nil {
			t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
		}
	})

	t.Run("network in use", func(t *testing.T) {
		containers := []string{"container1", "container2"}
		err := ErrNetworkInUse("test-net", containers)
		if err.Op != "network.remove" {
			t.Errorf("Op = %q, want %q", err.Op, "network.remove")
		}
	})
}

func TestNetwork_Struct(t *testing.T) {
	now := time.Now()
	net := Network{
		ID:         "net-123",
		Name:       "test-net",
		Driver:     "bridge",
		Scope:      "local",
		Labels:     map[string]string{"env": "test"},
		Containers: []string{"c1", "c2"},
		Created:    now,
		IPAM: &IPAM{
			Driver: "default",
			Config: []IPAMConfig{
				{Subnet: "172.16.0.0/16", Gateway: "172.16.0.1"},
			},
		},
	}

	if net.ID != "net-123" {
		t.Errorf("ID = %q, want %q", net.ID, "net-123")
	}
	if len(net.Containers) != 2 {
		t.Errorf("Containers = %d, want 2", len(net.Containers))
	}
	if net.IPAM.Config[0].Subnet != "172.16.0.0/16" {
		t.Errorf("Subnet = %q, want %q", net.IPAM.Config[0].Subnet, "172.16.0.0/16")
	}
}

func TestNetworkCreateOptions(t *testing.T) {
	opts := NetworkCreateOptions{
		Driver:     "bridge",
		Subnet:     "10.0.0.0/8",
		Gateway:    "10.0.0.1",
		Labels:     map[string]string{"yar.managed": "true"},
		Internal:   true,
		Attachable: true,
	}

	if opts.Driver != "bridge" {
		t.Errorf("Driver = %q, want %q", opts.Driver, "bridge")
	}
	if opts.Subnet != "10.0.0.0/8" {
		t.Errorf("Subnet = %q, want %q", opts.Subnet, "10.0.0.0/8")
	}
	if !opts.Internal {
		t.Error("Internal = false, want true")
	}
}
