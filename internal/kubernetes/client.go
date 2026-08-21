package kubernetes

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	yarerrors "github.com/yar-run/yar/internal/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	defaultNamespace    = "default"
	defaultTimeout      = 30 * time.Second
	defaultFieldManager = "yar"
	labelManagedBy      = "app.kubernetes.io/managed-by"
	managedByYar        = "yar"
	labelProject        = "yar.io/project"
	labelEnvironment    = "yar.io/environment"
)

// Client provides Kubernetes connection and resource operations.
type Client interface {
	// Context returns the kubeconfig context selected for this client.
	Context() string

	// Namespace returns the namespace selected for this client.
	Namespace() string

	// Probe verifies API-server connectivity and returns server version information.
	Probe(ctx context.Context) (*version.Info, error)

	// Apply applies a multi-document manifest stream through server-side apply.
	Apply(ctx context.Context, manifests []byte, opts ApplyOptions) error

	// Delete removes Yar-owned resources declared by a multi-document manifest stream.
	Delete(ctx context.Context, manifests []byte, opts DeleteOptions) error

	// Get retrieves one resource by GVK, namespace, and name.
	Get(ctx context.Context, gvk schema.GroupVersionKind, namespace, name string) (*unstructured.Unstructured, error)

	// List retrieves resources matching the supplied GVK, namespace, and options.
	List(ctx context.Context, gvk schema.GroupVersionKind, namespace string, opts ListOptions) (*unstructured.UnstructuredList, error)
}

type clientOptions struct {
	kubeconfig   string
	context      string
	namespace    string
	timeout      time.Duration
	owner        ownerIdentity
	fieldManager string
}

// Option configures a Kubernetes client.
type Option func(*clientOptions)

// WithKubeconfig selects an explicit kubeconfig path.
func WithKubeconfig(path string) Option {
	return func(options *clientOptions) {
		options.kubeconfig = path
	}
}

// WithContext selects a kubeconfig context in memory.
func WithContext(name string) Option {
	return func(options *clientOptions) {
		options.context = name
	}
}

// WithNamespace overrides the namespace selected from kubeconfig.
func WithNamespace(namespace string) Option {
	return func(options *clientOptions) {
		options.namespace = namespace
	}
}

// WithTimeout sets the default timeout for connection operations.
func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) {
		options.timeout = timeout
	}
}

// WithOwner identifies resources this client may manage.
func WithOwner(project, environment string) Option {
	return func(options *clientOptions) {
		options.owner = ownerIdentity{project: project, environment: environment}
	}
}

// WithFieldManager overrides the server-side apply field manager.
func WithFieldManager(name string) Option {
	return func(options *clientOptions) {
		options.fieldManager = name
	}
}

type client struct {
	context      string
	namespace    string
	timeout      time.Duration
	discovery    discovery.DiscoveryInterface
	restConfig   *rest.Config
	httpClient   *http.Client
	owner        ownerIdentity
	fieldManager string

	mapper      meta.RESTMapper
	dynamicOnce sync.Once
	dynamic     dynamic.Interface
	dynamicErr  error
}

// NewClient loads a trusted kubeconfig and creates a read-only Kubernetes client.
func NewClient(opts ...Option) (Client, error) {
	options := clientOptions{timeout: defaultTimeout, fieldManager: defaultFieldManager}
	for _, opt := range opts {
		opt(&options)
	}

	if options.namespace != "" {
		if validationErrors := validation.IsDNS1123Label(options.namespace); len(validationErrors) > 0 {
			return nil, &yarerrors.ValidationError{
				Field:   "namespace",
				Value:   options.namespace,
				Message: "must be a valid DNS-1123 label",
				Errors:  validationErrors,
			}
		}
	}
	if options.timeout < 0 {
		return nil, &yarerrors.ValidationError{
			Field:   "timeout",
			Value:   options.timeout,
			Message: "must not be negative",
		}
	}
	if options.fieldManager == "" {
		return nil, &yarerrors.ValidationError{Field: "fieldManager", Value: options.fieldManager, Message: "must not be empty"}
	}
	if options.owner.project != "" || options.owner.environment != "" {
		if err := options.owner.validate(); err != nil {
			return nil, err
		}
	}

	rawConfig, rules, err := loadKubeconfig(options.kubeconfig)
	if err != nil {
		return nil, err
	}

	contextName := options.context
	if contextName == "" {
		contextName = rawConfig.CurrentContext
	}
	if contextName == "" {
		return nil, &yarerrors.NotFoundError{
			Resource: "kubeconfig context",
			Name:     "current",
			Message:  "no current context is configured",
		}
	}
	if _, ok := rawConfig.Contexts[contextName]; !ok {
		return nil, &yarerrors.NotFoundError{
			Resource: "kubeconfig context",
			Name:     contextName,
			Message:  "context not found",
		}
	}

	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
		Context: clientcmdapi.Context{
			Namespace: options.namespace,
		},
	}
	clientConfig := clientcmd.NewNonInteractiveClientConfig(*rawConfig, contextName, overrides, nil)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, kubeconfigError(kubeconfigPath(rules), "failed to create Kubernetes client configuration", err)
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, kubeconfigError(kubeconfigPath(rules), "failed to determine Kubernetes namespace", err)
	}
	if namespace == "" {
		namespace = defaultNamespace
	}

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, kubeconfigError(kubeconfigPath(rules), "failed to create Kubernetes HTTP client", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, kubeconfigError(kubeconfigPath(rules), "failed to create Kubernetes discovery client", err)
	}
	return &client{
		context:      contextName,
		namespace:    namespace,
		timeout:      normalizedTimeout(options.timeout),
		discovery:    discoveryClient,
		restConfig:   restConfig,
		httpClient:   httpClient,
		owner:        ownerIdentity{project: options.owner.project, environment: options.owner.environment},
		fieldManager: options.fieldManager,
	}, nil
}

