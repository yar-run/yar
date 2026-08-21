package kubernetes

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	yarerrors "github.com/yar-run/yar/internal/errors"
	"k8s.io/apimachinery/pkg/api/errors"
)

func TestClient_ImplementsClient(t *testing.T) {
	t.Parallel()

	var _ Client = (*client)(nil)
}

func TestNewClient_SelectsContextAndNamespace(t *testing.T) {
	t.Parallel()

	firstServer := newVersionServer(t, "first-token", http.StatusOK, `{"major":"1","minor":"34","gitVersion":"v1.34.0"}`)
	secondServer := newVersionServer(t, "second-token", http.StatusOK, `{"major":"1","minor":"34","gitVersion":"v1.34.1"}`)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"first":  {server: firstServer, token: "first-token", namespace: "first-namespace"},
		"second": {server: secondServer, token: "second-token", namespace: "second-namespace"},
	}, "first")

	tests := map[string]struct {
		opts          []Option
		wantContext   string
		wantNamespace string
		wantVersion   string
	}{
		"uses kubeconfig current context": {
			opts:          []Option{WithKubeconfig(kubeconfig)},
			wantContext:   "first",
			wantNamespace: "first-namespace",
			wantVersion:   "v1.34.0",
		},
		"selects context in memory": {
			opts:          []Option{WithKubeconfig(kubeconfig), WithContext("second")},
			wantContext:   "second",
			wantNamespace: "second-namespace",
			wantVersion:   "v1.34.1",
		},
		"explicit namespace wins": {
			opts:          []Option{WithKubeconfig(kubeconfig), WithContext("second"), WithNamespace("yar-system")},
			wantContext:   "second",
			wantNamespace: "yar-system",
			wantVersion:   "v1.34.1",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kubeconfigBefore, err := os.ReadFile(kubeconfig)
			if err != nil {
				t.Fatalf("os.ReadFile() error = %v", err)
			}

			client, err := NewClient(tc.opts...)
			if err != nil {
				t.Fatalf("NewClient() error = %+v", err)
			}
			if got := client.Context(); got != tc.wantContext {
				t.Errorf("Context() = %q, want %q", got, tc.wantContext)
			}
			if got := client.Namespace(); got != tc.wantNamespace {
				t.Errorf("Namespace() = %q, want %q", got, tc.wantNamespace)
			}

			version, err := client.Probe(context.Background())
			if err != nil {
				t.Fatalf("Probe() error = %v", err)
			}
			if got := version.GitVersion; got != tc.wantVersion {
				t.Errorf("Probe().GitVersion = %q, want %q", got, tc.wantVersion)
			}

			assertKubeconfigUnchanged(t, kubeconfig, kubeconfigBefore)
		})
	}
}

func TestNewClient_DefaultsNamespace(t *testing.T) {
	t.Parallel()

	server := newVersionServer(t, "token", http.StatusOK, `{"gitVersion":"v1.34.0"}`)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "default")

	client, err := NewClient(WithKubeconfig(kubeconfig))
	if err != nil {
		t.Fatalf("NewClient() error = %#v", err)
	}
	if got, want := client.Namespace(), "default"; got != want {
		t.Errorf("Namespace() = %q, want %q", got, want)
	}
}

func TestNewClient_RejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	server := newVersionServer(t, "token", http.StatusOK, `{"gitVersion":"v1.34.0"}`)
	validKubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "default")
	invalidKubeconfig := filepath.Join(t.TempDir(), "invalid-kubeconfig")
	if err := os.WriteFile(invalidKubeconfig, []byte("not: [valid"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	tests := map[string]struct {
		opts       []Option
		checkError func(t *testing.T, err error)
	}{
		"missing explicit kubeconfig": {
			opts: []Option{WithKubeconfig(filepath.Join(t.TempDir(), "missing"))},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var notFound *yarerrors.NotFoundError
				if !stderrors.As(err, &notFound) || notFound.Resource != "kubeconfig" {
					t.Errorf("NewClient() error = %#v, want kubeconfig NotFoundError", err)
				}
			},
		},
		"missing selected context": {
			opts: []Option{WithKubeconfig(validKubeconfig), WithContext("missing")},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var notFound *yarerrors.NotFoundError
				if !stderrors.As(err, &notFound) || notFound.Resource != "kubeconfig context" || notFound.Name != "missing" {
					t.Errorf("NewClient() error = %#v, want missing context NotFoundError", err)
				}
			},
		},
		"invalid namespace": {
			opts: []Option{WithKubeconfig(validKubeconfig), WithNamespace("Invalid_Namespace")},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var validationError *yarerrors.ValidationError
				if !stderrors.As(err, &validationError) || validationError.Field != "namespace" {
					t.Errorf("NewClient() error = %#v, want namespace ValidationError", err)
				}
			},
		},
		"invalid owner project": {
			opts: []Option{WithKubeconfig(validKubeconfig), WithOwner("invalid/project/too/many/slashes", "local")},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var validationError *yarerrors.ValidationError
				if !stderrors.As(err, &validationError) || validationError.Field != "owner.project" {
					t.Errorf("NewClient() error = %#v, want owner.project ValidationError", err)
				}
			},
		},
		"empty field manager": {
			opts: []Option{WithKubeconfig(validKubeconfig), WithFieldManager("")},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var validationError *yarerrors.ValidationError
				if !stderrors.As(err, &validationError) || validationError.Field != "fieldManager" {
					t.Errorf("NewClient() error = %#v, want fieldManager ValidationError", err)
				}
			},
		},
		"negative timeout": {
			opts: []Option{WithKubeconfig(validKubeconfig), WithTimeout(-time.Second)},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var validationError *yarerrors.ValidationError
				if !stderrors.As(err, &validationError) || validationError.Field != "timeout" {
					t.Errorf("NewClient() error = %#v, want timeout ValidationError", err)
				}
			},
		},
		"invalid kubeconfig": {
			opts: []Option{WithKubeconfig(invalidKubeconfig)},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				var configError *yarerrors.ConfigError
				if !stderrors.As(err, &configError) || configError.Err == nil {
					t.Errorf("NewClient() error = %#v, want ConfigError with cause", err)
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := NewClient(tc.opts...)
			if err == nil {
				t.Fatal("NewClient() error = nil, want non-nil")
			}
			tc.checkError(t, err)
		})
	}
}

