package errors

import (
	"fmt"
	"strings"
)

// ConfigError represents configuration-related errors
type ConfigError struct {
	Path    string // file path that caused the error
	Field   string // specific field, if applicable
	Message string // human-readable description
	Err     error  // underlying error, if any
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error: %s: %s: %s", e.Path, e.Field, e.Message)
	}
	return fmt.Sprintf("config error: %s: %s", e.Path, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// ValidationError represents schema or value validation failures
type ValidationError struct {
	Field   string   // field that failed validation
	Value   any      // the invalid value
	Message string   // description of why validation failed
	Errors  []string // multiple validation errors, if applicable
}

func (e *ValidationError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("validation error: %s: %s\n  - %s", e.Field, e.Message, strings.Join(e.Errors, "\n  - "))
	}
	return fmt.Sprintf("validation error: %s: %s (got: %v)", e.Field, e.Message, e.Value)
}

// NotFoundError represents missing resources
type NotFoundError struct {
	Resource string // type of resource (file, secret, pack, service)
	Name     string // name/identifier of the resource
	Message  string // additional context
}

func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s not found: %s: %s", e.Resource, e.Name, e.Message)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Name)
}

// SecretError represents secret operation failures
type SecretError struct {
	Provider string // provider name (pass, keychain, azure, etc.)
	Key      string // secret key
	Op       string // operation: get, set, delete, list, sync
	Err      error  // underlying error
}

func (e *SecretError) Error() string {
	base := fmt.Sprintf("secret error: %s: %s %s", e.Provider, e.Op, e.Key)
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", base, e.Err)
	}
	return base
}

func (e *SecretError) Unwrap() error {
	return e.Err
}

// PackError represents pack-related errors
type PackError struct {
	Pack    string // pack name
	Message string // description
	Err     error  // underlying error
}

func (e *PackError) Error() string {
	base := fmt.Sprintf("pack error: %s: %s", e.Pack, e.Message)
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", base, e.Err)
	}
	return base
}

func (e *PackError) Unwrap() error {
	return e.Err
}

// DockerError represents Docker operation failures
type DockerError struct {
	Op      string // operation: create, start, stop, remove, etc.
	Target  string // container/network/volume name
	Message string // description
	Err     error  // underlying error
}

func (e *DockerError) Error() string {
	base := fmt.Sprintf("docker error: %s %s: %s", e.Op, e.Target, e.Message)
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", base, e.Err)
	}
	return base
}

func (e *DockerError) Unwrap() error {
	return e.Err
}

// KubernetesError represents Kubernetes operation failures
type KubernetesError struct {
	Op        string // operation: apply, delete, get, etc.
	Resource  string // resource type (deployment, service, etc.)
	Name      string // resource name
	Namespace string // namespace
	Err       error  // underlying error
}

func (e *KubernetesError) Error() string {
	if e.Namespace != "" {
		return fmt.Sprintf("kubernetes error: %s %s/%s in %s", e.Op, e.Resource, e.Name, e.Namespace)
	}
	return fmt.Sprintf("kubernetes error: %s %s/%s", e.Op, e.Resource, e.Name)
}

func (e *KubernetesError) Unwrap() error {
	return e.Err
}

// NetworkError represents network-related failures
type NetworkError struct {
	Op      string // operation: vpn, dns, hosts
	Target  string // target (hostname, IP, etc.)
	Message string // description
	Err     error  // underlying error
}

func (e *NetworkError) Error() string {
	base := fmt.Sprintf("network error: %s %s: %s", e.Op, e.Target, e.Message)
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", base, e.Err)
	}
	return base
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}
