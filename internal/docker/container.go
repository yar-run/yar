package docker

import (
	"context"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	containerapi "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

const defaultContainerStopTimeout = 10 * time.Second

// ContainerCreate creates a container from config without starting it.
func (c *dockerClient) ContainerCreate(ctx context.Context, config ContainerConfig) (string, error) {
	operationCtx, cancel := c.operationContext(ctx)
	defer cancel()

	dockerConfig, hostConfig, networkingConfig, err := containerCreateRequest(config)
	if err != nil {
		return "", ErrContainerCreate(config.Name, err)
	}

	createNetworkingConfig, additionalNetworks := containerNetworkingConfig(config.Networks)
	if createNetworkingConfig != nil {
		networkingConfig = createNetworkingConfig
	}

	response, err := c.cli.ContainerCreate(operationCtx, dockerConfig, hostConfig, networkingConfig, nil, config.Name)
	if err != nil {
		if containerAlreadyExists(err) {
			return "", ErrContainerAlreadyExists(config.Name)
		}
		return "", ErrContainerCreate(config.Name, err)
	}
	for _, networkName := range additionalNetworks {
		if err := c.cli.NetworkConnect(operationCtx, networkName, response.ID, &network.EndpointSettings{}); err != nil {
			cleanupCtx, cleanupCancel := c.operationContext(context.WithoutCancel(ctx))
			_ = c.cli.ContainerRemove(cleanupCtx, response.ID, containerapi.RemoveOptions{Force: true})
			cleanupCancel()
			return "", ErrContainerCreate(config.Name, fmt.Errorf("connect network %q: %w", networkName, err))
		}
	}
	return response.ID, nil
}

// ContainerStart starts a container. Starting an already-running container is a no-op.
func (c *dockerClient) ContainerStart(ctx context.Context, id string) error {
	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	err := c.cli.ContainerStart(ctx, id, containerapi.StartOptions{})
	if err == nil || containerNotModified(err) {
		return nil
	}
	if errdefs.IsNotFound(err) {
		return ErrContainerNotFound(id)
	}
	return ErrContainerStart(id, err)
}

// ContainerStop stops a container. Stopping an already-stopped container is a no-op.
func (c *dockerClient) ContainerStop(ctx context.Context, id string, timeout *time.Duration) error {
	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	seconds, err := containerStopTimeoutSeconds(timeout)
	if err != nil {
		return ErrContainerStop(id, err)
	}

	err = c.cli.ContainerStop(ctx, id, containerapi.StopOptions{Timeout: &seconds})
	if err == nil || containerNotModified(err) {
		return nil
	}
	if errdefs.IsNotFound(err) {
		return ErrContainerNotFound(id)
	}
	return ErrContainerStop(id, err)
}

// ContainerRemove removes a container by ID or name.
func (c *dockerClient) ContainerRemove(ctx context.Context, id string, opts ContainerRemoveOptions) error {
	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	err := c.cli.ContainerRemove(ctx, id, containerapi.RemoveOptions{
		Force:         opts.Force,
		RemoveVolumes: opts.RemoveVolumes,
	})
	if err == nil {
		return nil
	}
	if errdefs.IsNotFound(err) {
		return ErrContainerNotFound(id)
	}
	if !opts.Force && containerIsRunning(err) {
		return ErrContainerRunning(id)
	}
	return ErrContainerRemove(id, err)
}

// ContainerInspect returns detailed information about a container.
func (c *dockerClient) ContainerInspect(ctx context.Context, id string) (*Container, error) {
	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	response, err := c.cli.ContainerInspect(ctx, id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, ErrContainerNotFound(id)
		}
		return nil, ErrContainerInspect(id, err)
	}

	container := containerFromInspect(response)
	return &container, nil
}

// ContainerList lists Docker containers with the requested options.
func (c *dockerClient) ContainerList(ctx context.Context, opts ContainerListOptions) ([]Container, error) {
	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	filterArgs := filters.NewArgs()
	for key, values := range opts.Filters {
		for _, value := range values {
			filterArgs.Add(key, value)
		}
	}

	containers, err := c.cli.ContainerList(ctx, containerapi.ListOptions{
		All:     opts.All,
		Limit:   opts.Limit,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, ErrContainerList(err)
	}

	result := make([]Container, len(containers))
	for i, container := range containers {
		result[i] = containerFromSummary(container)
	}
	return result, nil
}

// ContainerLogs streams container logs. The returned reader must be closed by the caller.
func (c *dockerClient) ContainerLogs(ctx context.Context, id string, opts ContainerLogOptions) (io.ReadCloser, error) {
	reader, err := c.cli.ContainerLogs(ctx, id, containerLogOptions(opts))
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, ErrContainerNotFound(id)
		}
		return nil, ErrContainerLogs(id, err)
	}
	return reader, nil
}