func TestNewClient_ZeroTimeoutKeepsDefault(t *testing.T) {
	t.Parallel()

	server := newVersionServer(t, "token", http.StatusOK, `{"gitVersion":"v1.34.0"}`)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "default")

	configuredClient, err := NewClient(WithKubeconfig(kubeconfig), WithTimeout(0))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	concreteClient, ok := configuredClient.(*client)
	if !ok {
		t.Fatalf("NewClient() type = %T, want *client", configuredClient)
	}
	if got, want := concreteClient.timeout, defaultTimeout; got != want {
		t.Errorf("client timeout = %v, want %v", got, want)
	}
}

func TestNewClient_RejectsKubeconfigWithoutCurrentContext(t *testing.T) {
	t.Parallel()

	server := newVersionServer(t, "token", http.StatusOK, `{"gitVersion":"v1.34.0"}`)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "")

	_, err := NewClient(WithKubeconfig(kubeconfig))
	if err == nil {
		t.Fatal("NewClient() error = nil, want non-nil")
	}
	var notFound *yarerrors.NotFoundError
	if !stderrors.As(err, &notFound) || notFound.Resource != "kubeconfig context" || notFound.Name != "current" {
		t.Errorf("NewClient() error = %#v, want current-context NotFoundError", err)
	}
}

func TestClient_ProbeMapsFailures(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		statusCode int
		body       string
		check      func(t *testing.T, err error)
	}{
		"unauthorized": {
			statusCode: http.StatusUnauthorized,
			body:       `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"Unauthorized","code":401}`,
			check: func(t *testing.T, err error) {
				t.Helper()
				if !errors.IsUnauthorized(unwrapKubernetesError(t, err)) {
					t.Errorf("Probe() error = %v, want unauthorized API error", err)
				}
			},
		},
		"forbidden": {
			statusCode: http.StatusForbidden,
			body:       `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"Forbidden","code":403}`,
			check: func(t *testing.T, err error) {
				t.Helper()
				if !errors.IsForbidden(unwrapKubernetesError(t, err)) {
					t.Errorf("Probe() error = %v, want forbidden API error", err)
				}
			},
		},
		"service unavailable": {
			statusCode: http.StatusServiceUnavailable,
			body:       `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"ServiceUnavailable","code":503}`,
			check: func(t *testing.T, err error) {
				t.Helper()
				if !errors.IsServiceUnavailable(unwrapKubernetesError(t, err)) {
					t.Errorf("Probe() error = %v, want service unavailable API error", err)
				}
			},
		},
		"malformed version": {
			statusCode: http.StatusOK,
			body:       `{not-json}`,
			check: func(t *testing.T, err error) {
				t.Helper()
				if unwrapKubernetesError(t, err) == nil {
					t.Error("Probe() underlying error = nil, want non-nil")
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := newVersionServer(t, "token", tc.statusCode, tc.body)
			kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
				"default": {server: server, token: "token"},
			}, "default")
			client, err := NewClient(WithKubeconfig(kubeconfig))
			if err != nil {
				t.Fatalf("NewClient() error = %+v", err)
			}

			_, err = client.Probe(context.Background())
			if err == nil {
				t.Fatal("Probe() error = nil, want non-nil")
			}
			tc.check(t, err)
		})
	}
}

func TestClient_ProbeMapsTransportFailure(t *testing.T) {
	t.Parallel()

	server := newVersionServer(t, "token", http.StatusOK, `{"gitVersion":"v1.34.0"}`)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "default")
	client, err := NewClient(WithKubeconfig(kubeconfig))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	server.Close()

	_, err = client.Probe(context.Background())
	if err == nil {
		t.Fatal("Probe() error = nil, want non-nil")
	}
	if unwrapKubernetesError(t, err) == nil {
		t.Error("Probe() underlying error = nil, want non-nil")
	}
}

