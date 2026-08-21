package docker

import "fmt"

// DockerError represents a Docker operation failure.
type DockerError struct {
	Op      string // Operation: "network.create", "network.remove", etc.
	Name    string // Resource name
	Message string // Human-readable message
	Err     error  // Underlying error
}

// Error implements the error interface.
func (e *DockerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("docker %s %s: %s: %v", e.Op, e.Name, e.Message, e.Err)
	}
	return fmt.Sprintf("docker %s %s: %s", e.Op, e.Name, e.Message)
}

// Unwrap returns the underlying error.
func (e *DockerError) Unwrap() error {
	return e.Err
}

// NewDockerError creates a new DockerError.
func NewDockerError(op, name, message string, err error) *DockerError {
	return &DockerError{
		Op:      op,
		Name:    name,
		Message: message,
		Err:     err,
	}
}

// ErrNetworkCreate creates a network creation error.
func ErrNetworkCreate(name string, err error) *DockerError {
	return NewDockerError("network.create", name, "failed to create network", err)
}

// ErrNetworkRemove creates a network removal error.
func ErrNetworkRemove(name string, err error) *DockerError {
	return NewDockerError("network.remove", name, "failed to remove network", err)
}

// ErrNetworkList creates a network listing error.
func ErrNetworkList(err error) *DockerError {
	return NewDockerError("network.list", "", "failed to list networks", err)
}

// ErrNetworkInspect creates a network inspect error.
func ErrNetworkInspect(name string, err error) *DockerError {
	return NewDockerError("network.inspect", name, "failed to inspect network", err)
}

// ErrNetworkNotFound creates a network not found error.
func ErrNetworkNotFound(name string) *DockerError {
	return NewDockerError("network.inspect", name, "network not found", nil)
}

// ErrNetworkInUse creates a network in use error.
func ErrNetworkInUse(name string, containers []string) *DockerError {
	return &DockerError{
		Op:      "network.remove",
		Name:    name,
		Message: fmt.Sprintf("network has %d attached containers", len(containers)),
	}
}

// ErrDaemonConnection creates a Docker daemon connection error.
func ErrDaemonConnection(err error) *DockerError {
	return &DockerError{
		Op:      "connect",
		Name:    "",
		Message: "cannot connect to Docker daemon. Is Docker running?",
		Err:     err,
	}
}

// Container error constructors

// ErrContainerCreate creates a container creation error.
func ErrContainerCreate(name string, err error) *DockerError {
	return NewDockerError("container.create", name, "failed to create container", err)
}

// ErrContainerStart creates a container start error.
func ErrContainerStart(id string, err error) *DockerError {
	return NewDockerError("container.start", id, "failed to start container", err)
}

// ErrContainerStop creates a container stop error.
func ErrContainerStop(id string, err error) *DockerError {
	return NewDockerError("container.stop", id, "failed to stop container", err)
}

// ErrContainerRemove creates a container removal error.
func ErrContainerRemove(id string, err error) *DockerError {
	return NewDockerError("container.remove", id, "failed to remove container", err)
}

// ErrContainerInspect creates a container inspect error.
func ErrContainerInspect(id string, err error) *DockerError {
	return NewDockerError("container.inspect", id, "failed to inspect container", err)
}

// ErrContainerList creates a container list error.
func ErrContainerList(err error) *DockerError {
	return NewDockerError("container.list", "", "failed to list containers", err)
}

// ErrContainerLogs creates a container logs error.
func ErrContainerLogs(id string, err error) *DockerError {
	return NewDockerError("container.logs", id, "failed to get container logs", err)
}

// ErrContainerWait creates a container wait error.
func ErrContainerWait(id string, err error) *DockerError {
	return NewDockerError("container.wait", id, "failed to wait for container", err)
}

// ErrContainerNotFound creates a container not found error.
func ErrContainerNotFound(id string) *DockerError {
	return NewDockerError("container.inspect", id, "container not found", nil)
}

// ErrContainerAlreadyExists creates a container already exists error.
func ErrContainerAlreadyExists(name string) *DockerError {
	return NewDockerError("container.create", name, "container already exists", nil)
}

// ErrContainerRunning creates an error for operations on running containers.
func ErrContainerRunning(id string) *DockerError {
	return NewDockerError("container.remove", id, "container is running", nil)
}
