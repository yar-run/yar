package docker

import (
	"context"
	"sync"
)

// MockClient is a mock implementation of Client for testing.
type MockClient struct {
	mu sync.Mutex

	// Mock responses
	PingError            error
	CloseError           error
	NetworkCreateID      string
	NetworkCreateError   error
	NetworkRemoveError   error
	NetworkListResult    []Network
	NetworkListError     error
	NetworkInspectResult *Network
	NetworkInspectError  error

	// Track calls
	PingCalls           int
	CloseCalls          int
	NetworkCreateCalls  []NetworkCreateCall
	NetworkRemoveCalls  []string
	NetworkListCalls    []NetworkListOptions
	NetworkInspectCalls []string

	// Behavior callbacks (for complex scenarios)
	OnNetworkCreate  func(ctx context.Context, name string, opts NetworkCreateOptions) (string, error)
	OnNetworkRemove  func(ctx context.Context, name string) error
	OnNetworkList    func(ctx context.Context, opts NetworkListOptions) ([]Network, error)
	OnNetworkInspect func(ctx context.Context, name string) (*Network, error)
}

// NetworkCreateCall records a NetworkCreate call.
type NetworkCreateCall struct {
	Name string
	Opts NetworkCreateOptions
}

// NewMockClient creates a new MockClient.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Ping implements Client.Ping.
func (m *MockClient) Ping(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PingCalls++
	return m.PingError
}

// Close implements Client.Close.
func (m *MockClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CloseCalls++
	return m.CloseError
}

// NetworkCreate implements Client.NetworkCreate.
func (m *MockClient) NetworkCreate(ctx context.Context, name string, opts NetworkCreateOptions) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NetworkCreateCalls = append(m.NetworkCreateCalls, NetworkCreateCall{Name: name, Opts: opts})

	if m.OnNetworkCreate != nil {
		return m.OnNetworkCreate(ctx, name, opts)
	}

	if m.NetworkCreateError != nil {
		return "", m.NetworkCreateError
	}

	if m.NetworkCreateID != "" {
		return m.NetworkCreateID, nil
	}

	// Default: return a generated ID
	return "mock-network-id-" + name, nil
}

// NetworkRemove implements Client.NetworkRemove.
func (m *MockClient) NetworkRemove(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NetworkRemoveCalls = append(m.NetworkRemoveCalls, name)

	if m.OnNetworkRemove != nil {
		return m.OnNetworkRemove(ctx, name)
	}

	return m.NetworkRemoveError
}

// NetworkList implements Client.NetworkList.
func (m *MockClient) NetworkList(ctx context.Context, opts NetworkListOptions) ([]Network, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NetworkListCalls = append(m.NetworkListCalls, opts)

	if m.OnNetworkList != nil {
		return m.OnNetworkList(ctx, opts)
	}

	if m.NetworkListError != nil {
		return nil, m.NetworkListError
	}

	return m.NetworkListResult, nil
}

// NetworkInspect implements Client.NetworkInspect.
func (m *MockClient) NetworkInspect(ctx context.Context, name string) (*Network, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NetworkInspectCalls = append(m.NetworkInspectCalls, name)

	if m.OnNetworkInspect != nil {
		return m.OnNetworkInspect(ctx, name)
	}

	if m.NetworkInspectError != nil {
		return nil, m.NetworkInspectError
	}

	return m.NetworkInspectResult, nil
}

// Reset clears all recorded calls and resets mock state.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PingCalls = 0
	m.CloseCalls = 0
	m.NetworkCreateCalls = nil
	m.NetworkRemoveCalls = nil
	m.NetworkListCalls = nil
	m.NetworkInspectCalls = nil
}

// Ensure MockClient implements Client.
var _ Client = (*MockClient)(nil)