// ContainerWait waits asynchronously for a container to reach condition.
func (c *dockerClient) ContainerWait(ctx context.Context, id string, condition WaitCondition) (<-chan ContainerWaitResult, <-chan error) {
	resultC := make(chan ContainerWaitResult, 1)
	errC := make(chan error, 1)

	dockerCondition, err := containerWaitCondition(condition)
	if err != nil {
		errC <- ErrContainerWait(id, err)
		return resultC, errC
	}

	dockerResults, dockerErrors := c.cli.ContainerWait(ctx, id, dockerCondition)
	go func() {
		select {
		case response, ok := <-dockerResults:
			if !ok {
				return
			}
			result := ContainerWaitResult{StatusCode: response.StatusCode}
			if response.Error != nil {
				result.Error = response.Error.Message
			}
			resultC <- result
		case err, ok := <-dockerErrors:
			if !ok || err == nil {
				return
			}
			if errdefs.IsNotFound(err) {
				errC <- ErrContainerNotFound(id)
				return
			}
			errC <- ErrContainerWait(id, err)
		}
	}()

	return resultC, errC
}

func containerCreateRequest(config ContainerConfig) (*containerapi.Config, *containerapi.HostConfig, *network.NetworkingConfig, error) {
	if config.Name == "" {
		return nil, nil, nil, fmt.Errorf("container name is required")
	}
	if config.Image == "" {
		return nil, nil, nil, fmt.Errorf("container image is required")
	}
	if config.Memory < 0 {
		return nil, nil, nil, fmt.Errorf("memory limit cannot be negative")
	}
	if math.IsNaN(config.CPUs) || math.IsInf(config.CPUs, 0) || config.CPUs < 0 {
		return nil, nil, nil, fmt.Errorf("CPU limit must be finite and non-negative")
	}
	nanoCPUs := math.Round(config.CPUs * 1_000_000_000)
	if nanoCPUs > math.Nextafter(float64(math.MaxInt64), math.Inf(-1)) {
		return nil, nil, nil, fmt.Errorf("CPU limit is too large")
	}

	exposedPorts, portBindings, err := containerPortMappings(config.Ports)
	if err != nil {
		return nil, nil, nil, err
	}
	mounts, err := containerMounts(config.Volumes)
	if err != nil {
		return nil, nil, nil, err
	}
	restartPolicy, err := containerRestartPolicy(config.RestartPolicy)
	if err != nil {
		return nil, nil, nil, err
	}

	dockerConfig := &containerapi.Config{
		Image:        config.Image,
		Cmd:          config.Command,
		Entrypoint:   config.Entrypoint,
		Env:          config.Env,
		Labels:       config.Labels,
		WorkingDir:   config.WorkingDir,
		ExposedPorts: exposedPorts,
	}
	hostConfig := &containerapi.HostConfig{
		PortBindings:  portBindings,
		Mounts:        mounts,
		RestartPolicy: restartPolicy,
		Resources: containerapi.Resources{
			Memory:   config.Memory,
			NanoCPUs: int64(nanoCPUs),
		},
	}

	if len(config.Networks) == 0 {
		return dockerConfig, hostConfig, nil, nil
	}
	if err := validateContainerNetworks(config.Networks); err != nil {
		return nil, nil, nil, err
	}

	endpoints := make(map[string]*network.EndpointSettings, len(config.Networks))
	for _, networkName := range config.Networks {
		endpoints[networkName] = &network.EndpointSettings{}
	}
	return dockerConfig, hostConfig, &network.NetworkingConfig{EndpointsConfig: endpoints}, nil
}

func containerPortMappings(mappings []PortMapping) (nat.PortSet, nat.PortMap, error) {
	if len(mappings) == 0 {
		return nil, nil, nil
	}

	exposedPorts := make(nat.PortSet, len(mappings))
	portBindings := make(nat.PortMap, len(mappings))
	for _, mapping := range mappings {
		if mapping.ContainerPort == "" {
			return nil, nil, fmt.Errorf("container port is required")
		}
		protocol := mapping.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		if protocol != "tcp" && protocol != "udp" {
			return nil, nil, fmt.Errorf("unsupported port protocol %q", protocol)
		}

		port, err := nat.NewPort(protocol, mapping.ContainerPort)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid container port %q: %w", mapping.ContainerPort, err)
		}
		exposedPorts[port] = struct{}{}
		hostIP := mapping.HostIP
		if hostIP == "" {
			hostIP = "0.0.0.0"
		}
		portBindings[port] = append(portBindings[port], nat.PortBinding{
			HostIP:   hostIP,
			HostPort: mapping.HostPort,
		})
	}
	return exposedPorts, portBindings, nil
}

