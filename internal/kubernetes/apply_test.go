package kubernetes

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	yarerrors "github.com/yar-run/yar/internal/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const testOwnerProject = "backend"
const testOwnerEnvironment = "local"

func TestClient_ImplementsResourceClient(t *testing.T) {
	t.Parallel()

	var _ Client = (*client)(nil)
}

func TestClient_ApplyCreatesMissingResourceWithOwnershipLabels(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writeStatus(t, responseWriter, http.StatusNotFound, "NotFound")
			return
		}
		expectRequest(t, request, http.MethodPost, "/apis/apps/v1/namespaces/manifest-ns/deployments")
		if got, want := request.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("Content-Type = %q, want %q", got, want)
		}
		if got, want := request.URL.Query().Get("fieldManager"), "yar"; got != want {
			t.Errorf("fieldManager = %q, want %q", got, want)
		}
		if got := request.URL.Query().Get("force"); got != "" {
			t.Errorf("force = %q, want empty create query", got)
		}

		var body unstructured.Unstructured
		decodeJSON(t, request.Body, &body)
		labels := body.GetLabels()
		wantLabels := map[string]string{
			"existing":       "label",
			labelManagedBy:   managedByYar,
			labelProject:     testOwnerProject,
			labelEnvironment: testOwnerEnvironment,
		}
		if diff := cmp.Diff(wantLabels, labels); diff != "" {
			t.Errorf("applied labels mismatch (-want +got):\n%s", diff)
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write(mustJSON(t, body.Object))
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))

	manifest := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: manifest-ns
  labels:
    existing: label
spec:
  replicas: 1
`)
	if err := client.Apply(context.Background(), manifest, ApplyOptions{}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
}

func TestClient_ApplyUpdatesOwnedResourceWithServerSideApply(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		expectRequest(t, request, request.Method, "/api/v1/namespaces/default/services/api")
		switch request.Method {
		case http.MethodGet:
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write(mustJSON(t, map[string]any{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]any{
					"name":            "api",
					"namespace":       "default",
					"labels":          mergeLabels(ownerLabels(testOwnerProject, testOwnerEnvironment), map[string]string{"existing": "label"}),
					"uid":             "owned-uid",
					"resourceVersion": "42",
				},
			}))
		case http.MethodPatch:
			if got, want := request.Header.Get("Content-Type"), "application/apply-patch+yaml"; got != want {
				t.Errorf("Content-Type = %q, want %q", got, want)
			}
			if got, want := request.URL.Query().Get("fieldManager"), "yar"; got != want {
				t.Errorf("fieldManager = %q, want %q", got, want)
			}
			if got, want := request.URL.Query().Get("force"), "true"; got != want {
				t.Errorf("force = %q, want %q", got, want)
			}
			var body unstructured.Unstructured
			decodeJSON(t, request.Body, &body)
			if _, found := body.GetLabels()["existing"]; found {
				t.Error("PATCH claimed existing label that is absent from the desired manifest")
			}
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write(mustJSON(t, body.Object))
		}
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))

	if err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
}

func TestClient_ApplySupportsMultiDocumentAndScopes(t *testing.T) {
	t.Parallel()

	var requests []string
	var requestsMu sync.Mutex
	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writeStatus(t, responseWriter, http.StatusNotFound, "NotFound")
			requestsMu.Lock()
			requests = append(requests, request.Method+" "+request.URL.Path+"?"+request.URL.RawQuery)
			requestsMu.Unlock()
			return
		}
		requestsMu.Lock()
		requests = append(requests, request.Method+" "+request.URL.Path+"?"+request.URL.RawQuery)
		requestsMu.Unlock()
		var body unstructured.Unstructured
		decodeJSON(t, request.Body, &body)
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write(mustJSON(t, body.Object))
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment), WithNamespace("client-ns"))

	manifest := []byte(`
---
# empty document
---
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  selector:
    app: api
---
apiVersion: v1
kind: Namespace
metadata:
  name: backend
