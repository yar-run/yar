package docker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	containerapi "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/go-cmp/cmp"
)

func TestContainerCreateRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config ContainerConfig
		check  func(t *testing.T, config *containerapi.Config, hostConfig *containerapi.HostConfig, networkingConfig *network.NetworkingConfig)
	}{
		"maps all supported configuration": {
			config: ContainerConfig{
				Name:       "web",
				Image:      "nginx:latest",
				Command:    []string{"nginx", "-g", "daemon off;"},
				Entrypoint: []string{"/docker-entrypoint.sh"},
				Env:        []string{"PORT=8080", "DEBUG=true"},
				Labels:     map[string]string{"app": "web"},
				WorkingDir: "/srv/web",
				Ports: []PortMapping{
					{HostPort: "8080", ContainerPort: "80"},
					{HostIP: "127.0.0.1", HostPort: "5353", ContainerPort: "53", Protocol: "udp"},
				},
				Volumes: []VolumeMount{
					{Source: "/host/config", Target: "/etc/nginx", ReadOnly: true, Type: "bind"},
					{Source: "web-data", Target: "/var/lib/nginx", Type: "volume"},
					{Target: "/tmp", Type: "tmpfs"},
				},
				Networks:      []string{"frontend", "backend"},
				RestartPolicy: RestartPolicy{Name: "on-failure", MaximumRetryCount: 3},
				Memory:        512 * 1024 * 1024,
				CPUs:          1.5,
			},
			check: func(t *testing.T, config *containerapi.Config, hostConfig *containerapi.HostConfig, networkingConfig *network.NetworkingConfig) {
				t.Helper()

				if config.Image != "nginx:latest" {
					t.Errorf("Config.Image = %q, want %q", config.Image, "nginx:latest")
				}
				if diff := cmp.Diff([]string{"nginx", "-g", "daemon off;"}, []string(config.Cmd)); diff != "" {
					t.Errorf("Config.Cmd mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff([]string{"/docker-entrypoint.sh"}, []string(config.Entrypoint)); diff != "" {
					t.Errorf("Config.Entrypoint mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff([]string{"PORT=8080", "DEBUG=true"}, config.Env); diff != "" {
					t.Errorf("Config.Env mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(map[string]string{"app": "web"}, config.Labels); diff != "" {
					t.Errorf("Config.Labels mismatch (-want +got):\n%s", diff)
				}
				if config.WorkingDir != "/srv/web" {
					t.Errorf("Config.WorkingDir = %q, want %q", config.WorkingDir, "/srv/web")
				}

				tcp80, err := nat.NewPort("tcp", "80")
				if err != nil {
					t.Fatalf("nat.NewPort() error = %v", err)
				}
				udp53, err := nat.NewPort("udp", "53")
				if err != nil {
					t.Fatalf("nat.NewPort() error = %v", err)
				}
				wantPorts := nat.PortMap{
					tcp80: {{HostIP: "0.0.0.0", HostPort: "8080"}},
					udp53: {{HostIP: "127.0.0.1", HostPort: "5353"}},
				}
				if diff := cmp.Diff(wantPorts, hostConfig.PortBindings); diff != "" {
					t.Errorf("HostConfig.PortBindings mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(nat.PortSet{tcp80: {}, udp53: {}}, config.ExposedPorts); diff != "" {
					t.Errorf("Config.ExposedPorts mismatch (-want +got):\n%s", diff)
				}
				if len(hostConfig.Mounts) != 3 {
					t.Fatalf("HostConfig.Mounts length = %d, want 3", len(hostConfig.Mounts))
				}
				if hostConfig.Mounts[0].Type != "bind" || !hostConfig.Mounts[0].ReadOnly {
					t.Errorf("first mount = %#v, want read-only bind mount", hostConfig.Mounts[0])
				}
				if hostConfig.Mounts[1].Type != "volume" {
					t.Errorf("second mount type = %q, want volume", hostConfig.Mounts[1].Type)
				}
				if hostConfig.Mounts[2].Type != "tmpfs" || hostConfig.Mounts[2].Source != "" {
					t.Errorf("third mount = %#v, want tmpfs with no source", hostConfig.Mounts[2])
				}
				if hostConfig.RestartPolicy.Name != containerapi.RestartPolicyOnFailure || hostConfig.RestartPolicy.MaximumRetryCount != 3 {
					t.Errorf("HostConfig.RestartPolicy = %#v, want on-failure with 3 retries", hostConfig.RestartPolicy)
				}
				if hostConfig.Memory != 512*1024*1024 {
					t.Errorf("HostConfig.Memory = %d, want %d", hostConfig.Memory, 512*1024*1024)
				}
				if hostConfig.NanoCPUs != 1_500_000_000 {
					t.Errorf("HostConfig.NanoCPUs = %d, want %d", hostConfig.NanoCPUs, 1_500_000_000)
				}
				if len(networkingConfig.EndpointsConfig) != 2 {
					t.Fatalf("NetworkingConfig.EndpointsConfig length = %d, want 2", len(networkingConfig.EndpointsConfig))
				}
				if networkingConfig.EndpointsConfig["frontend"] == nil || networkingConfig.EndpointsConfig["backend"] == nil {
					t.Errorf("NetworkingConfig.EndpointsConfig = %#v, want frontend and backend", networkingConfig.EndpointsConfig)
				}
			},
		},
		"applies documented defaults": {
			config: ContainerConfig{
				Name:    "web",
				Image:   "nginx:latest",
				Ports:   []PortMapping{{HostPort: "8080", ContainerPort: "80"}},
				Volumes: []VolumeMount{{Source: "/host/config", Target: "/etc/nginx"}},
			},
			check: func(t *testing.T, config *containerapi.Config, hostConfig *containerapi.HostConfig, networkingConfig *network.NetworkingConfig) {
				t.Helper()

				tcp80, err := nat.NewPort("tcp", "80")
				if err != nil {
					t.Fatalf("nat.NewPort() error = %v", err)
				}
				wantPorts := nat.PortMap{tcp80: {{HostIP: "0.0.0.0", HostPort: "8080"}}}
				if diff := cmp.Diff(wantPorts, hostConfig.PortBindings); diff != "" {
					t.Errorf("HostConfig.PortBindings mismatch (-want +got):\n%s", diff)
				}
				if hostConfig.Mounts[0].Type != "bind" {
					t.Errorf("HostConfig.Mounts[0].Type = %q, want bind", hostConfig.Mounts[0].Type)
				}
				if hostConfig.RestartPolicy.Name != containerapi.RestartPolicyDisabled {
					t.Errorf("HostConfig.RestartPolicy.Name = %q, want %q", hostConfig.RestartPolicy.Name, containerapi.RestartPolicyDisabled)
				}
				if networkingConfig != nil {
					t.Errorf("NetworkingConfig = %#v, want nil", networkingConfig)
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config, hostConfig, networkingConfig, err := containerCreateRequest(tc.config)
			if err != nil {
				t.Fatalf("containerCreateRequest() error = %v", err)
			}
			tc.check(t, config, hostConfig, networkingConfig)
		})
	}
}

func TestContainerCreateRequest_RejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config ContainerConfig
	}{
		"missing name": {
			config: ContainerConfig{Image: "nginx:latest"},
		},
		"missing image": {
			config: ContainerConfig{Name: "web"},
		},
		"missing container port": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Ports: []PortMapping{{HostPort: "8080"}}},
		},
		"unsupported port protocol": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Ports: []PortMapping{{HostPort: "8080", ContainerPort: "80", Protocol: "sctp"}}},
		},
		"missing mount target": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Volumes: []VolumeMount{{Source: "/host/config"}}},
		},
		"unsupported mount type": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Volumes: []VolumeMount{{Source: "/host/config", Target: "/config", Type: "npipe"}}},
		},
		"tmpfs source": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Volumes: []VolumeMount{{Source: "/host/tmp", Target: "/tmp", Type: "tmpfs"}}},
		},
		"invalid restart policy": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", RestartPolicy: RestartPolicy{Name: "sometimes"}},
		},
		"retries without on failure": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", RestartPolicy: RestartPolicy{Name: "always", MaximumRetryCount: 1}},
		},
		"negative memory": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Memory: -1},
		},
		"negative CPUs": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", CPUs: -1},
		},
		"NaN CPUs": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", CPUs: math.NaN()},
		},
		"infinite CPUs": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", CPUs: math.Inf(1)},
		},
		"CPU limit too large": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", CPUs: math.MaxFloat64},
		},
		"empty network": {
			config: ContainerConfig{Name: "web", Image: "nginx:latest", Networks: []string{""}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, _, err := containerCreateRequest(tc.config)
			if err == nil {
				t.Error("containerCreateRequest() error = nil, want non-nil")
			}
		})
	}
}

