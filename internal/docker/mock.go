package docker

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"
)

// MockClient is a mock implementation of Client for testing.
type MockClient struct {
	mu sync.Mutex

	// Mock responses
	PingError              error
	CloseError             error
	NetworkCreateID        string
	NetworkCreateError     error
	NetworkRemoveError     error
	NetworkListResult      []Network
	NetworkListError       error
	NetworkInspectResult   *Network
	NetworkInspectError    error
	ContainerCreateID      string
	ContainerCreateError   error
	ContainerStartError    error
	ContainerStopError     error
	ContainerRemoveError   error
	ContainerInspectResult *Container
	ContainerInspectError  error
	ContainerListResult    []Container
	ContainerListError     error
	ContainerLogsResult    io.ReadCloser
	ContainerLogsError     error
	ContainerWaitResult    ContainerWaitResult
	ContainerWaitError     error

	// Track calls
	PingCalls             int
	CloseCalls            int
	NetworkCreateCalls    []NetworkCreateCall
	NetworkRemoveCalls    []string
	NetworkListCalls      []NetworkListOptions
	NetworkInspectCalls   []string
	ContainerCreateCalls  []ContainerConfig
	ContainerStartCalls   []string
	ContainerStopCalls    []ContainerStopCall
	ContainerRemoveCalls  []ContainerRemoveCall
	ContainerInspectCalls []string
	ContainerListCalls    []ContainerListOptions
	ContainerLogsCalls    []ContainerLogsCall
	ContainerWaitCalls    []ContainerWaitCall

	// Behavior callbacks (for complex scenarios)
	OnNetworkCreate    func(ctx context.Context, name string, opts NetworkCreateOptions) (string, error)
	OnNetworkRemove    func(ctx context.Context, name string) error
	OnNetworkList      func(ctx context.Context, opts NetworkListOptions) ([]Network, error)
	OnNetworkInspect   func(ctx context.Context, name string) (*Network, error)
	OnContainerCreate  func(ctx context.Context, config ContainerConfig) (string, error)
	OnContainerStart   func(ctx context.Context, id string) error
	OnContainerStop    func(ctx context.Context, id string, timeout *time.Duration) error
	OnContainerRemove  func(ctx context.Context, id string, opts ContainerRemoveOptions) error
	OnContainerInspect func(ctx context.Context, id string) (*Container, error)
	OnContainerList    func(ctx context.Context, opts ContainerListOptions) ([]Container, error)
	OnContainerLogs    func(ctx context.Context, id string, opts ContainerLogOptions) (io.ReadCloser, error)
	OnContainerWait    func(ctx context.Context, id string, condition WaitCondition) (ContainerWaitResult, error)
}

// NetworkCreateCall records a NetworkCreate call.
type NetworkCreateCall struct {
	Name string
	Opts NetworkCreateOptions
}

// ContainerStopCall records a ContainerStop call.
type ContainerStopCall struct {
	ID      string
	Timeout *time.Duration
}

// ContainerRemoveCall records a ContainerRemove call.
type ContainerRemoveCall struct {
	ID   string
	Opts ContainerRemoveOptions
}

// ContainerLogsCall records a ContainerLogs call.
type ContainerLogsCall struct {
	ID   string
	Opts ContainerLogOptions
}

