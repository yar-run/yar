package errors

import (
	"errors"
	"testing"
)

// TestConfigError tests ConfigError formatting and unwrapping
func TestConfigError(t *testing.T) {
	t.Run("Error formats with path and field", func(t *testing.T) {
		err := &ConfigError{
			Path:    "/home/user/.config/yar/config.yaml",
			Field:   "secrets.local.provider",
			Message: "invalid provider type",
		}
		got := err.Error()
		want := "config error: /home/user/.config/yar/config.yaml: secrets.local.provider: invalid provider type"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Error formats without field", func(t *testing.T) {
		err := &ConfigError{
			Path:    "/home/user/.config/yar/config.yaml",
			Message: "file not readable",
		}
		got := err.Error()
		want := "config error: /home/user/.config/yar/config.yaml: file not readable"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("permission denied")
		err := &ConfigError{
			Path:    "/etc/yar/config.yaml",
			Message: "cannot read file",
			Err:     underlying,
		}
		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})

	t.Run("Unwrap returns nil when no underlying error", func(t *testing.T) {
		err := &ConfigError{
			Path:    "/etc/yar/config.yaml",
			Message: "cannot read file",
		}
		if err.Unwrap() != nil {
			t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
		}
	})
}

// TestValidationError tests ValidationError formatting
func TestValidationError(t *testing.T) {
	t.Run("Error formats with single error", func(t *testing.T) {
		err := &ValidationError{
			Field:   "port",
			Value:   -1,
			Message: "must be positive",
		}
		got := err.Error()
		want := "validation error: port: must be positive (got: -1)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Error formats with multiple errors", func(t *testing.T) {
		err := &ValidationError{
			Field:   "config",
			Message: "multiple validation failures",
			Errors: []string{
				"port: must be positive",
				"host: required field",
			},
		}
		got := err.Error()
		// Should contain both errors
		if got == "" {
			t.Error("Error() returned empty string")
		}
		// Check that it includes the errors
		for _, e := range err.Errors {
			if !contains(got, e) {
				t.Errorf("Error() %q does not contain %q", got, e)
			}
		}
	})
}

// TestNotFoundError tests NotFoundError formatting
func TestNotFoundError(t *testing.T) {
	t.Run("Error formats with resource and name", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "secret",
			Name:     "redis_pass",
			Message:  "not found in pass store",
		}
		got := err.Error()
		want := "secret not found: redis_pass: not found in pass store"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Error formats without message", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "pack",
			Name:     "redis",
		}
		got := err.Error()
		want := "pack not found: redis"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

// TestSecretError tests SecretError formatting and unwrapping
func TestSecretError(t *testing.T) {
	t.Run("Error formats with provider key and op", func(t *testing.T) {
		err := &SecretError{
			Provider: "pass",
			Key:      "redis_pass",
			Op:       "get",
		}
		got := err.Error()
		want := "secret error: pass: get redis_pass"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Error includes underlying error", func(t *testing.T) {
		underlying := errors.New("gpg decryption failed")
		err := &SecretError{
			Provider: "pass",
			Key:      "redis_pass",
			Op:       "get",
			Err:      underlying,
		}
		got := err.Error()
		if !contains(got, "gpg decryption failed") {
			t.Errorf("Error() %q does not contain underlying error", got)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("gpg decryption failed")
		err := &SecretError{
			Provider: "pass",
			Key:      "redis_pass",
			Op:       "get",
			Err:      underlying,
		}
		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})
}

// TestPackError tests PackError formatting and unwrapping
func TestPackError(t *testing.T) {
	t.Run("Error formats with pack and message", func(t *testing.T) {
		err := &PackError{
			Pack:    "redis",
			Message: "schema validation failed",
		}
		got := err.Error()
		want := "pack error: redis: schema validation failed"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("invalid yaml")
		err := &PackError{
			Pack:    "redis",
			Message: "failed to parse",
			Err:     underlying,
		}
		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})
}

// TestDockerError tests DockerError formatting and unwrapping
func TestDockerError(t *testing.T) {
	t.Run("Error formats with op and target", func(t *testing.T) {
		err := &DockerError{
			Op:      "start",
			Target:  "redis-container",
			Message: "container already running",
		}
		got := err.Error()
		want := "docker error: start redis-container: container already running"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("connection refused")
		err := &DockerError{
			Op:      "start",
			Target:  "redis-container",
			Message: "failed to connect",
			Err:     underlying,
		}
		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})
}

// TestKubernetesError tests KubernetesError formatting and unwrapping
func TestKubernetesError(t *testing.T) {
	t.Run("Error formats with namespace", func(t *testing.T) {
		err := &KubernetesError{
			Op:        "apply",
			Resource:  "deployment",
			Name:      "redis",
			Namespace: "default",
		}
		got := err.Error()
		want := "kubernetes error: apply deployment/redis in default"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Error formats without namespace", func(t *testing.T) {
		err := &KubernetesError{
			Op:       "get",
			Resource: "namespace",
			Name:     "production",
		}
		got := err.Error()
		want := "kubernetes error: get namespace/production"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("forbidden")
		err := &KubernetesError{
			Op:        "delete",
			Resource:  "pod",
			Name:      "redis-0",
			Namespace: "default",
			Err:       underlying,
		}
		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})
}

// TestNetworkError tests NetworkError formatting and unwrapping
func TestNetworkError(t *testing.T) {
	t.Run("Error formats with op and target", func(t *testing.T) {
		err := &NetworkError{
			Op:      "vpn",
			Target:  "office.example.com",
			Message: "connection timed out",
		}
		got := err.Error()
		want := "network error: vpn office.example.com: connection timed out"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("no route to host")
		err := &NetworkError{
			Op:      "dns",
			Target:  "redis.local",
			Message: "resolution failed",
			Err:     underlying,
		}
		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})
}

// TestInterfaceCompliance verifies all error types satisfy the error interface
func TestInterfaceCompliance(t *testing.T) {
	var _ error = (*ConfigError)(nil)
	var _ error = (*ValidationError)(nil)
	var _ error = (*NotFoundError)(nil)
	var _ error = (*SecretError)(nil)
	var _ error = (*PackError)(nil)
	var _ error = (*DockerError)(nil)
	var _ error = (*KubernetesError)(nil)
	var _ error = (*NetworkError)(nil)
}

// TestUnwrapCompliance verifies error types with Err field satisfy Unwrap interface
func TestUnwrapCompliance(t *testing.T) {
	type unwrapper interface {
		Unwrap() error
	}
	var _ unwrapper = (*ConfigError)(nil)
	var _ unwrapper = (*SecretError)(nil)
	var _ unwrapper = (*PackError)(nil)
	var _ unwrapper = (*DockerError)(nil)
	var _ unwrapper = (*KubernetesError)(nil)
	var _ unwrapper = (*NetworkError)(nil)
}

// helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