`)
	force := false
	if err := client.Apply(context.Background(), manifest, ApplyOptions{FieldManager: "yar-tests", Force: &force, DryRun: true}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	requestsMu.Lock()
	defer requestsMu.Unlock()
	want := []string{
		"GET /api/v1/namespaces/client-ns/services/api?",
		"POST /api/v1/namespaces/client-ns/services?dryRun=All&fieldManager=yar-tests",
		"GET /api/v1/namespaces/backend?",
		"POST /api/v1/namespaces?dryRun=All&fieldManager=yar-tests",
	}
	if diff := cmp.Diff(want, requests); diff != "" {
		t.Errorf("requests mismatch (-want +got):\n%s", diff)
	}
}

func TestClient_ApplyRefusesExternalResource(t *testing.T) {
	t.Parallel()

	var mutations int
	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		expectRequest(t, request, request.Method, "/api/v1/namespaces/default/services/api")
		switch request.Method {
		case http.MethodGet:
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{"apiVersion":"v1","kind":"Service","metadata":{"name":"api","namespace":"default","labels":{"app":"external"}}}`))
		case http.MethodPatch, http.MethodPost:
			mutations++
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))

	err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{})
	if err == nil {
		t.Fatal("Apply() error = nil, want non-nil")
	}
	var kubernetesError *yarerrors.KubernetesError
	if !stderrors.As(err, &kubernetesError) || kubernetesError.Op != "apply" {
		t.Errorf("Apply() error = %#v, want apply KubernetesError", err)
	}
	if mutations != 0 {
		t.Errorf("mutation requests = %d, want 0", mutations)
	}
}

func TestClient_ApplyPinsExistingOwnedResourceIdentity(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		expectRequest(t, request, request.Method, "/api/v1/namespaces/default/services/api")
		switch request.Method {
		case http.MethodGet:
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write(mustJSON(t, map[string]any{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]any{
					"name":            "api",
					"namespace":       "default",
					"labels":          ownerLabels(testOwnerProject, testOwnerEnvironment),
					"uid":             "owned-uid",
					"resourceVersion": "42",
				},
			}))
		case http.MethodPatch:
			var body unstructured.Unstructured
			decodeJSON(t, request.Body, &body)
			if got, want := string(body.GetUID()), "owned-uid"; got != want {
				t.Errorf("PATCH UID = %q, want %q", got, want)
			}
			if got := body.GetResourceVersion(); got != "" {
				t.Errorf("PATCH resourceVersion = %q, want empty SSA payload version", got)
			}
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write(mustJSON(t, body.Object))
		}
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))

	if err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
}

