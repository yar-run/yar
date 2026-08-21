package docker

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestContainer_Struct(t *testing.T) {
	t.Parallel()

	now := time.Now()
	container := Container{
		ID:      "abc123",
		Name:    "test-container",
		Image:   "nginx:latest",
		Command: "/docker-entrypoint.sh nginx -g 'daemon off;'",
		Created: now,
		State: ContainerState{
			Status:     "running",
			Running:    true,
			Paused:     false,
			ExitCode:   0,
			StartedAt:  now,
			FinishedAt: time.Time{},
			Error:      "",
		},
		Ports: []PortBinding{
			{HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "80", Protocol: "tcp"},
		},
		Labels:     map[string]string{"yar.managed": "true", "app": "web"},
		NetworkIDs: []string{"net-1", "net-2"},
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"ID":                  {got: container.ID, want: "abc123"},
		"Name":                {got: container.Name, want: "test-container"},
		"Image":               {got: container.Image, want: "nginx:latest"},
		"Command":             {got: container.Command, want: "/docker-entrypoint.sh nginx -g 'daemon off;'"},
		"Created":             {got: container.Created, want: now},
		"State.Status":        {got: container.State.Status, want: "running"},
		"State.Running":       {got: container.State.Running, want: true},
		"State.Paused":        {got: container.State.Paused, want: false},
		"State.ExitCode":      {got: container.State.ExitCode, want: 0},
		"Ports count":         {got: len(container.Ports), want: 1},
		"Labels count":        {got: len(container.Labels), want: 2},
		"NetworkIDs count":    {got: len(container.NetworkIDs), want: 2},
		"First port HostPort": {got: container.Ports[0].HostPort, want: "8080"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("Container.%s = %v, want %v", name, tc.got, tc.want)
			}
		})
	}
}

func TestContainerState_StatusValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		status  string
		running bool
		paused  bool
	}{
		"created": {status: "created", running: false, paused: false},
		"running": {status: "running", running: true, paused: false},
		"paused":  {status: "paused", running: true, paused: true},
		"exited":  {status: "exited", running: false, paused: false},
		"dead":    {status: "dead", running: false, paused: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			state := ContainerState{
				Status:  tc.status,
				Running: tc.running,
				Paused:  tc.paused,
			}

			if state.Status != tc.status {
				t.Errorf("ContainerState.Status = %q, want %q", state.Status, tc.status)
			}
			if state.Running != tc.running {
				t.Errorf("ContainerState.Running = %v, want %v", state.Running, tc.running)
			}
			if state.Paused != tc.paused {
				t.Errorf("ContainerState.Paused = %v, want %v", state.Paused, tc.paused)
			}
		})
	}
}

func TestContainerState_ExitedWithError(t *testing.T) {
	t.Parallel()

	state := ContainerState{
		Status:   "exited",
		Running:  false,
		ExitCode: 137,
		Error:    "OOMKilled",
	}

	if state.ExitCode != 137 {
		t.Errorf("ContainerState.ExitCode = %d, want 137", state.ExitCode)
	}
	if state.Error != "OOMKilled" {
		t.Errorf("ContainerState.Error = %q, want %q", state.Error, "OOMKilled")
	}
}

func TestContainerConfig_AllFields(t *testing.T) {
	t.Parallel()

	config := ContainerConfig{
		Name:       "my-app",
		Image:      "myapp:v1",
		Command:    []string{"./app", "--port=8080"},
		Entrypoint: []string{"/bin/sh", "-c"},
		Env:        []string{"PORT=8080", "DEBUG=true"},
		Labels:     map[string]string{"version": "v1"},
		WorkingDir: "/app",
		Ports: []PortMapping{
			{HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "8080", Protocol: "tcp"},
		},
		Volumes: []VolumeMount{
			{Source: "/host/data", Target: "/data", ReadOnly: false, Type: "bind"},
		},
		Networks:      []string{"frontend", "backend"},
		RestartPolicy: RestartPolicy{Name: "always", MaximumRetryCount: 0},
		Memory:        536870912, // 512MB
		CPUs:          1.5,
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"Name":               {got: config.Name, want: "my-app"},
		"Image":              {got: config.Image, want: "myapp:v1"},
		"Command length":     {got: len(config.Command), want: 2},
		"Entrypoint length":  {got: len(config.Entrypoint), want: 2},
		"Env length":         {got: len(config.Env), want: 2},
		"WorkingDir":         {got: config.WorkingDir, want: "/app"},
		"Ports length":       {got: len(config.Ports), want: 1},
		"Volumes length":     {got: len(config.Volumes), want: 1},
		"Networks length":    {got: len(config.Networks), want: 2},
		"RestartPolicy.Name": {got: config.RestartPolicy.Name, want: "always"},
		"Memory":             {got: config.Memory, want: int64(536870912)},
		"CPUs":               {got: config.CPUs, want: 1.5},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("ContainerConfig.%s = %v, want %v", name, tc.got, tc.want)
			}
		})
	}
}