// Context returns the kubeconfig context selected for this client.
func (c *client) Context() string {
	return c.context
}

// Namespace returns the namespace selected for this client.
func (c *client) Namespace() string {
	return c.namespace
}

// Probe verifies API-server connectivity and returns server version information.
func (c *client) Probe(ctx context.Context) (*version.Info, error) {
	ctx, cancel := c.operationContext(ctx)
	defer cancel()

	body, err := c.discovery.RESTClient().Get().AbsPath("/version").Do(ctx).Raw()
	if err != nil {
		if isContextError(err) {
			return nil, err
		}
		return nil, probeError(c.context, err)
	}

	var info version.Info
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, probeError(c.context, fmt.Errorf("decode API server version: %w", err))
	}
	return &info, nil
}

func (c *client) operationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}

func normalizedTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	return timeout
}

func loadKubeconfig(explicitPath string) (*clientcmdapi.Config, *clientcmd.ClientConfigLoadingRules, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	// Loading must never migrate or rewrite a user's kubeconfig.
	rules.MigrationRules = nil
	if explicitPath != "" {
		rules.ExplicitPath = explicitPath
		if _, err := os.Stat(explicitPath); err != nil {
			if os.IsNotExist(err) {
				return nil, rules, &yarerrors.NotFoundError{
					Resource: "kubeconfig",
					Name:     explicitPath,
					Message:  "file not found",
				}
			}
			return nil, rules, kubeconfigError(explicitPath, "failed to read kubeconfig", err)
		}
	}

	rawConfig, err := rules.Load()
	if err != nil {
		return nil, rules, kubeconfigError(kubeconfigPath(rules), "failed to load kubeconfig", err)
	}
	if len(rawConfig.Contexts) == 0 {
		return nil, rules, &yarerrors.NotFoundError{
			Resource: "kubeconfig",
			Name:     kubeconfigPath(rules),
			Message:  "no contexts found",
		}
	}
	return rawConfig, rules, nil
}

func kubeconfigPath(rules *clientcmd.ClientConfigLoadingRules) string {
	if rules.ExplicitPath != "" {
		return rules.ExplicitPath
	}
	if len(rules.Precedence) == 1 {
		return rules.Precedence[0]
	}
	return "default kubeconfig"
}

func kubeconfigError(path, message string, err error) *yarerrors.ConfigError {
	return &yarerrors.ConfigError{
		Path:    path,
		Message: message,
		Err:     err,
	}
}

func probeError(contextName string, err error) *yarerrors.KubernetesError {
	return &yarerrors.KubernetesError{
		Op:       "probe",
		Resource: "api-server",
		Name:     contextName,
		Err:      err,
	}
}

func isContextError(err error) bool {
	return stderrors.Is(err, context.Canceled) || stderrors.Is(err, context.DeadlineExceeded)
}

type ownerIdentity struct {
	project     string
	environment string
}

func (o ownerIdentity) validate() error {
	if o.project == "" || o.environment == "" {
		return &yarerrors.ValidationError{Field: "owner", Value: "", Message: "project and environment are required for resource mutations"}
	}
	for field, value := range map[string]string{"owner.project": o.project, "owner.environment": o.environment} {
		if validationErrors := validation.IsValidLabelValue(value); len(validationErrors) > 0 {
			return &yarerrors.ValidationError{Field: field, Value: value, Message: "must be a valid Kubernetes label value", Errors: validationErrors}
		}
	}
	return nil
}

func (o ownerIdentity) labels() map[string]string {
	return map[string]string{
		labelManagedBy:   managedByYar,
		labelProject:     o.project,
		labelEnvironment: o.environment,
	}
}

func (c *client) resourceClient() (dynamic.Interface, meta.RESTMapper, error) {
	c.dynamicOnce.Do(func() {
		c.dynamic, c.dynamicErr = dynamic.NewForConfigAndClient(c.restConfig, c.httpClient)
		if c.dynamicErr != nil {
			return
		}
		if c.mapper == nil {
			c.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(c.discovery))
		}
	})
	return c.dynamic, c.mapper, c.dynamicErr
}