func TestClient_ApplyRefusesConcurrentExternalCreator(t *testing.T) {
	t.Parallel()

	var mutations int
	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		expectRequest(t, request, request.Method, "/api/v1/namespaces/default/services/api")
		switch request.Method {
		case http.MethodGet:
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{"apiVersion":"v1","kind":"Service","metadata":{"name":"api","namespace":"default","labels":{"app":"external"}}}`))
		case http.MethodPost:
			mutations++
			writeStatus(t, responseWriter, http.StatusConflict, "AlreadyExists")
		case http.MethodPatch:
			mutations++
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))

	err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{})
	if err == nil {
		t.Fatal("Apply() error = nil, want non-nil")
	}
	if mutations != 0 {
		t.Errorf("mutation requests = %d, want 0", mutations)
	}
}

func TestClient_ApplyRejectsInvalidManifest(t *testing.T) {
	t.Parallel()

	tests := map[string][]byte{
		"empty manifest":        nil,
		"malformed YAML":        []byte("apiVersion: v1\nmetadata: ["),
		"missing api version":   []byte("kind: Service\nmetadata:\n  name: api"),
		"missing kind":          []byte("apiVersion: v1\nmetadata:\n  name: api"),
		"missing name":          []byte("apiVersion: v1\nkind: Service"),
		"managed fields":        []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n  managedFields: []"),
		"no owner for mutation": []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api"),
	}

	for name, manifest := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var applyClient *client
			if name == "no owner for mutation" {
				applyClient = newTestClient(t)
			} else {
				applyClient = newTestClient(t, WithOwner(testOwnerProject, testOwnerEnvironment))
			}
			err := applyClient.Apply(context.Background(), manifest, ApplyOptions{})
			if err == nil {
				t.Fatal("Apply() error = nil, want non-nil")
			}
			var validationError *yarerrors.ValidationError
			if !stderrors.As(err, &validationError) {
				t.Errorf("Apply() error type = %T, want *errors.ValidationError", err)
			}
		})
	}
}

func TestClient_GetAndList(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/namespaces/client-ns/services/api":
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{"apiVersion":"v1","kind":"Service","metadata":{"name":"api","namespace":"client-ns"}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/services":
			if got, want := request.URL.Query().Get("labelSelector"), "app=api"; got != want {
				t.Errorf("labelSelector = %q, want %q", got, want)
			}
			if got, want := request.URL.Query().Get("limit"), "2"; got != want {
				t.Errorf("limit = %q, want %q", got, want)
			}
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{"apiVersion":"v1","kind":"ServiceList","items":[{"apiVersion":"v1","kind":"Service","metadata":{"name":"api","namespace":"client-ns"}}]}`))
		default:
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.String())
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	})
	client := newResourceClient(t, server, WithNamespace("client-ns"))

	gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
	service, err := client.Get(context.Background(), gvk, "", "api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got, want := service.GetNamespace(), "client-ns"; got != want {
		t.Errorf("Get().Namespace = %q, want %q", got, want)
	}
	services, err := client.List(context.Background(), gvk, "", ListOptions{LabelSelector: "app=api", Limit: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(services.Items) != 1 || services.Items[0].GetName() != "api" {
		t.Errorf("List().Items = %#v, want one api service", services.Items)
	}
}

func TestClient_DeleteOnlyDeletesExactOwnerMatches(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		liveStatus  int
		liveLabels  map[string]string
		liveUID     string
		liveVersion string
		wantErr     bool
		wantDeletes int
	}{
		"owned resource": {
			liveStatus:  http.StatusOK,
			liveLabels:  ownerLabels(testOwnerProject, testOwnerEnvironment),
			liveUID:     "owned-uid",
			liveVersion: "42",
			wantDeletes: 1,
		},
		"missing resource is idempotent": {
			liveStatus:  http.StatusNotFound,
			wantDeletes: 0,
		},
		"external resource": {
			liveStatus: http.StatusOK,
			liveLabels: map[string]string{"app": "external"},
			wantErr:    true,
		},
		"partial labels": {
			liveStatus: http.StatusOK,
			liveLabels: map[string]string{labelManagedBy: managedByYar},
			wantErr:    true,
		},
		"other project": {
			liveStatus: http.StatusOK,
			liveLabels: ownerLabels("other-project", testOwnerEnvironment),
			wantErr:    true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var deletes int
			server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
				expectRequest(t, request, request.Method, "/api/v1/namespaces/client-ns/services/api")
				switch request.Method {
				case http.MethodGet:
					if tc.liveStatus == http.StatusNotFound {
						writeStatus(t, responseWriter, http.StatusNotFound, "NotFound")
						return
					}
					responseWriter.Header().Set("Content-Type", "application/json")
					_, _ = responseWriter.Write(mustJSON(t, map[string]any{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]any{
							"name":            "api",
							"namespace":       "client-ns",
							"labels":          tc.liveLabels,
							"uid":             tc.liveUID,
							"resourceVersion": tc.liveVersion,
						},
					}))
				case http.MethodDelete:
					deletes++
					var deleteOptions metav1.DeleteOptions
					decodeJSON(t, request.Body, &deleteOptions)
					if tc.liveUID != "" && (deleteOptions.Preconditions == nil || deleteOptions.Preconditions.UID == nil || string(*deleteOptions.Preconditions.UID) != tc.liveUID) {
						t.Errorf("DELETE UID precondition = %#v, want %q", deleteOptions.Preconditions, tc.liveUID)
					}
					if tc.liveVersion != "" && (deleteOptions.Preconditions == nil || deleteOptions.Preconditions.ResourceVersion == nil || *deleteOptions.Preconditions.ResourceVersion != tc.liveVersion) {
						t.Errorf("DELETE resourceVersion precondition = %#v, want %q", deleteOptions.Preconditions, tc.liveVersion)
					}
					responseWriter.WriteHeader(http.StatusOK)
				default:
					responseWriter.WriteHeader(http.StatusMethodNotAllowed)
				}
			})
			client := newResourceClient(t, server, WithNamespace("client-ns"), WithOwner(testOwnerProject, testOwnerEnvironment))
			err := client.Delete(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), DeleteOptions{PropagationPolicy: "foreground"})
			if (err != nil) != tc.wantErr {
				t.Errorf("Delete() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				var kubernetesError *yarerrors.KubernetesError
				if !stderrors.As(err, &kubernetesError) || kubernetesError.Op != "delete" {
					t.Errorf("Delete() error = %#v, want delete KubernetesError", err)
				}
			}
			if deletes != tc.wantDeletes {
				t.Errorf("DELETE requests = %d, want %d", deletes, tc.wantDeletes)
			}
		})
	}
}