func TestContainerStopTimeoutSeconds(t *testing.T) {
	t.Parallel()

	tenSeconds := 10 * time.Second
	oneAndHalfSeconds := 1500 * time.Millisecond
	negative := -time.Second

	tests := map[string]struct {
		timeout *time.Duration
		want    int
		wantErr bool
	}{
		"uses the documented default": {
			want: 10,
		},
		"uses whole seconds": {
			timeout: &tenSeconds,
			want:    10,
		},
		"rounds up fractional seconds": {
			timeout: &oneAndHalfSeconds,
			want:    2,
		},
		"rejects negative durations": {
			timeout: &negative,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := containerStopTimeoutSeconds(tc.timeout)
			if (err != nil) != tc.wantErr {
				t.Errorf("containerStopTimeoutSeconds() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("containerStopTimeoutSeconds() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestContainerFromInspect(t *testing.T) {
	t.Parallel()

	created := "2026-08-18T12:00:00.123456789Z"
	started := "2026-08-18T12:00:02Z"
	finished := "0001-01-01T00:00:00Z"
	tcp80, err := nat.NewPort("tcp", "80")
	if err != nil {
		t.Fatalf("nat.NewPort() error = %v", err)
	}
	udp53, err := nat.NewPort("udp", "53")
	if err != nil {
		t.Fatalf("nat.NewPort() error = %v", err)
	}

	got := containerFromInspect(containerapi.InspectResponse{
		ContainerJSONBase: &containerapi.ContainerJSONBase{
			ID:      "container-id",
			Name:    "/web",
			Path:    "/docker-entrypoint.sh",
			Args:    []string{"nginx", "-g", "daemon off;"},
			Created: created,
			State: &containerapi.State{
				Status:     "running",
				Running:    true,
				ExitCode:   0,
				StartedAt:  started,
				FinishedAt: finished,
			},
		},
		Config: &containerapi.Config{
			Image:  "nginx:latest",
			Labels: map[string]string{"app": "web"},
		},
		NetworkSettings: &containerapi.NetworkSettings{
			NetworkSettingsBase: containerapi.NetworkSettingsBase{
				Ports: nat.PortMap{
					tcp80: {{HostIP: "0.0.0.0", HostPort: "8080"}},
					udp53: nil,
				},
			},
			Networks: map[string]*network.EndpointSettings{
				"frontend": {NetworkID: "network-b"},
				"backend":  {NetworkID: "network-a"},
			},
		},
	})

	wantCreated, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	wantStarted, err := time.Parse(time.RFC3339Nano, started)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	wantFinished, err := time.Parse(time.RFC3339Nano, finished)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	want := Container{
		ID:      "container-id",
		Name:    "web",
		Image:   "nginx:latest",
		Command: "/docker-entrypoint.sh nginx -g daemon off;",
		Created: wantCreated,
		State: ContainerState{
			Status:     "running",
			Running:    true,
			ExitCode:   0,
			StartedAt:  wantStarted,
			FinishedAt: wantFinished,
		},
		Ports: []PortBinding{
			{ContainerPort: "53", Protocol: "udp"},
			{HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "80", Protocol: "tcp"},
		},
		Labels:     map[string]string{"app": "web"},
		NetworkIDs: []string{"network-a", "network-b"},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("containerFromInspect() mismatch (-want +got):\n%s", diff)
	}
}

func TestContainerFromSummary(t *testing.T) {
	t.Parallel()

	got := containerFromSummary(containerapi.Summary{
		ID:      "container-id",
		Names:   []string{"/web"},
		Image:   "nginx:latest",
		Command: "nginx -g daemon off;",
		Created: 1_724_000_000,
		State:   "paused",
		Ports: []containerapi.Port{
			{IP: "127.0.0.1", PrivatePort: 80, PublicPort: 8080, Type: "tcp"},
		},
		Labels: map[string]string{"app": "web"},
		NetworkSettings: &containerapi.NetworkSettingsSummary{
			Networks: map[string]*network.EndpointSettings{
				"frontend": {NetworkID: "network-b"},
				"backend":  {NetworkID: "network-a"},
			},
		},
	})

	want := Container{
		ID:      "container-id",
		Name:    "web",
		Image:   "nginx:latest",
		Command: "nginx -g daemon off;",
		Created: time.Unix(1_724_000_000, 0).UTC(),
		State: ContainerState{
			Status:  "paused",
			Running: true,
			Paused:  true,
		},
		Ports:      []PortBinding{{HostIP: "127.0.0.1", HostPort: "8080", ContainerPort: "80", Protocol: "tcp"}},
		Labels:     map[string]string{"app": "web"},
		NetworkIDs: []string{"network-a", "network-b"},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("containerFromSummary() mismatch (-want +got):\n%s", diff)
	}
}

func TestContainerFromSummary_RestartingIsRunning(t *testing.T) {
	t.Parallel()

	got := containerFromSummary(containerapi.Summary{State: "restarting"})
	if !got.State.Running {
		t.Error("containerFromSummary().State.Running = false, want true")
	}
}

func TestContainerLogOptions(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, time.August, 18, 10, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	until := time.Date(2026, time.August, 18, 11, 0, 0, 0, time.UTC)

	got := containerLogOptions(ContainerLogOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "50",
		Since:      since,
		Until:      until,
		Timestamps: true,
	})
	want := containerapi.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "50",
		Since:      since.UTC().Format(time.RFC3339Nano),
		Until:      until.UTC().Format(time.RFC3339Nano),
		Timestamps: true,
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("containerLogOptions() mismatch (-want +got):\n%s", diff)
	}

	if got := containerLogOptions(ContainerLogOptions{}); got.Tail != "all" {
		t.Errorf("containerLogOptions().Tail = %q, want %q", got.Tail, "all")
	}
}

func TestContainerWaitCondition(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		condition WaitCondition
		want      containerapi.WaitCondition
		wantErr   bool
	}{
		"empty defaults to not running": {
			want: containerapi.WaitConditionNotRunning,
		},
		"not running": {
			condition: WaitConditionNotRunning,
			want:      containerapi.WaitConditionNotRunning,
		},
		"next exit": {
			condition: WaitConditionNextExit,
			want:      containerapi.WaitConditionNextExit,
		},
		"removed": {
			condition: WaitConditionRemoved,
			want:      containerapi.WaitConditionRemoved,
		},
		"invalid": {
			condition: "invalid",
			wantErr:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := containerWaitCondition(tc.condition)
			if (err != nil) != tc.wantErr {
				t.Errorf("containerWaitCondition() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("containerWaitCondition() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestContainerNotModified(t *testing.T) {
	t.Parallel()

	if !containerNotModified(errdefs.ErrNotModified) {
		t.Error("containerNotModified(ErrNotModified) = false, want true")
	}
	if containerNotModified(errors.New("start failed")) {
		t.Error("containerNotModified(unrelated error) = true, want false")
	}
}

func TestContainerConflictClassification(t *testing.T) {
	t.Parallel()

	alreadyExists := errdefs.ErrConflict.WithMessage("Conflict. The container name /web is already in use")
	if !containerAlreadyExists(alreadyExists) {
		t.Error("containerAlreadyExists() = false, want true")
	}
	if containerAlreadyExists(errdefs.ErrConflict.WithMessage("container is running")) {
		t.Error("containerAlreadyExists() = true, want false")
	}
}

func TestContainerNetworkingConfigDeduplicatesNetworks(t *testing.T) {
	t.Parallel()

	config, additional := containerNetworkingConfig([]string{"frontend", "backend", "frontend", "cache", "backend"})
	if config == nil || config.EndpointsConfig["frontend"] == nil || len(config.EndpointsConfig) != 1 {
		t.Fatalf("containerNetworkingConfig() config = %#v, want frontend only", config)
	}
	if diff := cmp.Diff([]string{"backend", "cache"}, additional); diff != "" {
		t.Errorf("additional networks mismatch (-want +got):\n%s", diff)
	}
}

func TestValidateContainerNetworks(t *testing.T) {
	t.Parallel()

	if err := validateContainerNetworks([]string{"frontend", ""}); err == nil {
		t.Error("validateContainerNetworks() error = nil, want non-nil")
	}
}

func TestContainerIsRunning(t *testing.T) {
	t.Parallel()

	if !containerIsRunning(errdefs.ErrConflict.WithMessage("container abc is running")) {
		t.Error("containerIsRunning() = false, want true")
	}
	if containerIsRunning(errdefs.ErrConflict.WithMessage("container abc is already being removed")) {
		t.Error("containerIsRunning() = true, want false")
	}
}

func TestDockerClient_ContainerCreateConnectsAdditionalNetworks(t *testing.T) {
	t.Parallel()

	var requests []string
	var requestsMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		path := dockerRequestPath(request.URL.Path)
		requestsMu.Lock()
		requests = append(requests, request.Method+" "+path)
		requestsMu.Unlock()

		switch path {
		case "/containers/create":
			var requestBody containerapi.CreateRequest
			if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
				t.Errorf("decode create request: %v", err)
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			if len(requestBody.NetworkingConfig.EndpointsConfig) != 1 || requestBody.NetworkingConfig.EndpointsConfig["frontend"] == nil {
				t.Errorf("create networks = %#v, want frontend only", requestBody.NetworkingConfig)
			}
			_, _ = io.Copy(io.Discard, request.Body)
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{"Id":"container-123"}`))
		case "/networks/backend/connect":
			var connectRequest network.ConnectOptions
			if err := json.NewDecoder(request.Body).Decode(&connectRequest); err != nil {
				t.Errorf("decode connect request: %v", err)
				responseWriter.WriteHeader(http.StatusBadRequest)
				return
			}
			if connectRequest.Container != "container-123" {
				t.Errorf("connect container = %q, want %q", connectRequest.Container, "container-123")
			}
			responseWriter.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	id, err := client.ContainerCreate(context.Background(), ContainerConfig{
		Name:     "web",
		Image:    "nginx:latest",
		Networks: []string{"frontend", "backend"},
	})
	if err != nil {
		t.Fatalf("ContainerCreate() error = %v", err)
	}
	if id != "container-123" {
		t.Errorf("ContainerCreate() = %q, want %q", id, "container-123")
	}

	requestsMu.Lock()
	defer requestsMu.Unlock()
	want := []string{"POST /containers/create", "POST /networks/backend/connect"}
	if diff := cmp.Diff(want, requests); diff != "" {
		t.Errorf("requests mismatch (-want +got):\n%s", diff)
	}
}

func TestDockerClient_ContainerCreateCleansUpAfterNetworkConnectionFailure(t *testing.T) {
	t.Parallel()

	var requests []string
	var requestsMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		path := dockerRequestPath(request.URL.Path)
		requestsMu.Lock()
		requests = append(requests, request.Method+" "+path)
		requestsMu.Unlock()

		switch path {
		case "/containers/create":
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{"Id":"container-123"}`))
		case "/networks/backend/connect":
			responseWriter.WriteHeader(http.StatusInternalServerError)
			_, _ = responseWriter.Write([]byte("network unavailable"))
		case "/containers/container-123":
			if request.Method != http.MethodDelete {
				t.Errorf("cleanup method = %s, want DELETE", request.Method)
			}
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	_, err := client.ContainerCreate(context.Background(), ContainerConfig{
		Name:     "web",
		Image:    "nginx:latest",
		Networks: []string{"frontend", "backend"},
	})
	if err == nil {
		t.Fatal("ContainerCreate() error = nil, want non-nil")
	}

	requestsMu.Lock()
	defer requestsMu.Unlock()
	want := []string{"POST /containers/create", "POST /networks/backend/connect", "DELETE /containers/container-123"}
	if diff := cmp.Diff(want, requests); diff != "" {
		t.Errorf("requests mismatch (-want +got):\n%s", diff)
	}
}

func TestDockerClient_ContainerListMapsResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		path := dockerRequestPath(request.URL.Path)
		if request.Method != http.MethodGet || path != "/containers/json" {
			t.Errorf("request = %s %s, want GET /containers/json", request.Method, path)
			responseWriter.WriteHeader(http.StatusNotFound)
			return
		}
		if request.URL.Query().Get("all") != "1" || request.URL.Query().Get("limit") != "10" {
			t.Errorf("query = %q, want all and limit", request.URL.RawQuery)
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`[{"Id":"container-123","Names":["/web"],"Image":"nginx:latest","Command":"nginx","Created":1724000000,"State":"restarting","Ports":[{"IP":"127.0.0.1","PrivatePort":80,"PublicPort":8080,"Type":"tcp"}],"Labels":{"app":"web"},"NetworkSettings":{"Networks":{"frontend":{"NetworkID":"network-123"}}}}]`))
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	got, err := client.ContainerList(context.Background(), ContainerListOptions{
		All:     true,
		Limit:   10,
		Filters: map[string][]string{"label": {"app=web"}},
	})
	if err != nil {
		t.Fatalf("ContainerList() error = %v", err)
	}
	want := []Container{{
		ID:      "container-123",
		Name:    "web",
		Image:   "nginx:latest",
		Command: "nginx",
		Created: time.Unix(1_724_000_000, 0).UTC(),
		State: ContainerState{
			Status:  "restarting",
			Running: true,
		},
		Ports:      []PortBinding{{HostIP: "127.0.0.1", HostPort: "8080", ContainerPort: "80", Protocol: "tcp"}},
		Labels:     map[string]string{"app": "web"},
		NetworkIDs: []string{"network-123"},
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ContainerList() mismatch (-want +got):\n%s", diff)
	}
}

func TestDockerClient_ContainerRemovePreservesNonRunningConflict(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusConflict)
		_, _ = responseWriter.Write([]byte("container is already being removed"))
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	err := client.ContainerRemove(context.Background(), "container-123", ContainerRemoveOptions{})
	if err == nil {
		t.Fatal("ContainerRemove() error = nil, want non-nil")
	}
	if dockerError, ok := err.(*DockerError); !ok || dockerError.Message != "failed to remove container" {
		t.Errorf("ContainerRemove() error = %#v, want generic remove error", err)
	}
}

func TestDockerClient_ContainerStartIsIdempotent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || dockerRequestPath(request.URL.Path) != "/containers/container-123/start" {
			t.Errorf("request = %s %s, want POST /containers/container-123/start", request.Method, dockerRequestPath(request.URL.Path))
		}
		responseWriter.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	if err := client.ContainerStart(context.Background(), "container-123"); err != nil {
		t.Errorf("ContainerStart() error = %v, want nil", err)
	}
}

func TestDockerClient_ContainerStopMapsDefaultTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || dockerRequestPath(request.URL.Path) != "/containers/container-123/stop" {
			t.Errorf("request = %s %s, want POST /containers/container-123/stop", request.Method, dockerRequestPath(request.URL.Path))
		}
		if got := request.URL.Query().Get("t"); got != "10" {
			t.Errorf("stop timeout = %q, want %q", got, "10")
		}
		responseWriter.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	if err := client.ContainerStop(context.Background(), "container-123", nil); err != nil {
		t.Errorf("ContainerStop() error = %v, want nil", err)
	}
}

func TestDockerClient_ContainerInspectMapsResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || dockerRequestPath(request.URL.Path) != "/containers/container-123/json" {
			t.Errorf("request = %s %s, want GET /containers/container-123/json", request.Method, dockerRequestPath(request.URL.Path))
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`{"Id":"container-123","Name":"/web","Created":"2026-08-18T12:00:00Z","Path":"nginx","Args":["-g","daemon off;"],"State":{"Status":"running","Running":true,"StartedAt":"2026-08-18T12:00:01Z"},"Config":{"Image":"nginx:latest","Labels":{"app":"web"}},"NetworkSettings":{"Ports":{"80/tcp":[{"HostIp":"0.0.0.0","HostPort":"8080"}]},"Networks":{"frontend":{"NetworkID":"network-123"}}}}`))
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	got, err := client.ContainerInspect(context.Background(), "container-123")
	if err != nil {
		t.Fatalf("ContainerInspect() error = %v", err)
	}
	want := &Container{
		ID:      "container-123",
		Name:    "web",
		Image:   "nginx:latest",
		Command: "nginx -g daemon off;",
		Created: time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC),
		State: ContainerState{
			Status:    "running",
			Running:   true,
			StartedAt: time.Date(2026, time.August, 18, 12, 0, 1, 0, time.UTC),
		},
		Ports:      []PortBinding{{HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "80", Protocol: "tcp"}},
		Labels:     map[string]string{"app": "web"},
		NetworkIDs: []string{"network-123"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ContainerInspect() mismatch (-want +got):\n%s", diff)
	}
}

