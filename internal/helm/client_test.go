package helm

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	yarerrors "github.com/yar-run/yar/internal/errors"
	"github.com/yar-run/yar/internal/kubernetes"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"io"
)

func TestClient_ReleaseLifecycle(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	release := Release{Name: "backend-api", Namespace: "development", Project: "backend", Environment: "dev"}

	installed, err := client.Install(context.Background(), chart, release, map[string]any{"message": "installed"}, InstallOptions{})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if diff := cmp.Diff(releaseLabels(release), installed.Labels); diff != "" {
		t.Errorf("Install().Labels mismatch (-want +got):\n%s", diff)
	}
	if _, err := client.Install(context.Background(), chart, release, nil, InstallOptions{}); err == nil {
		t.Error("Install() duplicate error = nil, want non-nil")
	}

	upgraded, err := client.Upgrade(context.Background(), chart, release, map[string]any{"message": "upgraded"}, UpgradeOptions{})
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}
	if got, want := upgraded.Version, 2; got != want {
		t.Errorf("Upgrade().Version = %d, want %d", got, want)
	}
	if err := client.Uninstall(context.Background(), release, UninstallOptions{}); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if err := client.Uninstall(context.Background(), release, UninstallOptions{}); err != nil {
		t.Errorf("Uninstall() missing release error = %v, want nil", err)
	}
}

func TestClient_ReleaseLifecycleRefusesExternalOwnership(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		labels    map[string]string
		namespace string
	}{
		"external": {
			labels: map[string]string{"app": "external"},
		},
		"partial Yar labels": {
			labels: map[string]string{labelProject: "backend"},
		},
		"other project": {
			labels: releaseLabels(Release{Project: "other", Environment: "dev"}),
		},
		"other environment": {
			labels: releaseLabels(Release{Project: "backend", Environment: "prod"}),
		},
		"other namespace": {
			labels:    releaseLabels(Release{Project: "backend", Environment: "dev"}),
			namespace: "other",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := newMemoryClient(t)
			chart, err := client.LoadChart(chartFixturePath(t, "chart"))
			if err != nil {
				t.Fatalf("LoadChart() error = %v", err)
			}
			release := Release{Name: "external-api", Namespace: "development", Project: "backend", Environment: "dev"}
			namespace := tc.namespace
			if namespace == "" {
				namespace = release.Namespace
			}
			seedRelease(t, client, release.Name, namespace, tc.labels)

			_, err = client.Upgrade(context.Background(), chart, release, nil, UpgradeOptions{})
			assertOwnershipError(t, err, "helm.upgrade")
			err = client.Uninstall(context.Background(), release, UninstallOptions{})
			assertOwnershipError(t, err, "helm.uninstall")
		})
	}
}

func TestClient_LifecycleDryRunDoesNotCommitRelease(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	releaseID := Release{Name: "dry-run-api", Namespace: "development", Project: "backend", Environment: "dev"}
	if _, err := client.Install(context.Background(), chart, releaseID, nil, InstallOptions{DryRun: true}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := client.release(releaseID); err == nil {
		t.Error("dry-run release exists, want absent")
	}

	seedRelease(t, client, "dry-run-upgrade", releaseID.Namespace, releaseLabels(releaseID))
	upgradeRelease := releaseID
	upgradeRelease.Name = "dry-run-upgrade"
	if _, err := client.Upgrade(context.Background(), chart, upgradeRelease, map[string]any{"message": "changed"}, UpgradeOptions{DryRun: true}); err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}
	stored, err := client.release(upgradeRelease)
	if err != nil {
		t.Fatalf("release() error = %v", err)
	}
	if got, want := stored.Version, 1; got != want {
		t.Errorf("dry-run Upgrade() stored version = %d, want %d", got, want)
	}

	if err := client.Uninstall(context.Background(), upgradeRelease, UninstallOptions{DryRun: true}); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	stored, err = client.release(upgradeRelease)
	if err != nil {
		t.Fatalf("release() error = %v", err)
	}
	if got, want := stored.Info.Status, release.StatusDeployed; got != want {
		t.Errorf("dry-run Uninstall() status = %q, want %q", got, want)
	}
}

func TestClient_LifecycleRejectsCanceledContext(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Install(ctx, chart, Release{Name: "cancelled-api", Namespace: "development", Project: "backend", Environment: "dev"}, nil, InstallOptions{})
	if !stderrors.Is(err, context.Canceled) {
		t.Errorf("Install() error = %v, want context.Canceled", err)
	}
}

func TestClient_RejectsInvalidRelease(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	_, err = client.Install(context.Background(), chart, Release{Name: "api", Namespace: "development", Project: "", Environment: "dev"}, nil, InstallOptions{})
	if err == nil {
		t.Fatal("Install() error = nil, want non-nil")
	}
	var validationError *yarerrors.ValidationError
	if !stderrors.As(err, &validationError) || validationError.Field != "release.project" {
		t.Errorf("Install() error = %#v, want release.project ValidationError", err)
	}
}

func TestClient_DefaultTimeout(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t, WithTimeout(0))
	if got, want := client.timeout, defaultTimeout; got != want {
		t.Errorf("client timeout = %v, want %v", got, want)
	}
	if _, err := NewClient(WithTimeout(-time.Second)); err == nil {
		t.Error("NewClient() negative timeout error = nil, want non-nil")
	}
}