func TestClient_ResourceOperationsPreserveContextAndAPIStatusErrors(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/v1/namespaces/client-ns/services/api" {
			writeStatus(t, responseWriter, http.StatusForbidden, "Forbidden")
			return
		}
		responseWriter.WriteHeader(http.StatusNotFound)
	})
	client := newResourceClient(t, server, WithNamespace("client-ns"))
	_, err := client.Get(context.Background(), schema.GroupVersionKind{Version: "v1", Kind: "Service"}, "", "api")
	if err == nil {
		t.Fatal("Get() error = nil, want non-nil")
	}
	var kubernetesError *yarerrors.KubernetesError
	if !stderrors.As(err, &kubernetesError) || !apierrors.IsForbidden(kubernetesError.Unwrap()) {
		t.Errorf("Get() error = %#v, want forbidden KubernetesError", err)
	}
}

func TestClient_ApplyPreservesAPIStatusErrors(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		writeStatus(t, responseWriter, http.StatusConflict, "Conflict")
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))
	err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{})
	if err == nil {
		t.Fatal("Apply() error = nil, want non-nil")
	}
	var kubernetesError *yarerrors.KubernetesError
	if !stderrors.As(err, &kubernetesError) || !apierrors.IsConflict(kubernetesError.Unwrap()) {
		t.Errorf("Apply() error = %#v, want conflict KubernetesError", err)
	}
}

func TestClient_ApplyMapsMalformedResponseAndTransportFailure(t *testing.T) {
	t.Parallel()

	t.Run("malformed create response", func(t *testing.T) {
		t.Parallel()
		server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				writeStatus(t, responseWriter, http.StatusNotFound, "NotFound")
			case http.MethodPost:
				responseWriter.Header().Set("Content-Type", "application/json")
				_, _ = responseWriter.Write([]byte("{not-json}"))
			}
		})
		client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))
		err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{})
		if err == nil {
			t.Fatal("Apply() error = nil, want non-nil")
		}
		var kubernetesError *yarerrors.KubernetesError
		if !stderrors.As(err, &kubernetesError) || kubernetesError.Unwrap() == nil {
			t.Errorf("Apply() error = %#v, want KubernetesError with cause", err)
		}
	})

	t.Run("transport failure", func(t *testing.T) {
		t.Parallel()
		server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
			writeStatus(t, responseWriter, http.StatusNotFound, "NotFound")
		})
		client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment))
		server.Close()
		err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{})
		if err == nil {
			t.Fatal("Apply() error = nil, want non-nil")
		}
		var kubernetesError *yarerrors.KubernetesError
		if !stderrors.As(err, &kubernetesError) || kubernetesError.Unwrap() == nil {
			t.Errorf("Apply() error = %#v, want KubernetesError with cause", err)
		}
	})
}

func TestClient_ApplyHonorsTimeout(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		close(started)
		<-request.Context().Done()
	})
	client := newResourceClient(t, server, WithOwner(testOwnerProject, testOwnerEnvironment), WithTimeout(25*time.Millisecond))
	err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{})
	if !stderrors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Apply() error = %v, want context.DeadlineExceeded", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("Apply() did not reach the server")
	}
}

func TestClient_ListUsesRequestedNamespace(t *testing.T) {
	t.Parallel()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		expectRequest(t, request, http.MethodGet, "/api/v1/namespaces/requested-ns/services")
		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`{"apiVersion":"v1","kind":"ServiceList","items":[]}`))
	})
	client := newResourceClient(t, server, WithNamespace("client-ns"))
	_, err := client.List(context.Background(), schema.GroupVersionKind{Version: "v1", Kind: "Service"}, "requested-ns", ListOptions{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
}

func TestClient_ResourceOperationsPropagateCancellation(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		close(started)
		<-request.Context().Done()
	})
	client := newResourceClient(t, server, WithNamespace("client-ns"))

	ctx, cancel := context.WithCancel(context.Background())
	errC := make(chan error, 1)
	go func() {
		_, err := client.Get(ctx, schema.GroupVersionKind{Version: "v1", Kind: "Service"}, "", "api")
		errC <- err
	}()
	select {
	case <-started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("Get() did not reach the server")
	}
	select {
	case err := <-errC:
		if !stderrors.Is(err, context.Canceled) {
			t.Errorf("Get() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Get() did not return after cancellation")
	}
}

