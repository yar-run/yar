package docker

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMockClient_ImplementsClient(t *testing.T) {
	t.Parallel()

	// Compile-time assertion that MockClient implements Client interface.
	var _ Client = (*MockClient)(nil)
}

func TestMockClient_Ping(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock func(*MockClient)
		wantErr   bool
		wantCalls int
	}{
		"success": {
			setupMock: func(m *MockClient) {},
			wantErr:   false,
			wantCalls: 1,
		},
		"returns configured error": {
			setupMock: func(m *MockClient) {
				m.PingError = errors.New("connection refused")
			},
			wantErr:   true,
			wantCalls: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			err := mock.Ping(context.Background())

			if (err != nil) != tc.wantErr {
				t.Errorf("Ping() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if mock.PingCalls != tc.wantCalls {
				t.Errorf("Ping() calls = %d, want %d", mock.PingCalls, tc.wantCalls)
			}
		})
	}
}

func TestMockClient_Close(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock func(*MockClient)
		wantErr   bool
		wantCalls int
	}{
		"success": {
			setupMock: func(m *MockClient) {},
			wantErr:   false,
			wantCalls: 1,
		},
		"returns configured error": {
			setupMock: func(m *MockClient) {
				m.CloseError = errors.New("close failed")
			},
			wantErr:   true,
			wantCalls: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			err := mock.Close()

			if (err != nil) != tc.wantErr {
				t.Errorf("Close() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if mock.CloseCalls != tc.wantCalls {
				t.Errorf("Close() calls = %d, want %d", mock.CloseCalls, tc.wantCalls)
			}
		})
	}
}

func TestMockClient_NetworkCreate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock   func(*MockClient)
		name        string
		opts        NetworkCreateOptions
		wantID      string
		wantErr     bool
		wantRecords int
	}{
		"returns default generated ID": {
			setupMock:   func(m *MockClient) {},
			name:        "test-net",
			opts:        NetworkCreateOptions{},
			wantID:      "mock-network-id-test-net",
			wantErr:     false,
			wantRecords: 1,
		},
		"returns custom configured ID": {
			setupMock: func(m *MockClient) {
				m.NetworkCreateID = "custom-id-123"
			},
			name:        "test-net",
			opts:        NetworkCreateOptions{},
			wantID:      "custom-id-123",
			wantErr:     false,
			wantRecords: 1,
		},
		"returns configured error": {
			setupMock: func(m *MockClient) {
				m.NetworkCreateError = errors.New("network error")
			},
			name:        "test-net",
			opts:        NetworkCreateOptions{},
			wantID:      "",
			wantErr:     true,
			wantRecords: 1,
		},
		"uses callback when provided": {
			setupMock: func(m *MockClient) {
				m.OnNetworkCreate = func(ctx context.Context, name string, opts NetworkCreateOptions) (string, error) {
					return "callback-id-" + name, nil
				}
			},
			name:        "my-net",
			opts:        NetworkCreateOptions{},
			wantID:      "callback-id-my-net",
			wantErr:     false,
			wantRecords: 1,
		},
		"callback returns error": {
			setupMock: func(m *MockClient) {
				m.OnNetworkCreate = func(ctx context.Context, name string, opts NetworkCreateOptions) (string, error) {
					return "", errors.New("callback error")
				}
			},
			name:        "test-net",
			opts:        NetworkCreateOptions{},
			wantID:      "",
			wantErr:     true,
			wantRecords: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			gotID, err := mock.NetworkCreate(context.Background(), tc.name, tc.opts)

			if (err != nil) != tc.wantErr {
				t.Errorf("NetworkCreate(%q) error = %v, wantErr = %v", tc.name, err, tc.wantErr)
			}
			if gotID != tc.wantID {
				t.Errorf("NetworkCreate(%q) = %q, want %q", tc.name, gotID, tc.wantID)
			}
			if len(mock.NetworkCreateCalls) != tc.wantRecords {
				t.Errorf("NetworkCreate(%q) recorded %d calls, want %d", tc.name, len(mock.NetworkCreateCalls), tc.wantRecords)
			}
		})
	}
}