func TestPortMapping_Defaults(t *testing.T) {
	t.Parallel()

	// Zero-value PortMapping
	pm := PortMapping{}

	if pm.HostIP != "" {
		t.Errorf("PortMapping.HostIP default = %q, want empty", pm.HostIP)
	}
	if pm.HostPort != "" {
		t.Errorf("PortMapping.HostPort default = %q, want empty", pm.HostPort)
	}
	if pm.Protocol != "" {
		t.Errorf("PortMapping.Protocol default = %q, want empty", pm.Protocol)
	}
}

func TestPortMapping_WithValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		pm   PortMapping
		want PortMapping
	}{
		"tcp mapping": {
			pm: PortMapping{
				HostIP:        "0.0.0.0",
				HostPort:      "8080",
				ContainerPort: "80",
				Protocol:      "tcp",
			},
			want: PortMapping{
				HostIP:        "0.0.0.0",
				HostPort:      "8080",
				ContainerPort: "80",
				Protocol:      "tcp",
			},
		},
		"udp mapping": {
			pm: PortMapping{
				HostIP:        "127.0.0.1",
				HostPort:      "53",
				ContainerPort: "53",
				Protocol:      "udp",
			},
			want: PortMapping{
				HostIP:        "127.0.0.1",
				HostPort:      "53",
				ContainerPort: "53",
				Protocol:      "udp",
			},
		},
		"port range": {
			pm: PortMapping{
				HostPort:      "8080-8090",
				ContainerPort: "80-90",
			},
			want: PortMapping{
				HostPort:      "8080-8090",
				ContainerPort: "80-90",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(tc.want, tc.pm); diff != "" {
				t.Errorf("PortMapping mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVolumeMount_Types(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mount VolumeMount
	}{
		"bind mount": {
			mount: VolumeMount{
				Source:   "/host/path",
				Target:   "/container/path",
				ReadOnly: false,
				Type:     "bind",
			},
		},
		"volume mount": {
			mount: VolumeMount{
				Source:   "my-volume",
				Target:   "/data",
				ReadOnly: false,
				Type:     "volume",
			},
		},
		"tmpfs mount": {
			mount: VolumeMount{
				Source:   "",
				Target:   "/tmp",
				ReadOnly: false,
				Type:     "tmpfs",
			},
		},
		"read-only bind": {
			mount: VolumeMount{
				Source:   "/host/config",
				Target:   "/etc/config",
				ReadOnly: true,
				Type:     "bind",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vm := tc.mount
			if vm.Target == "" {
				t.Error("VolumeMount.Target should not be empty")
			}
		})
	}
}

func TestRestartPolicy_Values(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		policy RestartPolicy
	}{
		"no restart": {
			policy: RestartPolicy{Name: "no", MaximumRetryCount: 0},
		},
		"always restart": {
			policy: RestartPolicy{Name: "always", MaximumRetryCount: 0},
		},
		"on-failure with retries": {
			policy: RestartPolicy{Name: "on-failure", MaximumRetryCount: 5},
		},
		"unless-stopped": {
			policy: RestartPolicy{Name: "unless-stopped", MaximumRetryCount: 0},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rp := tc.policy
			if rp.Name == "" {
				t.Error("RestartPolicy.Name should not be empty")
			}
		})
	}
}

func TestContainerRemoveOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := ContainerRemoveOptions{}

	if opts.Force {
		t.Error("ContainerRemoveOptions.Force default = true, want false")
	}
	if opts.RemoveVolumes {
		t.Error("ContainerRemoveOptions.RemoveVolumes default = true, want false")
	}
}

func TestContainerRemoveOptions_WithValues(t *testing.T) {
	t.Parallel()

	opts := ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}

	if !opts.Force {
		t.Error("ContainerRemoveOptions.Force = false, want true")
	}
	if !opts.RemoveVolumes {
		t.Error("ContainerRemoveOptions.RemoveVolumes = false, want true")
	}
}

func TestContainerListOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := ContainerListOptions{}

	if opts.All {
		t.Error("ContainerListOptions.All default = true, want false")
	}
	if opts.Limit != 0 {
		t.Errorf("ContainerListOptions.Limit default = %d, want 0", opts.Limit)
	}
	if opts.Filters != nil {
		t.Errorf("ContainerListOptions.Filters default = %v, want nil", opts.Filters)
	}
}

