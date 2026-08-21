package docker

import "time"

// Container represents a Docker container.
type Container struct {
	ID         string            `json:"id" yaml:"id"`
	Name       string            `json:"name" yaml:"name"`
	Image      string            `json:"image" yaml:"image"`
	Command    string            `json:"command" yaml:"command"`
	Created    time.Time         `json:"created" yaml:"created"`
	State      ContainerState    `json:"state" yaml:"state"`
	Ports      []PortBinding     `json:"ports,omitempty" yaml:"ports,omitempty"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	NetworkIDs []string          `json:"network_ids,omitempty" yaml:"network_ids,omitempty"`
}

// ContainerState represents the current state of a container.
type ContainerState struct {
	Status     string    `json:"status" yaml:"status"` // created, running, paused, restarting, removing, exited, dead
	Running    bool      `json:"running" yaml:"running"`
	Paused     bool      `json:"paused" yaml:"paused"`
	ExitCode   int       `json:"exit_code" yaml:"exit_code"`
	StartedAt  time.Time `json:"started_at" yaml:"started_at"`
	FinishedAt time.Time `json:"finished_at" yaml:"finished_at"`
	Error      string    `json:"error,omitempty" yaml:"error,omitempty"`
}

// ContainerConfig specifies how to create a container.
type ContainerConfig struct {
	Name       string            `json:"name" yaml:"name"`                                   // Container name (required)
	Image      string            `json:"image" yaml:"image"`                                 // Image reference (required)
	Command    []string          `json:"command,omitempty" yaml:"command,omitempty"`         // Override CMD
	Entrypoint []string          `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`   // Override ENTRYPOINT
	Env        []string          `json:"env,omitempty" yaml:"env,omitempty"`                 // Environment variables ("KEY=value")
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`           // Container labels
	WorkingDir string            `json:"working_dir,omitempty" yaml:"working_dir,omitempty"` // Working directory inside container

	// Host configuration
	Ports         []PortMapping `json:"ports,omitempty" yaml:"ports,omitempty"`                   // Port mappings
	Volumes       []VolumeMount `json:"volumes,omitempty" yaml:"volumes,omitempty"`               // Volume mounts
	Networks      []string      `json:"networks,omitempty" yaml:"networks,omitempty"`             // Networks to connect to
	RestartPolicy RestartPolicy `json:"restart_policy,omitempty" yaml:"restart_policy,omitempty"` // Restart policy

	// Resource constraints
	Memory int64   `json:"memory,omitempty" yaml:"memory,omitempty"` // Memory limit in bytes (0 = unlimited)
	CPUs   float64 `json:"cpus,omitempty" yaml:"cpus,omitempty"`     // CPU limit (0 = unlimited)
}

// PortBinding represents an exposed port on a container (read from container).
type PortBinding struct {
	HostIP        string `json:"host_ip,omitempty" yaml:"host_ip,omitempty"`
	HostPort      string `json:"host_port,omitempty" yaml:"host_port,omitempty"`
	ContainerPort string `json:"container_port" yaml:"container_port"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"` // tcp or udp
}

// PortMapping maps a host port to a container port (used for container creation).
type PortMapping struct {
	HostIP        string `json:"host_ip,omitempty" yaml:"host_ip,omitempty"`     // Host IP to bind (default: "0.0.0.0")
	HostPort      string `json:"host_port,omitempty" yaml:"host_port,omitempty"` // Host port (can be range "8080-8090")
	ContainerPort string `json:"container_port" yaml:"container_port"`           // Container port (required)
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"`   // tcp or udp (default: tcp)
}

// VolumeMount mounts a volume or bind mount into a container.
type VolumeMount struct {
	Source   string `json:"source,omitempty" yaml:"source,omitempty"`       // Host path or volume name
	Target   string `json:"target" yaml:"target"`                           // Container path (required)
	ReadOnly bool   `json:"read_only,omitempty" yaml:"read_only,omitempty"` // Mount as read-only
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`           // bind, volume, or tmpfs (default: bind)
}

// RestartPolicy defines when to restart a container.
type RestartPolicy struct {
	Name              string `json:"name,omitempty" yaml:"name,omitempty"`                               // no, always, on-failure, unless-stopped
	MaximumRetryCount int    `json:"maximum_retry_count,omitempty" yaml:"maximum_retry_count,omitempty"` // For on-failure policy
}

// ContainerRemoveOptions configures container removal.
type ContainerRemoveOptions struct {
	Force         bool `json:"force,omitempty" yaml:"force,omitempty"`                   // Remove even if running
	RemoveVolumes bool `json:"remove_volumes,omitempty" yaml:"remove_volumes,omitempty"` // Remove associated volumes
}

// ContainerListOptions configures container listing.
type ContainerListOptions struct {
	All     bool                `json:"all,omitempty" yaml:"all,omitempty"`         // Include stopped containers
	Limit   int                 `json:"limit,omitempty" yaml:"limit,omitempty"`     // Max number to return (0 = no limit)
	Filters map[string][]string `json:"filters,omitempty" yaml:"filters,omitempty"` // Filter by label, name, status, etc.
}

// ContainerLogOptions configures log streaming.
type ContainerLogOptions struct {
	ShowStdout bool      `json:"show_stdout,omitempty" yaml:"show_stdout,omitempty"` // Include stdout
	ShowStderr bool      `json:"show_stderr,omitempty" yaml:"show_stderr,omitempty"` // Include stderr
	Follow     bool      `json:"follow,omitempty" yaml:"follow,omitempty"`           // Follow log output (tail -f)
	Tail       string    `json:"tail,omitempty" yaml:"tail,omitempty"`               // Number of lines from end ("all" or number)
	Since      time.Time `json:"since,omitempty" yaml:"since,omitempty"`             // Show logs since timestamp
	Until      time.Time `json:"until,omitempty" yaml:"until,omitempty"`             // Show logs until timestamp
	Timestamps bool      `json:"timestamps,omitempty" yaml:"timestamps,omitempty"`   // Include timestamps
}

// WaitCondition specifies what container state to wait for.
type WaitCondition string

const (
	// WaitConditionNotRunning waits until the container is not running.
	WaitConditionNotRunning WaitCondition = "not-running"
	// WaitConditionNextExit waits until the container exits next.
	WaitConditionNextExit WaitCondition = "next-exit"
	// WaitConditionRemoved waits until the container is removed.
	WaitConditionRemoved WaitCondition = "removed"
)

// ContainerWaitResult is returned when container wait completes.
type ContainerWaitResult struct {
	StatusCode int64  `json:"status_code" yaml:"status_code"`         // Exit code
	Error      string `json:"error,omitempty" yaml:"error,omitempty"` // Error message if any
}