func TestMockClient_NetworkCreate_RecordsCallDetails(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	opts := NetworkCreateOptions{
		Driver:     "bridge",
		Subnet:     "172.16.34.0/23",
		Gateway:    "172.16.34.1",
		Labels:     map[string]string{"yar.managed": "true"},
		Internal:   true,
		Attachable: true,
	}

	_, err := mock.NetworkCreate(context.Background(), "test-net", opts)
	if err != nil {
		t.Fatalf("NetworkCreate() error = %v, want nil", err)
	}

	if len(mock.NetworkCreateCalls) != 1 {
		t.Fatalf("NetworkCreateCalls = %d, want 1", len(mock.NetworkCreateCalls))
	}

	recorded := mock.NetworkCreateCalls[0]

	if recorded.Name != "test-net" {
		t.Errorf("recorded.Name = %q, want %q", recorded.Name, "test-net")
	}
	if diff := cmp.Diff(opts, recorded.Opts); diff != "" {
		t.Errorf("recorded.Opts mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_NetworkRemove(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock   func(*MockClient)
		name        string
		wantErr     bool
		wantRecords int
	}{
		"success": {
			setupMock:   func(m *MockClient) {},
			name:        "test-net",
			wantErr:     false,
			wantRecords: 1,
		},
		"returns configured error": {
			setupMock: func(m *MockClient) {
				m.NetworkRemoveError = errors.New("network in use")
			},
			name:        "test-net",
			wantErr:     true,
			wantRecords: 1,
		},
		"uses callback when provided": {
			setupMock: func(m *MockClient) {
				m.OnNetworkRemove = func(ctx context.Context, name string) error {
					if name == "protected-net" {
						return errors.New("cannot remove protected network")
					}
					return nil
				}
			},
			name:        "protected-net",
			wantErr:     true,
			wantRecords: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			err := mock.NetworkRemove(context.Background(), tc.name)

			if (err != nil) != tc.wantErr {
				t.Errorf("NetworkRemove(%q) error = %v, wantErr = %v", tc.name, err, tc.wantErr)
			}
			if len(mock.NetworkRemoveCalls) != tc.wantRecords {
				t.Errorf("NetworkRemove(%q) recorded %d calls, want %d", tc.name, len(mock.NetworkRemoveCalls), tc.wantRecords)
			}
		})
	}
}

func TestMockClient_NetworkRemove_RecordsNetworkName(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()

	_ = mock.NetworkRemove(context.Background(), "my-network")

	if len(mock.NetworkRemoveCalls) != 1 {
		t.Fatalf("NetworkRemoveCalls = %d, want 1", len(mock.NetworkRemoveCalls))
	}
	if mock.NetworkRemoveCalls[0] != "my-network" {
		t.Errorf("NetworkRemoveCalls[0] = %q, want %q", mock.NetworkRemoveCalls[0], "my-network")
	}
}

func TestMockClient_NetworkList(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock   func(*MockClient)
		opts        NetworkListOptions
		wantLen     int
		wantErr     bool
		wantRecords int
	}{
		"empty list by default": {
			setupMock:   func(m *MockClient) {},
			opts:        NetworkListOptions{},
			wantLen:     0,
			wantErr:     false,
			wantRecords: 1,
		},
		"returns configured results": {
			setupMock: func(m *MockClient) {
				m.NetworkListResult = []Network{
					{ID: "id1", Name: "net1", Driver: "bridge"},
					{ID: "id2", Name: "net2", Driver: "bridge"},
				}
			},
			opts:        NetworkListOptions{},
			wantLen:     2,
			wantErr:     false,
			wantRecords: 1,
		},
		"returns configured error": {
			setupMock: func(m *MockClient) {
				m.NetworkListError = errors.New("list failed")
			},
			opts:        NetworkListOptions{},
			wantLen:     0,
			wantErr:     true,
			wantRecords: 1,
		},
		"uses callback when provided": {
			setupMock: func(m *MockClient) {
				m.OnNetworkList = func(ctx context.Context, opts NetworkListOptions) ([]Network, error) {
					return []Network{{ID: "callback-id", Name: "callback-net"}}, nil
				}
			},
			opts:        NetworkListOptions{},
			wantLen:     1,
			wantErr:     false,
			wantRecords: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			got, err := mock.NetworkList(context.Background(), tc.opts)

			if (err != nil) != tc.wantErr {
				t.Errorf("NetworkList() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if len(got) != tc.wantLen {
				t.Errorf("NetworkList() returned %d networks, want %d", len(got), tc.wantLen)
			}
			if len(mock.NetworkListCalls) != tc.wantRecords {
				t.Errorf("NetworkList() recorded %d calls, want %d", len(mock.NetworkListCalls), tc.wantRecords)
			}
		})
	}
}