func TestClient_LifecycleRequiresYarKubernetesClient(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	_, err = client.Install(context.Background(), chart, Release{Name: "api", Namespace: "development", Project: "backend", Environment: "dev"}, nil, InstallOptions{})
	if err == nil {
		t.Fatal("Install() error = nil, want non-nil")
	}
	var validationError *yarerrors.ValidationError
	if !stderrors.As(err, &validationError) || validationError.Field != "kubernetes" {
		t.Errorf("Install() error = %#v, want kubernetes ValidationError", err)
	}
}

func TestClient_AcceptsYarKubernetesClient(t *testing.T) {
	t.Parallel()

	server := newHelmVersionServer(t)
	kubeconfig := writeHelmKubeconfig(t, server)
	kubernetesClient, err := kubernetes.NewClient(kubernetes.WithKubeconfig(kubeconfig))
	if err != nil {
		t.Fatalf("kubernetes.NewClient() error = %v", err)
	}
	configured, err := NewClient(WithKubernetesClient(kubernetesClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	concreteClient := configured.(*client)
	if concreteClient.config == nil {
		t.Error("NewClient() configuration = nil, want non-nil")
	}
}

func TestClient_LifecycleUsesSelectedKubernetesNamespace(t *testing.T) {
	t.Parallel()

	server := newHelmVersionServer(t)
	kubeconfig := writeHelmKubeconfig(t, server)
	kubernetesClient, err := kubernetes.NewClient(kubernetes.WithKubeconfig(kubeconfig))
	if err != nil {
		t.Fatalf("kubernetes.NewClient() error = %v", err)
	}
	configured, err := NewClient(WithKubernetesClient(kubernetesClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client := configured.(*client)
	chart, err := client.LoadChart(chartFixturePath(t, "chart"))
	if err != nil {
		t.Fatalf("LoadChart() error = %v", err)
	}
	_, err = client.Install(context.Background(), chart, Release{Name: "api", Namespace: "other", Project: "backend", Environment: "dev"}, nil, InstallOptions{})
	if err == nil {
		t.Fatal("Install() error = nil, want non-nil")
	}
	var validationError *yarerrors.ValidationError
	if !stderrors.As(err, &validationError) || validationError.Field != "release.namespace" {
		t.Errorf("Install() error = %#v, want release.namespace ValidationError", err)
	}
}

func TestClient_LifecycleRedactsTemplateFailures(t *testing.T) {
	t.Parallel()

	client := newMemoryClient(t)
	chart, err := client.LoadChartFiles([]ChartFile{
		{Name: "Chart.yaml", Data: []byte("apiVersion: v2\nname: failing-chart\nversion: 0.1.0\n")},
		{Name: "templates/configmap.yaml", Data: []byte("{{ fail .Values.password }}")},
	})
	if err != nil {
		t.Fatalf("LoadChartFiles() error = %v", err)
	}
	const secret = "literal-secret-must-not-appear-in-lifecycle-error"
	_, err = client.Install(context.Background(), chart, Release{Name: "api", Namespace: "development", Project: "backend", Environment: "dev"}, map[string]any{"password": secret}, InstallOptions{})
	if err == nil {
		t.Fatal("Install() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("Install() error leaked secret: %q", err)
	}
	var kubernetesError *yarerrors.KubernetesError
	if !stderrors.As(err, &kubernetesError) || kubernetesError.Unwrap() == nil {
		t.Errorf("Install() error = %#v, want KubernetesError with redacted cause", err)
	}
}

func newHelmVersionServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`{"gitVersion":"v1.34.0"}`))
	}))
	t.Cleanup(server.Close)
	return server
}

func writeHelmKubeconfig(t *testing.T, server *httptest.Server) string {
	t.Helper()

	certificate, err := x509.ParseCertificate(server.Certificate().Raw)
	if err != nil {
		t.Fatalf("x509.ParseCertificate() error = %v", err)
	}
	caData := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw}))
	contents := "apiVersion: v1\nkind: Config\ncurrent-context: default\nclusters:\n- name: default\n  cluster:\n    server: " + server.URL + "\n    certificate-authority-data: " + caData + "\ncontexts:\n- name: default\n  context:\n    cluster: default\n    user: default\n    namespace: development\nusers:\n- name: default\n  user:\n    token: token\n"
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	return path
}

func newTestClient(t *testing.T, opts ...Option) *client {
	t.Helper()

	configured, err := NewClient(opts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client := configured.(*client)
	return client
}

func newMemoryClient(t *testing.T, opts ...Option) *client {
	t.Helper()

	client := newTestClient(t, opts...)
	client.config = &action.Configuration{
		Releases:     storage.Init(driver.NewMemory()),
		KubeClient:   &fake.FailingKubeClient{PrintingKubeClient: fake.PrintingKubeClient{Out: io.Discard, LogOutput: io.Discard}},
		Capabilities: chartutil.DefaultCapabilities.Copy(),
		Log:          func(string, ...interface{}) {},
	}
	return client
}

func seedRelease(t *testing.T, client *client, name, namespace string, labels map[string]string) {
	t.Helper()

	release := &release.Release{
		Name:      name,
		Namespace: namespace,
		Version:   1,
		Info:      &release.Info{Status: release.StatusDeployed},
		Labels:    labels,
	}
	if err := client.config.Releases.Create(release); err != nil {
		t.Fatalf("Releases.Create() error = %v", err)
	}
}

func assertOwnershipError(t *testing.T, err error, wantOp string) {
	t.Helper()

	if err == nil {
		t.Fatal("operation error = nil, want ownership error")
	}
	var kubernetesError *yarerrors.KubernetesError
	if !stderrors.As(err, &kubernetesError) || kubernetesError.Op != wantOp {
		t.Errorf("operation error = %#v, want %s KubernetesError", err, wantOp)
	}
}