func TestContainerListOptions_WithFilters(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filters map[string][]string
	}{
		"filter by name": {
			filters: map[string][]string{"name": {"my-container"}},
		},
		"filter by label": {
			filters: map[string][]string{"label": {"yar.managed=true"}},
		},
		"filter by status": {
			filters: map[string][]string{"status": {"running", "paused"}},
		},
		"multiple filters": {
			filters: map[string][]string{
				"label":  {"yar.managed=true"},
				"status": {"running"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := ContainerListOptions{Filters: tc.filters}

			if diff := cmp.Diff(tc.filters, opts.Filters); diff != "" {
				t.Errorf("ContainerListOptions.Filters mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestContainerLogOptions_Combinations(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := map[string]struct {
		opts ContainerLogOptions
	}{
		"stdout only": {
			opts: ContainerLogOptions{ShowStdout: true, ShowStderr: false},
		},
		"stderr only": {
			opts: ContainerLogOptions{ShowStdout: false, ShowStderr: true},
		},
		"both streams": {
			opts: ContainerLogOptions{ShowStdout: true, ShowStderr: true},
		},
		"follow mode": {
			opts: ContainerLogOptions{ShowStdout: true, Follow: true},
		},
		"tail last 100": {
			opts: ContainerLogOptions{ShowStdout: true, Tail: "100"},
		},
		"tail all": {
			opts: ContainerLogOptions{ShowStdout: true, Tail: "all"},
		},
		"since timestamp": {
			opts: ContainerLogOptions{ShowStdout: true, Since: now.Add(-time.Hour)},
		},
		"until timestamp": {
			opts: ContainerLogOptions{ShowStdout: true, Until: now},
		},
		"with timestamps": {
			opts: ContainerLogOptions{ShowStdout: true, Timestamps: true},
		},
		"full options": {
			opts: ContainerLogOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     true,
				Tail:       "50",
				Timestamps: true,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Verify options can be set without issues
			opts := tc.opts
			_ = opts
		})
	}
}

func TestWaitCondition_Constants(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		condition WaitCondition
		want      string
	}{
		"not-running": {condition: WaitConditionNotRunning, want: "not-running"},
		"next-exit":   {condition: WaitConditionNextExit, want: "next-exit"},
		"removed":     {condition: WaitConditionRemoved, want: "removed"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if string(tc.condition) != tc.want {
				t.Errorf("WaitCondition = %q, want %q", tc.condition, tc.want)
			}
		})
	}
}

func TestContainerWaitResult_Success(t *testing.T) {
	t.Parallel()

	result := ContainerWaitResult{
		StatusCode: 0,
		Error:      "",
	}

	if result.StatusCode != 0 {
		t.Errorf("ContainerWaitResult.StatusCode = %d, want 0", result.StatusCode)
	}
	if result.Error != "" {
		t.Errorf("ContainerWaitResult.Error = %q, want empty", result.Error)
	}
}

func TestContainerWaitResult_Failure(t *testing.T) {
	t.Parallel()

	result := ContainerWaitResult{
		StatusCode: 1,
		Error:      "exit status 1",
	}

	if result.StatusCode != 1 {
		t.Errorf("ContainerWaitResult.StatusCode = %d, want 1", result.StatusCode)
	}
	if result.Error != "exit status 1" {
		t.Errorf("ContainerWaitResult.Error = %q, want %q", result.Error, "exit status 1")
	}
}

func TestContainer_Comparison(t *testing.T) {
	t.Parallel()

	now := time.Now()
	c1 := Container{
		ID:      "abc123",
		Name:    "test",
		Image:   "nginx:latest",
		Created: now,
		State:   ContainerState{Status: "running", Running: true},
		Labels:  map[string]string{"app": "web"},
	}

	c2 := Container{
		ID:      "abc123",
		Name:    "test",
		Image:   "nginx:latest",
		Created: now,
		State:   ContainerState{Status: "running", Running: true},
		Labels:  map[string]string{"app": "web"},
	}

	if diff := cmp.Diff(c1, c2); diff != "" {
		t.Errorf("Container mismatch (-want +got):\n%s", diff)
	}
}

func TestContainer_Comparison_DetectsDifferences(t *testing.T) {
	t.Parallel()

	c1 := Container{ID: "abc123", Name: "test"}
	c2 := Container{ID: "xyz789", Name: "test"}

	diff := cmp.Diff(c1, c2)
	if diff == "" {
		t.Error("cmp.Diff should detect difference between containers with different IDs")
	}
}

func TestPortBinding_Struct(t *testing.T) {
	t.Parallel()

	pb := PortBinding{
		HostIP:        "0.0.0.0",
		HostPort:      "8080",
		ContainerPort: "80",
		Protocol:      "tcp",
	}

	tests := map[string]struct {
		got  string
		want string
	}{
		"HostIP":        {got: pb.HostIP, want: "0.0.0.0"},
		"HostPort":      {got: pb.HostPort, want: "8080"},
		"ContainerPort": {got: pb.ContainerPort, want: "80"},
		"Protocol":      {got: pb.Protocol, want: "tcp"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("PortBinding.%s = %q, want %q", name, tc.got, tc.want)
			}
		})
	}
}
