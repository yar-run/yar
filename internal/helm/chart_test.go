package helm

import (
	"context"
	stderrors "errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	yarerrors "github.com/yar-run/yar/internal/errors"
)

func TestClient_ImplementsClient(t *testing.T) {
	t.Parallel()

	var _ Client = (*client)(nil)
}

func TestClient_LoadChart(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	if got, want := chart.Name(), "yar-test-chart"; got != want {
		t.Errorf("LoadChart().Name() = %q, want %q", got, want)
	}
}

func TestClient_LoadChartFiles(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChartFiles([]ChartFile{
		{Name: "Chart.yaml", Data: []byte("apiVersion: v2\nname: memory-chart\nversion: 0.1.0\n")},
		{Name: "values.yaml", Data: []byte("message: default\n")},
		{Name: "templates/configmap.yaml", Data: []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ .Release.Name }}\n")},
	})
	if err != nil {
		t.Fatalf("LoadChartFiles() error = %v", err)
	}
	if got, want := chart.Name(), "memory-chart"; got != want {
		t.Errorf("LoadChartFiles().Name() = %q, want %q", got, want)
	}
}

func TestClient_LoadChartRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	tests := map[string]struct {
		load  func() error
		check func(t *testing.T, err error)
	}{
		"missing chart": {
			load: func() error {
				_, err := client.LoadChart(filepath.Join(t.TempDir(), "missing"))
				return err
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				var notFound *yarerrors.NotFoundError
				if !stderrors.As(err, &notFound) || notFound.Resource != "chart" {
					t.Errorf("LoadChart() error = %#v, want chart NotFoundError", err)
				}
			},
		},
		"invalid chart": {
			load: func() error {
				_, err := client.LoadChart(chartFixturePath(t, "invalid-chart"))
				return err
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				var packError *yarerrors.PackError
				if !stderrors.As(err, &packError) || packError.Err == nil {
					t.Errorf("LoadChart() error = %#v, want PackError with cause", err)
				}
			},
		},
		"invalid chart files": {
			load: func() error {
				_, err := client.LoadChartFiles([]ChartFile{{Name: "templates/configmap.yaml", Data: []byte("invalid")}})
				return err
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				var packError *yarerrors.PackError
				if !stderrors.As(err, &packError) || packError.Err == nil {
					t.Errorf("LoadChartFiles() error = %#v, want PackError with cause", err)
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.load()
			if err == nil {
				t.Fatal("load error = nil, want non-nil")
			}
			tc.check(t, err)
		})
	}
}

func TestClient_RenderIsDeterministicAndSecretSafe(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	values := map[string]any{
		"message":  "override-message",
		"password": "literal-secret-must-not-be-rendered",
		"nested":   map[string]any{"key": "original"},
	}
	release := Release{Name: "backend-api", Namespace: "development", Project: "backend", Environment: "dev"}
	opts := RenderOptions{KubeVersion: "v1.34.0", APIVersions: []string{"yar.io/v1/TestCapability"}}

	first, err := client.Render(context.Background(), chart, release, values, opts)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	second, err := client.Render(context.Background(), chart, release, values, opts)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if diff := cmp.Diff(first, second); diff != "" {
		t.Errorf("Render() results differ (-want +got):\n%s", diff)
	}
	if !strings.Contains(first.Manifest, "override-message") || !strings.Contains(first.Manifest, "capability: enabled") {
		t.Errorf("Render().Manifest = %q, want overridden values and capability output", first.Manifest)
	}
	if strings.Contains(first.Manifest, "literal-secret-must-not-be-rendered") || strings.Contains(first.Manifest, "kind: Secret") {
		t.Errorf("Render().Manifest exposed secret output: %q", first.Manifest)
	}
	if got, want := first.Notes, "Release backend-api is ready.\n"; got != want {
		t.Errorf("Render().Notes = %q, want %q", got, want)
	}
	if got, want := values["message"], "override-message"; got != want {
		t.Errorf("Render() mutated caller values: message = %v, want %q", got, want)
	}
	if got, want := values["password"], "literal-secret-must-not-be-rendered"; got != want {
		t.Errorf("Render() mutated caller values: password = %v, want %q", got, want)
	}
	if got, want := values["nested"].(map[string]any)["key"], "original"; got != want {
		t.Errorf("Render() mutated nested caller values: key = %v, want %q", got, want)
	}
}

func TestClient_RenderDoesNotWriteOutput(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	before, err := os.ReadDir(t.TempDir())
	if err != nil {
		t.Fatalf("os.ReadDir() error = %v", err)
	}
	_, err = client.Render(context.Background(), chart, Release{Name: "backend-api", Namespace: "development", Project: "backend", Environment: "dev"}, nil, RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	after, err := os.ReadDir(t.TempDir())
	if err != nil {
		t.Fatalf("os.ReadDir() error = %v", err)
	}
	if diff := cmp.Diff(before, after); diff != "" {
		t.Errorf("render output directory changed (-want +got):\n%s", diff)
	}
}

func TestClient_RenderRejectsInvalidReleaseWithoutSecretLeakage(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	const secret = "literal-secret-must-not-appear-in-error"
	_, err = client.Render(context.Background(), chart, Release{Name: "Invalid_Release", Namespace: "development", Project: "backend", Environment: "dev"}, map[string]any{"password": secret}, RenderOptions{})
	if err == nil {
		t.Fatal("Render() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("Render() error leaked secret: %q", err)
	}
	var validationError *yarerrors.ValidationError
	if !stderrors.As(err, &validationError) || validationError.Field != "release.name" {
		t.Errorf("Render() error = %#v, want release.name ValidationError", err)
	}
}

func TestClient_RenderRedactsTemplateFailures(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChartFiles([]ChartFile{
		{Name: "Chart.yaml", Data: []byte("apiVersion: v2\nname: failing-chart\nversion: 0.1.0\n")},
		{Name: "templates/configmap.yaml", Data: []byte("{{ fail .Values.password }}")},
	})
	if err != nil {
		t.Fatalf("LoadChartFiles() error = %v", err)
	}
	const secret = "literal-secret-must-not-appear-in-render-error"
	_, err = client.Render(context.Background(), chart, Release{Name: "api", Namespace: "development", Project: "backend", Environment: "dev"}, map[string]any{"password": secret}, RenderOptions{})
	if err == nil {
		t.Fatal("Render() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("Render() error leaked secret: %q", err)
	}
}

func TestClient_RenderRejectsUnsafeValueShapeWithoutLoggingValues(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	const secret = "literal-secret-must-not-appear-in-shape-error"
	_, err = client.Render(context.Background(), chart, Release{Name: "api", Namespace: "development", Project: "backend", Environment: "dev"}, map[string]any{
		"password": map[string]any{"nested": secret},
	}, RenderOptions{})
	if err == nil {
		t.Fatal("Render() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("Render() error leaked secret: %q", err)
	}
	var validationError *yarerrors.ValidationError
	if !stderrors.As(err, &validationError) || validationError.Field != "values.password" {
		t.Errorf("Render() error = %#v, want values.password ValidationError", err)
	}
}

func chartFixturePath(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join("testdata", name)
}