func containerMounts(volumeMounts []VolumeMount) ([]mount.Mount, error) {
	if len(volumeMounts) == 0 {
		return nil, nil
	}

	mounts := make([]mount.Mount, 0, len(volumeMounts))
	for _, volumeMount := range volumeMounts {
		if volumeMount.Target == "" {
			return nil, fmt.Errorf("mount target is required")
		}

		mountType := volumeMount.Type
		if mountType == "" {
			mountType = string(mount.TypeBind)
		}
		if mountType != string(mount.TypeBind) && mountType != string(mount.TypeVolume) && mountType != string(mount.TypeTmpfs) {
			return nil, fmt.Errorf("unsupported mount type %q", mountType)
		}
		if mountType == string(mount.TypeTmpfs) && volumeMount.Source != "" {
			return nil, fmt.Errorf("tmpfs mount source must be empty")
		}
		if mountType != string(mount.TypeTmpfs) && volumeMount.Source == "" {
			return nil, fmt.Errorf("%s mount source is required", mountType)
		}

		mounts = append(mounts, mount.Mount{
			Type:     mount.Type(mountType),
			Source:   volumeMount.Source,
			Target:   volumeMount.Target,
			ReadOnly: volumeMount.ReadOnly,
		})
	}
	return mounts, nil
}

func containerRestartPolicy(policy RestartPolicy) (containerapi.RestartPolicy, error) {
	name := policy.Name
	if name == "" {
		name = string(containerapi.RestartPolicyDisabled)
	}
	restartPolicy := containerapi.RestartPolicy{
		Name:              containerapi.RestartPolicyMode(name),
		MaximumRetryCount: policy.MaximumRetryCount,
	}
	if err := containerapi.ValidateRestartPolicy(restartPolicy); err != nil {
		return containerapi.RestartPolicy{}, err
	}
	return restartPolicy, nil
}

func containerStopTimeoutSeconds(timeout *time.Duration) (int, error) {
	duration := defaultContainerStopTimeout
	if timeout != nil {
		duration = *timeout
	}
	if duration < 0 {
		return 0, fmt.Errorf("stop timeout cannot be negative")
	}

	seconds := int64(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	if seconds > int64(^uint(0)>>1) {
		return 0, fmt.Errorf("stop timeout is too large")
	}
	return int(seconds), nil
}

func containerFromInspect(response containerapi.InspectResponse) Container {
	container := Container{}
	if response.ContainerJSONBase != nil {
		container.ID = response.ID
		container.Name = strings.TrimPrefix(response.Name, "/")
		container.Command = strings.TrimSpace(strings.Join(append([]string{response.Path}, response.Args...), " "))
		container.Created = parseDockerTime(response.Created)
		if response.State != nil {
			container.State = ContainerState{
				Status:     string(response.State.Status),
				Running:    response.State.Running,
				Paused:     response.State.Paused,
				ExitCode:   response.State.ExitCode,
				StartedAt:  parseDockerTime(response.State.StartedAt),
				FinishedAt: parseDockerTime(response.State.FinishedAt),
				Error:      response.State.Error,
			}
		}
	}
	if response.Config != nil {
		container.Image = response.Config.Image
		container.Labels = response.Config.Labels
	}
	if response.NetworkSettings != nil {
		container.Ports = portBindingsFromDocker(response.NetworkSettings.Ports)
		container.NetworkIDs = networkIDsFromDocker(response.NetworkSettings.Networks)
	}
	return container
}

func containerFromSummary(summary containerapi.Summary) Container {
	name := ""
	if len(summary.Names) > 0 {
		name = strings.TrimPrefix(summary.Names[0], "/")
	}

	state := ContainerState{Status: summary.State}
	switch summary.State {
	case "running", "restarting":
		state.Running = true
	case "paused":
		state.Running = true
		state.Paused = true
	}

	ports := make([]PortBinding, len(summary.Ports))
	for i, port := range summary.Ports {
		ports[i] = PortBinding{
			HostIP:        port.IP,
			HostPort:      portNumber(port.PublicPort),
			ContainerPort: portNumber(port.PrivatePort),
			Protocol:      port.Type,
		}
	}
	sortPortBindings(ports)

	var networks map[string]*network.EndpointSettings
	if summary.NetworkSettings != nil {
		networks = summary.NetworkSettings.Networks
	}

	return Container{
		ID:         summary.ID,
		Name:       name,
		Image:      summary.Image,
		Command:    summary.Command,
		Created:    time.Unix(summary.Created, 0).UTC(),
		State:      state,
		Ports:      ports,
		Labels:     summary.Labels,
		NetworkIDs: networkIDsFromDocker(networks),
	}
}

func portBindingsFromDocker(portMap nat.PortMap) []PortBinding {
	if len(portMap) == 0 {
		return nil
	}

	ports := make([]PortBinding, 0, len(portMap))
	for port, bindings := range portMap {
		protocol, containerPort := nat.SplitProtoPort(string(port))
		if len(bindings) == 0 {
			ports = append(ports, PortBinding{ContainerPort: containerPort, Protocol: protocol})
			continue
		}
		for _, binding := range bindings {
			ports = append(ports, PortBinding{
				HostIP:        binding.HostIP,
				HostPort:      binding.HostPort,
				ContainerPort: containerPort,
				Protocol:      protocol,
			})
		}
	}
	sortPortBindings(ports)
	return ports
}

func sortPortBindings(ports []PortBinding) {
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].ContainerPort != ports[j].ContainerPort {
			return ports[i].ContainerPort < ports[j].ContainerPort
		}
		if ports[i].Protocol != ports[j].Protocol {
			return ports[i].Protocol < ports[j].Protocol
		}
		if ports[i].HostIP != ports[j].HostIP {
			return ports[i].HostIP < ports[j].HostIP
		}
		return ports[i].HostPort < ports[j].HostPort
	})
}