func TestMockClient_NetworkList_RecordsFilters(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	opts := NetworkListOptions{
		Filters: map[string][]string{
			"label": {"yar.managed=true"},
		},
	}

	_, _ = mock.NetworkList(context.Background(), opts)

	if len(mock.NetworkListCalls) != 1 {
		t.Fatalf("NetworkListCalls = %d, want 1", len(mock.NetworkListCalls))
	}

	recorded := mock.NetworkListCalls[0]
	if diff := cmp.Diff(opts.Filters, recorded.Filters); diff != "" {
		t.Errorf("recorded.Filters mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_NetworkInspect(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock   func(*MockClient)
		name        string
		wantNetwork *Network
		wantErr     bool
		wantRecords int
	}{
		"returns nil by default": {
			setupMock:   func(m *MockClient) {},
			name:        "test-net",
			wantNetwork: nil,
			wantErr:     false,
			wantRecords: 1,
		},
		"returns configured result": {
			setupMock: func(m *MockClient) {
				m.NetworkInspectResult = &Network{
					ID:     "net-123",
					Name:   "test-net",
					Driver: "bridge",
				}
			},
			name: "test-net",
			wantNetwork: &Network{
				ID:     "net-123",
				Name:   "test-net",
				Driver: "bridge",
			},
			wantErr:     false,
			wantRecords: 1,
		},
		"returns configured error": {
			setupMock: func(m *MockClient) {
				m.NetworkInspectError = ErrNetworkNotFound("test-net")
			},
			name:        "test-net",
			wantNetwork: nil,
			wantErr:     true,
			wantRecords: 1,
		},
		"uses callback when provided": {
			setupMock: func(m *MockClient) {
				m.OnNetworkInspect = func(ctx context.Context, name string) (*Network, error) {
					return &Network{ID: "callback-" + name, Name: name}, nil
				}
			},
			name: "my-net",
			wantNetwork: &Network{
				ID:   "callback-my-net",
				Name: "my-net",
			},
			wantErr:     false,
			wantRecords: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			got, err := mock.NetworkInspect(context.Background(), tc.name)

			if (err != nil) != tc.wantErr {
				t.Errorf("NetworkInspect(%q) error = %v, wantErr = %v", tc.name, err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.wantNetwork, got); diff != "" {
				t.Errorf("NetworkInspect(%q) mismatch (-want +got):\n%s", tc.name, diff)
			}
			if len(mock.NetworkInspectCalls) != tc.wantRecords {
				t.Errorf("NetworkInspect(%q) recorded %d calls, want %d", tc.name, len(mock.NetworkInspectCalls), tc.wantRecords)
			}
		})
	}
}

func TestMockClient_NetworkInspect_RecordsName(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()

	_, _ = mock.NetworkInspect(context.Background(), "queried-net")

	if len(mock.NetworkInspectCalls) != 1 {
		t.Fatalf("NetworkInspectCalls = %d, want 1", len(mock.NetworkInspectCalls))
	}
	if mock.NetworkInspectCalls[0] != "queried-net" {
		t.Errorf("NetworkInspectCalls[0] = %q, want %q", mock.NetworkInspectCalls[0], "queried-net")
	}
}

func TestMockClient_Reset(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	ctx := context.Background()

	// Make calls to populate counters
	_ = mock.Ping(ctx)
	_, _ = mock.NetworkCreate(ctx, "net1", NetworkCreateOptions{})
	_ = mock.NetworkRemove(ctx, "net1")
	_, _ = mock.NetworkList(ctx, NetworkListOptions{})
	_, _ = mock.NetworkInspect(ctx, "net1")

	// Verify calls were recorded
	if mock.PingCalls != 1 {
		t.Errorf("Before reset: PingCalls = %d, want 1", mock.PingCalls)
	}
	if len(mock.NetworkCreateCalls) != 1 {
		t.Errorf("Before reset: NetworkCreateCalls = %d, want 1", len(mock.NetworkCreateCalls))
	}
	if len(mock.NetworkRemoveCalls) != 1 {
		t.Errorf("Before reset: NetworkRemoveCalls = %d, want 1", len(mock.NetworkRemoveCalls))
	}
	if len(mock.NetworkListCalls) != 1 {
		t.Errorf("Before reset: NetworkListCalls = %d, want 1", len(mock.NetworkListCalls))
	}
	if len(mock.NetworkInspectCalls) != 1 {
		t.Errorf("Before reset: NetworkInspectCalls = %d, want 1", len(mock.NetworkInspectCalls))
	}

	// Reset
	mock.Reset()

	// Verify all counters are cleared
	if mock.PingCalls != 0 {
		t.Errorf("After reset: PingCalls = %d, want 0", mock.PingCalls)
	}
	if mock.CloseCalls != 0 {
		t.Errorf("After reset: CloseCalls = %d, want 0", mock.CloseCalls)
	}
	if len(mock.NetworkCreateCalls) != 0 {
		t.Errorf("After reset: NetworkCreateCalls = %d, want 0", len(mock.NetworkCreateCalls))
	}
	if len(mock.NetworkRemoveCalls) != 0 {
		t.Errorf("After reset: NetworkRemoveCalls = %d, want 0", len(mock.NetworkRemoveCalls))
	}
	if len(mock.NetworkListCalls) != 0 {
		t.Errorf("After reset: NetworkListCalls = %d, want 0", len(mock.NetworkListCalls))
	}
	if len(mock.NetworkInspectCalls) != 0 {
		t.Errorf("After reset: NetworkInspectCalls = %d, want 0", len(mock.NetworkInspectCalls))
	}
}

func TestMockClient_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	ctx := context.Background()

	// Run multiple goroutines accessing the mock concurrently.
	// This test verifies the mutex protects against data races.
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_ = mock.Ping(ctx)
			_, _ = mock.NetworkCreate(ctx, "net", NetworkCreateOptions{})
			_ = mock.NetworkRemove(ctx, "net")
			_, _ = mock.NetworkList(ctx, NetworkListOptions{})
			_, _ = mock.NetworkInspect(ctx, "net")
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counts are consistent
	if mock.PingCalls != 10 {
		t.Errorf("Concurrent PingCalls = %d, want 10", mock.PingCalls)
	}
	if len(mock.NetworkCreateCalls) != 10 {
		t.Errorf("Concurrent NetworkCreateCalls = %d, want 10", len(mock.NetworkCreateCalls))
	}
}