func newResourceClient(t *testing.T, server *httptest.Server, opts ...Option) *client {
	t.Helper()

	kubeconfig := writeKubeconfig(t, map[string]kubeconfigCluster{
		"default": {server: server, token: "token", namespace: "default"},
	}, "default")
	options := append([]Option{WithKubeconfig(kubeconfig)}, opts...)
	configured, err := NewClient(options...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client, ok := configured.(*client)
	if !ok {
		t.Fatalf("NewClient() type = %T, want *client", configured)
	}
	client.mapper = testRESTMapper()
	return client
}

func newTestClient(t *testing.T, opts ...Option) *client {
	t.Helper()

	server := newResourceServer(t, func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusNotFound)
	})
	return newResourceClient(t, server, opts...)
}

func testRESTMapper() meta.RESTMapper {
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{
		{Version: "v1"},
		{Group: "apps", Version: "v1"},
	})
	mapper.AddSpecific(
		schema.GroupVersionKind{Version: "v1", Kind: "Service"},
		schema.GroupVersionResource{Version: "v1", Resource: "services"},
		schema.GroupVersionResource{Version: "v1", Resource: "service"},
		meta.RESTScopeNamespace,
	)
	mapper.AddSpecific(
		schema.GroupVersionKind{Version: "v1", Kind: "Namespace"},
		schema.GroupVersionResource{Version: "v1", Resource: "namespaces"},
		schema.GroupVersionResource{Version: "v1", Resource: "namespace"},
		meta.RESTScopeRoot,
	)
	mapper.AddSpecific(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployment"},
		meta.RESTScopeNamespace,
	)
	return mapper
}

func newResourceServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	return server
}

func expectRequest(t *testing.T, request *http.Request, method, path string) {
	t.Helper()
	if request.Method != method || request.URL.Path != path {
		t.Errorf("request = %s %s, want %s %s", request.Method, request.URL.Path, method, path)
	}
}

func decodeJSON(t *testing.T, reader io.Reader, value any) {
	t.Helper()
	if err := json.NewDecoder(reader).Decode(value); err != nil {
		t.Fatalf("decode JSON error = %v", err)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return encoded
}

func writeStatus(t *testing.T, responseWriter http.ResponseWriter, statusCode int, reason string) {
	t.Helper()
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	_, _ = responseWriter.Write(mustJSON(t, metav1.Status{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Status"},
		Status:   metav1.StatusFailure,
		Reason:   metav1.StatusReason(reason),
		Code:     int32(statusCode),
	}))
}

func ownerLabels(project, environment string) map[string]string {
	return map[string]string{
		labelManagedBy:   managedByYar,
		labelProject:     project,
		labelEnvironment: environment,
	}
}

func mergeLabels(first, second map[string]string) map[string]string {
	merged := make(map[string]string, len(first)+len(second))
	for key, value := range first {
		merged[key] = value
	}
	for key, value := range second {
		merged[key] = value
	}
	return merged
}

func TestValidateListOptions(t *testing.T) {
	t.Parallel()

	if err := validateListOptions(ListOptions{Limit: -1}); err == nil {
		t.Error("validateListOptions() error = nil, want non-nil")
	}
	if err := validateListOptions(ListOptions{LabelSelector: "invalid===selector"}); err == nil {
		t.Error("validateListOptions() error = nil, want non-nil")
	}
}

func TestDeleteOptions(t *testing.T) {
	t.Parallel()

	if _, err := deletePropagationPolicy(DeleteOptions{PropagationPolicy: "invalid"}); err == nil {
		t.Error("deletePropagationPolicy() error = nil, want non-nil")
	}
}

func TestClient_ApplyRejectsInvalidFieldManager(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, WithOwner(testOwnerProject, testOwnerEnvironment))
	err := client.Apply(context.Background(), []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: api\n"), ApplyOptions{FieldManager: "invalid\nmanager"})
	if err == nil {
		t.Fatal("Apply() error = nil, want non-nil")
	}
	var validationError *yarerrors.ValidationError
	if !stderrors.As(err, &validationError) || validationError.Field != "fieldManager" {
		t.Errorf("Apply() error = %#v, want fieldManager ValidationError", err)
	}
}

func TestOwnerIdentityRequiresBothValues(t *testing.T) {
	t.Parallel()

	tests := []ownerIdentity{{}, {project: "backend"}, {environment: "local"}}
	for _, owner := range tests {
		if err := owner.validate(); err == nil {
			t.Errorf("owner.validate() error = nil for %#v, want non-nil", owner)
		}
	}
}
