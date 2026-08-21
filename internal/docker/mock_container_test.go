package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/google/go-cmp/cmp"
)

func TestMockClient_ContainerCreate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock func(*MockClient)
		config    ContainerConfig
		wantID    string
		wantErr   bool
	}{
		"returns generated ID": {
			setupMock: func(*MockClient) {},
			config:    ContainerConfig{Name: "web", Image: "nginx:latest"},
			wantID:    "mock-container-id-web",
		},
		"returns configured ID": {
			setupMock: func(mock *MockClient) {
				mock.ContainerCreateID = "container-123"
			},
			config: ContainerConfig{Name: "web", Image: "nginx:latest"},
			wantID: "container-123",
		},
		"returns configured error": {
			setupMock: func(mock *MockClient) {
				mock.ContainerCreateError = errors.New("create failed")
			},
			config:  ContainerConfig{Name: "web", Image: "nginx:latest"},
			wantErr: true,
		},
		"uses callback": {
			setupMock: func(mock *MockClient) {
				mock.OnContainerCreate = func(_ context.Context, config ContainerConfig) (string, error) {
					return "callback-" + config.Name, nil
				}
			},
			config: ContainerConfig{Name: "api", Image: "nginx:latest"},
			wantID: "callback-api",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)

			got, err := mock.ContainerCreate(context.Background(), tc.config)
			if (err != nil) != tc.wantErr {
				t.Errorf("ContainerCreate() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if got != tc.wantID {
				t.Errorf("ContainerCreate() = %q, want %q", got, tc.wantID)
			}
			if len(mock.ContainerCreateCalls) != 1 {
				t.Errorf("ContainerCreateCalls = %d, want 1", len(mock.ContainerCreateCalls))
			}
		})
	}
}

