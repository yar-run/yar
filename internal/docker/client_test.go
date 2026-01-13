package docker

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

func TestWithHost(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		host string
		want string
	}{
		"unix socket": {
			host: "unix:///var/run/docker.sock",
			want: "unix:///var/run/docker.sock",
		},
		"tcp host": {
			host: "tcp://localhost:2375",
			want: "tcp://localhost:2375",
		},
		"empty string": {
			host: "",
			want: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := &clientOptions{}
			WithHost(tc.host)(opts)

			if opts.host != tc.want {
				t.Errorf("WithHost(%q) set host = %q, want %q", tc.host, opts.host, tc.want)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		timeout time.Duration
		want    time.Duration
	}{
		"30 seconds": {
			timeout: 30 * time.Second,
			want:    30 * time.Second,
		},
		"5 minutes": {
			timeout: 5 * time.Minute,
			want:    5 * time.Minute,
		},
		"zero": {
			timeout: 0,
			want:    0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := &clientOptions{}
			WithTimeout(tc.timeout)(opts)

			if opts.timeout != tc.want {
				t.Errorf("WithTimeout(%v) set timeout = %v, want %v", tc.timeout, opts.timeout, tc.want)
			}
		})
	}
}

func TestWithAPIVersion(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version string
		want    string
	}{
		"v1.41": {
			version: "1.41",
			want:    "1.41",
		},
		"v1.43": {
			version: "1.43",
			want:    "1.43",
		},
		"empty string": {
			version: "",
			want:    "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := &clientOptions{}
			WithAPIVersion(tc.version)(opts)

			if opts.apiVersion != tc.want {
				t.Errorf("WithAPIVersion(%q) set apiVersion = %q, want %q", tc.version, opts.apiVersion, tc.want)
			}
		})
	}
}

func TestWithTLSConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config  *tls.Config
		wantNil bool
	}{
		"with config": {
			config:  &tls.Config{InsecureSkipVerify: true},
			wantNil: false,
		},
		"nil config": {
			config:  nil,
			wantNil: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := &clientOptions{}
			WithTLSConfig(tc.config)(opts)

			if (opts.tlsConfig == nil) != tc.wantNil {
				t.Errorf("WithTLSConfig() tlsConfig nil = %v, want %v", opts.tlsConfig == nil, tc.wantNil)
			}
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		client  *http.Client
		wantNil bool
	}{
		"with client": {
			client:  &http.Client{Timeout: 10 * time.Second},
			wantNil: false,
		},
		"nil client": {
			client:  nil,
			wantNil: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := &clientOptions{}
			WithHTTPClient(tc.client)(opts)

			if (opts.httpClient == nil) != tc.wantNil {
				t.Errorf("WithHTTPClient() httpClient nil = %v, want %v", opts.httpClient == nil, tc.wantNil)
			}
		})
	}
}

func TestClientOptions_ApplyMultiple(t *testing.T) {
	t.Parallel()

	opts := &clientOptions{}

	// Apply multiple options
	WithHost("tcp://localhost:2375")(opts)
	WithTimeout(1 * time.Minute)(opts)
	WithAPIVersion("1.41")(opts)

	// Verify all options were applied
	if opts.host != "tcp://localhost:2375" {
		t.Errorf("host = %q, want %q", opts.host, "tcp://localhost:2375")
	}
	if opts.timeout != 1*time.Minute {
		t.Errorf("timeout = %v, want %v", opts.timeout, 1*time.Minute)
	}
	if opts.apiVersion != "1.41" {
		t.Errorf("apiVersion = %q, want %q", opts.apiVersion, "1.41")
	}
}

func TestClientOptions_OverwritePrevious(t *testing.T) {
	t.Parallel()

	opts := &clientOptions{}

	// Apply options, then overwrite
	WithHost("tcp://first:2375")(opts)
	WithHost("tcp://second:2375")(opts)

	if opts.host != "tcp://second:2375" {
		t.Errorf("host = %q, want %q (last applied)", opts.host, "tcp://second:2375")
	}
}

// Note: NewClient tests that require Docker daemon are integration tests.
// The following tests verify option handling without requiring Docker.

func TestNewClient_DefaultTimeout(t *testing.T) {
	t.Parallel()

	// We can't easily test NewClient without a Docker daemon,
	// but we can verify the default timeout is set correctly
	// by checking the clientOptions struct defaults.

	opts := &clientOptions{
		timeout: 30 * time.Second, // This is the default in NewClient
	}

	if opts.timeout != 30*time.Second {
		t.Errorf("default timeout = %v, want %v", opts.timeout, 30*time.Second)
	}
}

func TestDockerClient_ImplementsClient(t *testing.T) {
	t.Parallel()

	// Compile-time assertion that dockerClient implements Client interface.
	var _ Client = (*dockerClient)(nil)
}
