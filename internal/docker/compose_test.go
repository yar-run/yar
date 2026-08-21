package docker

import (
	"context"
	stderrors "errors"
	"os"
	"path/filepath"
	"testing"

	composetypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/google/go-cmp/cmp"
	yarerrors "github.com/yar-run/yar/internal/errors"
)

func TestLoadCompose_ValidProject(t *testing.T) {
	t.Parallel()

	project, err := LoadCompose(context.Background(), composeFixturePath(t, "valid.yaml"))
	if err != nil {
		t.Fatalf("LoadCompose() error = %v", err)
	}

	if got, want := project.Name, "yar-compose-fixture"; got != want {
		t.Errorf("Project.Name = %q, want %q", got, want)
	}
	if diff := cmp.Diff([]string{"app", "redis"}, project.ServiceNames()); diff != "" {
		t.Errorf("Project.ServiceNames() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"app-net"}, project.NetworkNames()); diff != "" {
		t.Errorf("Project.NetworkNames() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"redis-data"}, project.VolumeNames()); diff != "" {
		t.Errorf("Project.VolumeNames() mismatch (-want +got):\n%s", diff)
	}

	redis, err := project.GetService("redis")
	if err != nil {
		t.Fatalf("Project.GetService(redis) error = %v", err)
	}
	if len(redis.Ports) != 1 {
		t.Fatalf("redis ports = %d, want 1", len(redis.Ports))
	}
	wantRedisPort := composetypes.ServicePortConfig{HostIP: "127.0.0.1", Target: 6379, Published: "6379", Protocol: "tcp", Mode: "ingress"}
	if diff := cmp.Diff(wantRedisPort, redis.Ports[0]); diff != "" {
		t.Errorf("redis port mismatch (-want +got):\n%s", diff)
	}
	if redis.HealthCheck == nil {
		t.Fatal("redis healthcheck = nil, want non-nil")
	}
	if diff := cmp.Diff(composetypes.HealthCheckTest{"CMD", "redis-cli", "ping"}, redis.HealthCheck.Test); diff != "" {
		t.Errorf("redis healthcheck test mismatch (-want +got):\n%s", diff)
	}

	app, err := project.GetService("app")
	if err != nil {
		t.Fatalf("Project.GetService(app) error = %v", err)
	}
	if app.Build == nil {
		t.Fatal("app build = nil, want non-nil")
	}
	fixtureDir, err := filepath.Abs(filepath.Dir(composeFixturePath(t, "valid.yaml")))
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
	if got, want := app.Build.Context, filepath.Join(fixtureDir, "app"); got != want {
		t.Errorf("app build context = %q, want %q", got, want)
	}
	if len(app.Volumes) != 1 {
		t.Fatalf("app volumes = %d, want 1", len(app.Volumes))
	}
	if got, want := app.Volumes[0].Source, filepath.Join(fixtureDir, "config"); got != want {
		t.Errorf("app bind source = %q, want %q", got, want)
	}
	dependency, ok := app.DependsOn["redis"]
	if !ok {
		t.Fatal("app dependency on redis is missing")
	}
	if got, want := dependency.Condition, composetypes.ServiceConditionHealthy; got != want {
		t.Errorf("app dependency condition = %q, want %q", got, want)
	}
}

func TestLoadCompose_MergesOverrideFiles(t *testing.T) {
	t.Parallel()

	project, err := LoadCompose(
		context.Background(),
		composeFixturePath(t, "valid.yaml"),
		composeFixturePath(t, "override.yaml"),
	)
	if err != nil {
		t.Fatalf("LoadCompose() error = %v", err)
	}

	app, err := project.GetService("app")
	if err != nil {
		t.Fatalf("Project.GetService(app) error = %v", err)
	}
	if got, want := app.Image, "ghcr.io/acme/app:override"; got != want {
		t.Errorf("app image = %q, want %q", got, want)
	}
	if got := app.Environment["LOG_LEVEL"]; got == nil || *got != "debug" {
		t.Errorf("app environment LOG_LEVEL = %v, want %q", got, "debug")
	}
	if diff := cmp.Diff([]composetypes.ServicePortConfig{
		{Target: 8080, Published: "8080", Protocol: "tcp", Mode: "ingress"},
		{Target: 8080, Published: "9090", Protocol: "tcp", Mode: "ingress"},
	}, app.Ports); diff != "" {
		t.Errorf("app ports mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadCompose_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		paths []string
	}{
		"no paths": {},
		"missing file": {
			paths: []string{composeFixturePath(t, "missing.yaml")},
		},
		"invalid YAML": {
			paths: []string{composeFixturePath(t, "invalid-yaml.yaml")},
		},
		"invalid schema": {
			paths: []string{composeFixturePath(t, "invalid-schema.yaml")},
		},
		"undefined dependency": {
			paths: []string{composeFixturePath(t, "invalid-consistency.yaml")},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := LoadCompose(context.Background(), tc.paths...)
			if err == nil {
				t.Fatal("LoadCompose() error = nil, want non-nil")
			}
			var configErr *yarerrors.ConfigError
			if !stderrors.As(err, &configErr) {
				t.Fatalf("LoadCompose() error type = %T, want *errors.ConfigError", err)
			}
			if configErr.Path == "" {
				t.Error("ConfigError.Path = empty, want non-empty")
			}
			if configErr.Message == "" {
				t.Error("ConfigError.Message = empty, want non-empty")
			}
			if name != "no paths" && configErr.Err == nil {
				t.Error("ConfigError.Err = nil, want underlying load error")
			}
		})
	}
}

func TestLoadCompose_DoesNotResolveAmbientEnvironmentOrFiles(t *testing.T) {
	project, err := LoadCompose(context.Background(), composeFixturePath(t, "literal-interpolation.yaml"))
	if err != nil {
		t.Fatalf("LoadCompose() error = %v", err)
	}
	app, err := project.GetService("app")
	if err != nil {
		t.Fatalf("Project.GetService(app) error = %v", err)
	}
	if got, want := app.Image, "example:${YAR_COMPOSE_TEST_VALUE}"; got != want {
		t.Errorf("app image = %q, want literal %q", got, want)
	}
	if got := app.Environment["FROM_ENV"]; got == nil || *got != "${YAR_COMPOSE_TEST_VALUE}" {
		t.Errorf("app environment FROM_ENV = %v, want literal %q", got, "${YAR_COMPOSE_TEST_VALUE}")
	}
	fixtureDir, err := filepath.Abs(filepath.Dir(composeFixturePath(t, "literal-interpolation.yaml")))
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
	if len(app.EnvFiles) != 1 || app.EnvFiles[0].Path != filepath.Join(fixtureDir, "does-not-exist.env") {
		t.Errorf("app env files = %#v, want unresolved env file", app.EnvFiles)
	}
	if diff := cmp.Diff([]string{filepath.Join(fixtureDir, "does-not-exist.labels")}, app.LabelFiles); diff != "" {
		t.Errorf("app label files mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadCompose_DoesNotReadDotEnv(t *testing.T) {
	projectDir := t.TempDir()
	composePath := filepath.Join(projectDir, "compose.yaml")
	if err := os.WriteFile(composePath, []byte("services:\n  app:\n    image: example:${YAR_COMPOSE_TEST_VALUE}\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(compose.yaml) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".env"), []byte("YAR_COMPOSE_TEST_VALUE=secret-from-dotenv\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(.env) error = %v", err)
	}
	t.Setenv("YAR_COMPOSE_TEST_VALUE", "secret-from-process-environment")

	project, err := LoadCompose(context.Background(), composePath)
	if err != nil {
		t.Fatalf("LoadCompose() error = %v", err)
	}
	app, err := project.GetService("app")
	if err != nil {
		t.Fatalf("Project.GetService(app) error = %v", err)
	}
	if got, want := app.Image, "example:${YAR_COMPOSE_TEST_VALUE}"; got != want {
		t.Errorf("app image = %q, want literal %q", got, want)
	}
}

func TestLoadCompose_PropagatesCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := LoadCompose(ctx, composeFixturePath(t, "valid.yaml"))
	if !stderrors.Is(err, context.Canceled) {
		t.Errorf("LoadCompose() error = %v, want context.Canceled", err)
	}
	var configErr *yarerrors.ConfigError
	if stderrors.As(err, &configErr) {
		t.Errorf("LoadCompose() wrapped context cancellation in ConfigError: %#v", configErr)
	}
}

func composeFixturePath(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join("testdata", "compose", name)
}
