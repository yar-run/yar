package kubernetes

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHelmRESTClientGetterUsesSelectedClientWithoutChangingKubeconfig(t *testing.T) {
	t.Parallel()

	server := newVersionServer(t, "token", 200, `{"gitVersion":"v1.34.0"}`)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token", namespace: "selected-namespace"},
	}, "default")
	kubeconfigBefore, err := os.ReadFile(kubeconfig)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	configured, err := NewClient(WithKubeconfig(kubeconfig))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	getter, err := HelmRESTClientGetter(configured)
	if err != nil {
		t.Fatalf("HelmRESTClientGetter() error = %v", err)
	}
	restConfig, err := getter.ToRESTConfig()
	if err != nil {
		t.Fatalf("ToRESTConfig() error = %v", err)
	}
	if got, want := restConfig.Host, server.URL; got != want {
		t.Errorf("ToRESTConfig().Host = %q, want %q", got, want)
	}
	namespace, _, err := getter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		t.Fatalf("ToRawKubeConfigLoader().Namespace() error = %v", err)
	}
	if got, want := namespace, "selected-namespace"; got != want {
		t.Errorf("getter namespace = %q, want %q", got, want)
	}

	assertKubeconfigUnchanged(t, kubeconfig, kubeconfigBefore)
	if diff := cmp.Diff(kubeconfigBefore, mustReadFile(t, kubeconfig)); diff != "" {
		t.Errorf("kubeconfig changed (-want +got):\n%s", diff)
	}
}

func TestHelmRESTClientGetterRejectsUnsupportedClient(t *testing.T) {
	t.Parallel()

	if _, err := HelmRESTClientGetter(nil); err == nil {
		t.Error("HelmRESTClientGetter(nil) error = nil, want non-nil")
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	return contents
}