func networkIDsFromDocker(networks map[string]*network.EndpointSettings) []string {
	if len(networks) == 0 {
		return nil
	}

	ids := make([]string, 0, len(networks))
	for _, endpoint := range networks {
		if endpoint != nil && endpoint.NetworkID != "" {
			ids = append(ids, endpoint.NetworkID)
		}
	}
	sort.Strings(ids)
	return ids
}

func containerLogOptions(opts ContainerLogOptions) containerapi.LogsOptions {
	tail := opts.Tail
	if tail == "" {
		tail = "all"
	}

	return containerapi.LogsOptions{
		ShowStdout: opts.ShowStdout,
		ShowStderr: opts.ShowStderr,
		Follow:     opts.Follow,
		Tail:       tail,
		Since:      formatDockerTime(opts.Since),
		Until:      formatDockerTime(opts.Until),
		Timestamps: opts.Timestamps,
	}
}

func containerWaitCondition(condition WaitCondition) (containerapi.WaitCondition, error) {
	if condition == "" {
		return containerapi.WaitConditionNotRunning, nil
	}
	switch condition {
	case WaitConditionNotRunning:
		return containerapi.WaitConditionNotRunning, nil
	case WaitConditionNextExit:
		return containerapi.WaitConditionNextExit, nil
	case WaitConditionRemoved:
		return containerapi.WaitConditionRemoved, nil
	default:
		return "", fmt.Errorf("unsupported wait condition %q", condition)
	}
}

func containerNotModified(err error) bool {
	return err != nil && errdefs.IsNotModified(err)
}

func containerAlreadyExists(err error) bool {
	return err != nil && (errdefs.IsAlreadyExists(err) || (errdefs.IsConflict(err) && strings.Contains(strings.ToLower(err.Error()), "already in use")))
}

func containerIsRunning(err error) bool {
	return err != nil && errdefs.IsConflict(err) && strings.Contains(strings.ToLower(err.Error()), "is running")
}

func containerNetworkingConfig(networks []string) (*network.NetworkingConfig, []string) {
	uniqueNetworks := make([]string, 0, len(networks))
	seen := make(map[string]struct{}, len(networks))
	for _, networkName := range networks {
		if _, ok := seen[networkName]; ok {
			continue
		}
		seen[networkName] = struct{}{}
		uniqueNetworks = append(uniqueNetworks, networkName)
	}
	if len(uniqueNetworks) == 0 {
		return nil, nil
	}
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			uniqueNetworks[0]: {},
		},
	}, uniqueNetworks[1:]
}

func validateContainerNetworks(networks []string) error {
	for _, networkName := range networks {
		if networkName == "" {
			return fmt.Errorf("network name cannot be empty")
		}
	}
	return nil
}

func cloneDuration(timeout *time.Duration) *time.Duration {
	if timeout == nil {
		return nil
	}
	clone := *timeout
	return &clone
}

func parseDockerTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func formatDockerTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func portNumber(port uint16) string {
	if port == 0 {
		return ""
	}
	return fmt.Sprintf("%d", port)
}