func TestClient_ProbePropagatesContext(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		close(started)
		<-request.Context().Done()
	}))
	t.Cleanup(server.Close)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "default")
	client, err := NewClient(WithKubeconfig(kubeconfig))
	if err != nil {
		t.Fatalf("NewClient() error = %+v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errC := make(chan error, 1)
	go func() {
		_, err := client.Probe(ctx)
		errC <- err
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("Probe() did not reach the server")
	}

	select {
	case err := <-errC:
		if !stderrors.Is(err, context.Canceled) {
			t.Errorf("Probe() error = %v, want context.Canceled", err)
		}
		var kubernetesError *yarerrors.KubernetesError
		if stderrors.As(err, &kubernetesError) {
			t.Errorf("Probe() wrapped context cancellation in KubernetesError: %#v", kubernetesError)
		}
	case <-time.After(time.Second):
		t.Fatal("Probe() did not return after cancellation")
	}
}

func TestClient_ProbeHonorsTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	t.Cleanup(server.Close)
	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token"},
	}, "default")
	client, err := NewClient(WithKubeconfig(kubeconfig), WithTimeout(25*time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient() error = %+v", err)
	}

	_, err = client.Probe(context.Background())
	if !stderrors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Probe() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestNewClient_DoesNotExposeKubeconfigCredentials(t *testing.T) {
	t.Parallel()

	const token = "secret-token-that-must-not-appear-in-errors"
	kubeconfig := filepath.Join(t.TempDir(), "invalid-config")
	contents := "apiVersion: v1\nkind: Config\nusers:\n  - name: default\n    user:\n      token: " + token + "\ncontexts: invalid\n"
	if err := os.WriteFile(kubeconfig, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err := NewClient(WithKubeconfig(kubeconfig))
	if err == nil {
		t.Fatal("NewClient() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), token) {
		t.Errorf("NewClient() error exposed token: %q", err)
	}
}

type kubeconfigCluster struct {
	server    *httptest.Server
	token     string
	namespace string
}

func newVersionServer(t *testing.T, wantToken string, statusCode int, body string) *httptest.Server {
	t.Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/version" {
			t.Errorf("request = %s %s, want GET /version", request.Method, request.URL.Path)
			responseWriter.WriteHeader(http.StatusNotFound)
			return
		}
		if got, want := request.Header.Get("Authorization"), "Bearer "+wantToken; got != want {
			t.Errorf("Authorization = %q, want %q", got, want)
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(statusCode)
		_, _ = responseWriter.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func writeKubeconfig(t *testing.T, clusters map[string]kubeconfigCluster, currentContext string) string {
	t.Helper()

	names := make([]string, 0, len(clusters))
	for name := range clusters {
		names = append(names, name)
	}
	sort.Strings(names)

	var contexts []string
	var clusterEntries []string
	var userEntries []string
	for _, name := range names {
		cluster := clusters[name]
		certificate, err := x509.ParseCertificate(cluster.server.Certificate().Raw)
		if err != nil {
			t.Fatalf("x509.ParseCertificate() error = %v", err)
		}
		caData := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw}))
		contexts = append(contexts, fmt.Sprintf("- name: %s\n  context:\n    cluster: %s\n    user: %s\n    namespace: %s", name, name, name, cluster.namespace))
		clusterEntries = append(clusterEntries, fmt.Sprintf("- name: %s\n  cluster:\n    server: %s\n    certificate-authority-data: %s", name, cluster.server.URL, caData))
		userEntries = append(userEntries, fmt.Sprintf("- name: %s\n  user:\n    token: %s", name, cluster.token))
	}

	contents := fmt.Sprintf("apiVersion: v1\nkind: Config\ncurrent-context: %s\nclusters:\n%s\ncontexts:\n%s\nusers:\n%s\n", currentContext, indentKubeconfigEntries(clusterEntries), indentKubeconfigEntries(contexts), indentKubeconfigEntries(userEntries))
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	return path
}

func indentKubeconfigEntries(entries []string) string {
	indented := make([]string, len(entries))
	for i, entry := range entries {
		indented[i] = "  " + strings.ReplaceAll(entry, "\n", "\n  ")
	}
	return strings.Join(indented, "\n")
}

func assertKubeconfigUnchanged(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("kubeconfig changed (-want +got):\n%s", diff)
	}
	if _, err := os.Stat(path + ".lock"); !os.IsNotExist(err) {
		t.Errorf("kubeconfig lock stat error = %v, want no lock file", err)
	}
}

func unwrapKubernetesError(t *testing.T, err error) error {
	t.Helper()

	var kubernetesError *yarerrors.KubernetesError
	if !stderrors.As(err, &kubernetesError) {
		t.Fatalf("error type = %T, want *errors.KubernetesError", err)
	}
	if kubernetesError.Op != "probe" || kubernetesError.Resource != "api-server" {
		t.Errorf("KubernetesError = %#v, want probe api-server error", kubernetesError)
	}
	return kubernetesError.Unwrap()
}