func TestDockerClient_ContainerLogsMapsOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || dockerRequestPath(request.URL.Path) != "/containers/container-123/logs" {
			t.Errorf("request = %s %s, want GET /containers/container-123/logs", request.Method, dockerRequestPath(request.URL.Path))
		}
		query := request.URL.Query()
		if query.Get("stdout") != "1" || query.Get("tail") != "all" || query.Get("timestamps") != "1" {
			t.Errorf("logs query = %q, want stdout, tail=all, and timestamps", request.URL.RawQuery)
		}
		_, _ = responseWriter.Write([]byte("line one\n"))
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	reader, err := client.ContainerLogs(context.Background(), "container-123", ContainerLogOptions{ShowStdout: true, Timestamps: true})
	if err != nil {
		t.Fatalf("ContainerLogs() error = %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	contents, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if got, want := string(contents), "line one\n"; got != want {
		t.Errorf("ContainerLogs() contents = %q, want %q", got, want)
	}
}

func TestDockerClient_ContainerWaitMapsResult(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || dockerRequestPath(request.URL.Path) != "/containers/container-123/wait" {
			t.Errorf("request = %s %s, want POST /containers/container-123/wait", request.Method, dockerRequestPath(request.URL.Path))
		}
		if got := request.URL.Query().Get("condition"); got != string(WaitConditionNextExit) {
			t.Errorf("wait condition = %q, want %q", got, WaitConditionNextExit)
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`{"StatusCode":42,"Error":{"Message":"process failed"}}`))
	}))
	defer server.Close()

	client := newDockerClientForTest(t, server)
	resultC, errC := client.ContainerWait(context.Background(), "container-123", WaitConditionNextExit)
	select {
	case result := <-resultC:
		want := ContainerWaitResult{StatusCode: 42, Error: "process failed"}
		if diff := cmp.Diff(want, result); diff != "" {
			t.Errorf("ContainerWait() mismatch (-want +got):\n%s", diff)
		}
	case err := <-errC:
		t.Fatalf("ContainerWait() error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("ContainerWait() timed out")
	}
}

func newDockerClientForTest(t *testing.T, server *httptest.Server) *dockerClient {
	t.Helper()

	baseURL := server.URL
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.WithHost(baseURL),
		dockerclient.WithVersion("1.44"),
	)
	if err != nil {
		t.Fatalf("NewClientWithOpts() error = %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	return &dockerClient{cli: cli, timeout: time.Second}
}

func dockerRequestPath(path string) string {
	return strings.TrimPrefix(path, "/v1.44")
}