func TestMockClient_ContainerCreateRecordsConfiguration(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	config := ContainerConfig{
		Name:   "web",
		Image:  "nginx:latest",
		Ports:  []PortMapping{{HostPort: "8080", ContainerPort: "80"}},
		Labels: map[string]string{"app": "web"},
	}

	_, err := mock.ContainerCreate(context.Background(), config)
	if err != nil {
		t.Fatalf("ContainerCreate() error = %v", err)
	}
	if len(mock.ContainerCreateCalls) != 1 {
		t.Fatalf("ContainerCreateCalls = %d, want 1", len(mock.ContainerCreateCalls))
	}
	if diff := cmp.Diff(config, mock.ContainerCreateCalls[0]); diff != "" {
		t.Errorf("ContainerCreateCalls[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_ContainerCallbackCanReenter(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	mock.OnContainerCreate = func(ctx context.Context, config ContainerConfig) (string, error) {
		if err := mock.ContainerStart(ctx, config.Name); err != nil {
			return "", err
		}
		return "container-123", nil
	}

	id, err := mock.ContainerCreate(context.Background(), ContainerConfig{Name: "web", Image: "nginx:latest"})
	if err != nil {
		t.Fatalf("ContainerCreate() error = %v", err)
	}
	if id != "container-123" {
		t.Errorf("ContainerCreate() = %q, want %q", id, "container-123")
	}
	if diff := cmp.Diff([]string{"web"}, mock.ContainerStartCalls); diff != "" {
		t.Errorf("ContainerStartCalls mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_ContainerStart(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock func(*MockClient)
		id        string
		wantErr   bool
	}{
		"success": {
			setupMock: func(*MockClient) {},
			id:        "container-123",
		},
		"returns configured error": {
			setupMock: func(mock *MockClient) {
				mock.ContainerStartError = errors.New("start failed")
			},
			id:      "container-123",
			wantErr: true,
		},
		"is idempotent": {
			setupMock: func(mock *MockClient) {
				mock.ContainerStartError = errdefs.ErrNotModified
			},
			id: "container-123",
		},
		"uses callback": {
			setupMock: func(mock *MockClient) {
				mock.OnContainerStart = func(_ context.Context, id string) error {
					if id != "container-123" {
						return errors.New("unexpected container ID")
					}
					return nil
				}
			},
			id: "container-123",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)
			err := mock.ContainerStart(context.Background(), tc.id)
			if (err != nil) != tc.wantErr {
				t.Errorf("ContainerStart() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if diff := cmp.Diff([]string{tc.id}, mock.ContainerStartCalls); diff != "" {
				t.Errorf("ContainerStartCalls mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockClient_ContainerStop(t *testing.T) {
	t.Parallel()

	timeout := 30 * time.Second
	tests := map[string]struct {
		setupMock func(*MockClient)
		wantErr   bool
	}{
		"success": {
			setupMock: func(*MockClient) {},
		},
		"returns configured error": {
			setupMock: func(mock *MockClient) {
				mock.ContainerStopError = errors.New("stop failed")
			},
			wantErr: true,
		},
		"is idempotent for stopped container": {
			setupMock: func(mock *MockClient) {
				mock.ContainerStopError = errdefs.ErrNotModified
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)
			err := mock.ContainerStop(context.Background(), "container-123", &timeout)
			if (err != nil) != tc.wantErr {
				t.Errorf("ContainerStop() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if len(mock.ContainerStopCalls) != 1 {
				t.Fatalf("ContainerStopCalls = %d, want 1", len(mock.ContainerStopCalls))
			}
			if mock.ContainerStopCalls[0].ID != "container-123" {
				t.Errorf("ContainerStopCalls[0].ID = %q, want %q", mock.ContainerStopCalls[0].ID, "container-123")
			}
			if mock.ContainerStopCalls[0].Timeout == nil || *mock.ContainerStopCalls[0].Timeout != timeout {
				t.Errorf("ContainerStopCalls[0].Timeout = %v, want %v", mock.ContainerStopCalls[0].Timeout, timeout)
			}
		})
	}
}

func TestMockClient_ContainerRemove(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setupMock func(*MockClient)
		opts      ContainerRemoveOptions
		wantErr   bool
	}{
		"success": {
			setupMock: func(*MockClient) {},
			opts:      ContainerRemoveOptions{},
		},
		"force with volumes": {
			setupMock: func(*MockClient) {},
			opts:      ContainerRemoveOptions{Force: true, RemoveVolumes: true},
		},
		"returns configured error": {
			setupMock: func(mock *MockClient) {
				mock.ContainerRemoveError = errors.New("remove failed")
			},
			opts:    ContainerRemoveOptions{Force: true},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)
			err := mock.ContainerRemove(context.Background(), "container-123", tc.opts)
			if (err != nil) != tc.wantErr {
				t.Errorf("ContainerRemove() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if len(mock.ContainerRemoveCalls) != 1 {
				t.Fatalf("ContainerRemoveCalls = %d, want 1", len(mock.ContainerRemoveCalls))
			}
			call := mock.ContainerRemoveCalls[0]
			if call.ID != "container-123" {
				t.Errorf("ContainerRemoveCalls[0].ID = %q, want %q", call.ID, "container-123")
			}
			if diff := cmp.Diff(tc.opts, call.Opts); diff != "" {
				t.Errorf("ContainerRemoveCalls[0].Opts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockClient_ContainerInspect(t *testing.T) {
	t.Parallel()

	want := &Container{ID: "container-123", Name: "web", State: ContainerState{Status: "running", Running: true}}
	tests := map[string]struct {
		setupMock func(*MockClient)
		want      *Container
		wantErr   bool
	}{
		"returns configured result": {
			setupMock: func(mock *MockClient) {
				mock.ContainerInspectResult = want
			},
			want: want,
		},
		"returns not found error": {
			setupMock: func(mock *MockClient) {
				mock.ContainerInspectError = ErrContainerNotFound("container-123")
			},
			wantErr: true,
		},
		"uses callback": {
			setupMock: func(mock *MockClient) {
				mock.OnContainerInspect = func(_ context.Context, id string) (*Container, error) {
					return &Container{ID: id, Name: "callback"}, nil
				}
			},
			want: &Container{ID: "container-123", Name: "callback"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := NewMockClient()
			tc.setupMock(mock)
			got, err := mock.ContainerInspect(context.Background(), "container-123")
			if (err != nil) != tc.wantErr {
				t.Errorf("ContainerInspect() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ContainerInspect() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff([]string{"container-123"}, mock.ContainerInspectCalls); diff != "" {
				t.Errorf("ContainerInspectCalls mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockClient_ContainerList(t *testing.T) {
	t.Parallel()

	opts := ContainerListOptions{All: true, Limit: 10, Filters: map[string][]string{"label": {"yar.managed=true"}}}
	mock := NewMockClient()
	mock.ContainerListResult = []Container{{ID: "container-1"}, {ID: "container-2"}}

	got, err := mock.ContainerList(context.Background(), opts)
	if err != nil {
		t.Fatalf("ContainerList() error = %v", err)
	}
	if diff := cmp.Diff(mock.ContainerListResult, got); diff != "" {
		t.Errorf("ContainerList() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]ContainerListOptions{opts}, mock.ContainerListCalls); diff != "" {
		t.Errorf("ContainerListCalls mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_ContainerLogs(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	mock.ContainerLogsResult = io.NopCloser(strings.NewReader("line one\nline two\n"))
	opts := ContainerLogOptions{ShowStdout: true, Follow: true, Tail: "2"}

	reader, err := mock.ContainerLogs(context.Background(), "container-123", opts)
	if err != nil {
		t.Fatalf("ContainerLogs() error = %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })

	contents, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if got, want := string(contents), "line one\nline two\n"; got != want {
		t.Errorf("ContainerLogs() contents = %q, want %q", got, want)
	}
	if diff := cmp.Diff([]ContainerLogsCall{{ID: "container-123", Opts: opts}}, mock.ContainerLogsCalls); diff != "" {
		t.Errorf("ContainerLogsCalls mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_ContainerWait(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	mock.ContainerWaitResult = ContainerWaitResult{StatusCode: 42, Error: "process failed"}

	resultC, errC := mock.ContainerWait(context.Background(), "container-123", WaitConditionNextExit)
	select {
	case result := <-resultC:
		if diff := cmp.Diff(mock.ContainerWaitResult, result); diff != "" {
			t.Errorf("ContainerWait() result mismatch (-want +got):\n%s", diff)
		}
	case err := <-errC:
		t.Fatalf("ContainerWait() error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("ContainerWait() did not return a result")
	}

	if diff := cmp.Diff([]ContainerWaitCall{{ID: "container-123", Condition: WaitConditionNextExit}}, mock.ContainerWaitCalls); diff != "" {
		t.Errorf("ContainerWaitCalls mismatch (-want +got):\n%s", diff)
	}
}

func TestMockClient_ContainerWaitReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	mock.ContainerWaitError = errors.New("wait failed")

	_, errC := mock.ContainerWait(context.Background(), "container-123", WaitConditionNotRunning)
	select {
	case err := <-errC:
		if err == nil || err.Error() != "wait failed" {
			t.Errorf("ContainerWait() error = %v, want wait failed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ContainerWait() did not return an error")
	}
}

func TestMockClient_ResetClearsContainerCalls(t *testing.T) {
	t.Parallel()

	mock := NewMockClient()
	ctx := context.Background()
	timeout := time.Second
	_, _ = mock.ContainerCreate(ctx, ContainerConfig{Name: "web", Image: "nginx:latest"})
	_ = mock.ContainerStart(ctx, "container-123")
	_ = mock.ContainerStop(ctx, "container-123", &timeout)
	_ = mock.ContainerRemove(ctx, "container-123", ContainerRemoveOptions{Force: true})
	_, _ = mock.ContainerInspect(ctx, "container-123")
	_, _ = mock.ContainerList(ctx, ContainerListOptions{})
	reader, _ := mock.ContainerLogs(ctx, "container-123", ContainerLogOptions{})
	_ = reader.Close()
	_, _ = mock.ContainerWait(ctx, "container-123", WaitConditionNotRunning)

	mock.Reset()

	if len(mock.ContainerCreateCalls) != 0 || len(mock.ContainerStartCalls) != 0 || len(mock.ContainerStopCalls) != 0 || len(mock.ContainerRemoveCalls) != 0 || len(mock.ContainerInspectCalls) != 0 || len(mock.ContainerListCalls) != 0 || len(mock.ContainerLogsCalls) != 0 || len(mock.ContainerWaitCalls) != 0 {
		t.Error("Reset() did not clear all container call records")
	}
}
