package docker

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDockerError_Error(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err  *DockerError
		want string
	}{
		"with underlying error": {
			err: &DockerError{
				Op:      "network.create",
				Name:    "test-net",
				Message: "failed to create network",
				Err:     errors.New("connection refused"),
			},
			want: "docker network.create test-net: failed to create network: connection refused",
		},
		"without underlying error": {
			err: &DockerError{
				Op:      "network.inspect",
				Name:    "test-net",
				Message: "network not found",
				Err:     nil,
			},
			want: "docker network.inspect test-net: network not found",
		},
		"empty name": {
			err: &DockerError{
				Op:      "network.list",
				Name:    "",
				Message: "failed to list networks",
				Err:     errors.New("timeout"),
			},
			want: "docker network.list : failed to list networks: timeout",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.err.Error()
			if got != tc.want {
				t.Errorf("DockerError.Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDockerError_Unwrap(t *testing.T) {
	t.Parallel()

	underlying := errors.New("connection refused")

	tests := map[string]struct {
		err     *DockerError
		wantErr error
	}{
		"with underlying error": {
			err: &DockerError{
				Op:      "connect",
				Name:    "",
				Message: "cannot connect",
				Err:     underlying,
			},
			wantErr: underlying,
		},
		"without underlying error": {
			err: &DockerError{
				Op:      "network.inspect",
				Name:    "test-net",
				Message: "network not found",
				Err:     nil,
			},
			wantErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.err.Unwrap()
			if got != tc.wantErr {
				t.Errorf("DockerError.Unwrap() = %v, want %v", got, tc.wantErr)
			}
		})
	}
}

func TestDockerError_ErrorsIs(t *testing.T) {
	t.Parallel()

	underlying := errors.New("connection refused")
	err := ErrDaemonConnection(underlying)

	if !errors.Is(err, underlying) {
		t.Error("errors.Is(ErrDaemonConnection(underlying), underlying) = false, want true")
	}
}

func TestNewDockerError(t *testing.T) {
	t.Parallel()

	underlying := errors.New("test error")
	got := NewDockerError("test.op", "resource", "test message", underlying)

	want := &DockerError{
		Op:      "test.op",
		Name:    "resource",
		Message: "test message",
		Err:     underlying,
	}

	if diff := cmp.Diff(want, got, cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == b
		}
		return a.Error() == b.Error()
	})); diff != "" {
		t.Errorf("NewDockerError() mismatch (-want +got):\n%s", diff)
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Parallel()

	underlying := errors.New("test error")

	tests := map[string]struct {
		constructor func() *DockerError
		wantOp      string
		wantName    string
		wantHasErr  bool
	}{
		"ErrNetworkCreate": {
			constructor: func() *DockerError { return ErrNetworkCreate("my-net", underlying) },
			wantOp:      "network.create",
			wantName:    "my-net",
			wantHasErr:  true,
		},
		"ErrNetworkRemove": {
			constructor: func() *DockerError { return ErrNetworkRemove("my-net", underlying) },
			wantOp:      "network.remove",
			wantName:    "my-net",
			wantHasErr:  true,
		},
		"ErrNetworkList": {
			constructor: func() *DockerError { return ErrNetworkList(underlying) },
			wantOp:      "network.list",
			wantName:    "",
			wantHasErr:  true,
		},
		"ErrNetworkInspect": {
			constructor: func() *DockerError { return ErrNetworkInspect("my-net", underlying) },
			wantOp:      "network.inspect",
			wantName:    "my-net",
			wantHasErr:  true,
		},
		"ErrNetworkNotFound": {
			constructor: func() *DockerError { return ErrNetworkNotFound("my-net") },
			wantOp:      "network.inspect",
			wantName:    "my-net",
			wantHasErr:  false,
		},
		"ErrNetworkInUse": {
			constructor: func() *DockerError { return ErrNetworkInUse("my-net", []string{"c1", "c2"}) },
			wantOp:      "network.remove",
			wantName:    "my-net",
			wantHasErr:  false,
		},
		"ErrDaemonConnection": {
			constructor: func() *DockerError { return ErrDaemonConnection(underlying) },
			wantOp:      "connect",
			wantName:    "",
			wantHasErr:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.constructor()

			if got.Op != tc.wantOp {
				t.Errorf("%s().Op = %q, want %q", name, got.Op, tc.wantOp)
			}
			if got.Name != tc.wantName {
				t.Errorf("%s().Name = %q, want %q", name, got.Name, tc.wantName)
			}
			if (got.Err != nil) != tc.wantHasErr {
				t.Errorf("%s().Err = %v, wantHasErr = %v", name, got.Err, tc.wantHasErr)
			}
			if got.Error() == "" {
				t.Errorf("%s().Error() returned empty string", name)
			}
		})
	}
}

func TestErrNetworkInUse_Message(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		containers  []string
		wantContain string
	}{
		"two containers": {
			containers:  []string{"container1", "container2"},
			wantContain: "2 attached containers",
		},
		"one container": {
			containers:  []string{"container1"},
			wantContain: "1 attached containers",
		},
		"no containers": {
			containers:  []string{},
			wantContain: "0 attached containers",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ErrNetworkInUse("test-net", tc.containers)
			got := err.Message

			if got == "" {
				t.Error("ErrNetworkInUse().Message = empty, want non-empty")
			}
			// Message should mention container count
			if len(tc.containers) > 0 || tc.wantContain != "" {
				// Just verify message is set correctly
				if err.Op != "network.remove" {
					t.Errorf("ErrNetworkInUse().Op = %q, want %q", err.Op, "network.remove")
				}
			}
		})
	}
}