// ContainerWaitCall records a ContainerWait call.
type ContainerWaitCall struct {
	ID        string
	Condition WaitCondition
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

// ContainerCreate implements Client.ContainerCreate.
func (m *MockClient) ContainerCreate(ctx context.Context, config ContainerConfig) (string, error) {
	m.mu.Lock()
	m.ContainerCreateCalls = append(m.ContainerCreateCalls, config)
	callback := m.OnContainerCreate
	createError := m.ContainerCreateError
	createID := m.ContainerCreateID
	m.mu.Unlock()

	if callback != nil {
		return callback(ctx, config)
	}
	if createError != nil {
		return "", createError
	}
	if createID != "" {
		return createID, nil
	}
	return "mock-container-id-" + config.Name, nil
}

// ContainerStart implements Client.ContainerStart.
func (m *MockClient) ContainerStart(ctx context.Context, id string) error {
	m.mu.Lock()
	m.ContainerStartCalls = append(m.ContainerStartCalls, id)
	callback := m.OnContainerStart
	startError := m.ContainerStartError
	m.mu.Unlock()

	var err error
	if callback != nil {
		err = callback(ctx, id)
	} else {
		err = startError
	}
	if containerNotModified(err) {
		return nil
	}
	return err
}

// ContainerStop implements Client.ContainerStop.
func (m *MockClient) ContainerStop(ctx context.Context, id string, timeout *time.Duration) error {
	m.mu.Lock()
	recordedTimeout := cloneDuration(timeout)
	m.ContainerStopCalls = append(m.ContainerStopCalls, ContainerStopCall{ID: id, Timeout: recordedTimeout})
	callback := m.OnContainerStop
	stopError := m.ContainerStopError
	m.mu.Unlock()

	var err error
	if callback != nil {
		err = callback(ctx, id, recordedTimeout)
	} else {
		err = stopError
	}
	if containerNotModified(err) {
		return nil
	}
	return err
}

// ContainerRemove implements Client.ContainerRemove.
func (m *MockClient) ContainerRemove(ctx context.Context, id string, opts ContainerRemoveOptions) error {
	m.mu.Lock()
	m.ContainerRemoveCalls = append(m.ContainerRemoveCalls, ContainerRemoveCall{ID: id, Opts: opts})
	callback := m.OnContainerRemove
	removeError := m.ContainerRemoveError
	m.mu.Unlock()

	if callback != nil {
		return callback(ctx, id, opts)
	}
	return removeError
}

// ContainerInspect implements Client.ContainerInspect.
func (m *MockClient) ContainerInspect(ctx context.Context, id string) (*Container, error) {
	m.mu.Lock()
	m.ContainerInspectCalls = append(m.ContainerInspectCalls, id)
	callback := m.OnContainerInspect
	inspectError := m.ContainerInspectError
	inspectResult := m.ContainerInspectResult
	m.mu.Unlock()

	if callback != nil {
		return callback(ctx, id)
	}
	if inspectError != nil {
		return nil, inspectError
	}
	return inspectResult, nil
}

// ContainerList implements Client.ContainerList.
func (m *MockClient) ContainerList(ctx context.Context, opts ContainerListOptions) ([]Container, error) {
	m.mu.Lock()
	m.ContainerListCalls = append(m.ContainerListCalls, opts)
	callback := m.OnContainerList
	listError := m.ContainerListError
	listResult := m.ContainerListResult
	m.mu.Unlock()

	if callback != nil {
		return callback(ctx, opts)
	}
	if listError != nil {
		return nil, listError
	}
	return listResult, nil
}

// ContainerLogs implements Client.ContainerLogs.
func (m *MockClient) ContainerLogs(ctx context.Context, id string, opts ContainerLogOptions) (io.ReadCloser, error) {
	m.mu.Lock()
	m.ContainerLogsCalls = append(m.ContainerLogsCalls, ContainerLogsCall{ID: id, Opts: opts})
	callback := m.OnContainerLogs
	logsError := m.ContainerLogsError
	logsResult := m.ContainerLogsResult
	m.mu.Unlock()

	if callback != nil {
		return callback(ctx, id, opts)
	}
	if logsError != nil {
		return nil, logsError
	}
	if logsResult != nil {
		return logsResult, nil
	}
	return io.NopCloser(strings.NewReader("")), nil
}

// ContainerWait implements Client.ContainerWait.
func (m *MockClient) ContainerWait(ctx context.Context, id string, condition WaitCondition) (<-chan ContainerWaitResult, <-chan error) {
	m.mu.Lock()
	m.ContainerWaitCalls = append(m.ContainerWaitCalls, ContainerWaitCall{ID: id, Condition: condition})
	result := m.ContainerWaitResult
	err := m.ContainerWaitError
	callback := m.OnContainerWait
	m.mu.Unlock()

	if callback != nil {
		result, err = callback(ctx, id, condition)
	}

	resultC := make(chan ContainerWaitResult, 1)
	errC := make(chan error, 1)
	if err != nil {
		errC <- err
	} else {
		resultC <- result
	}
	return resultC, errC
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
	m.ContainerCreateCalls = nil
	m.ContainerStartCalls = nil
	m.ContainerStopCalls = nil
	m.ContainerRemoveCalls = nil
	m.ContainerInspectCalls = nil
	m.ContainerListCalls = nil
	m.ContainerLogsCalls = nil
	m.ContainerWaitCalls = nil
}

// Ensure MockClient implements Client.
var _ Client = (*MockClient)(nil)
