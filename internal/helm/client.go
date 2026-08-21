package helm

import (
	"context"
	stderrors "errors"
	"sync"
	"time"

	"github.com/mitchellh/copystructure"
	yarerrors "github.com/yar-run/yar/internal/errors"
	"github.com/yar-run/yar/internal/kubernetes"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	defaultTimeout   = 30 * time.Second
	labelProject     = "yar.io/project"
	labelEnvironment = "yar.io/environment"
)

var errReleaseNotOwned = stderrors.New("release is not managed by this Yar project and environment")

// Client provides chart and release operations.
type Client interface {
	LoadChart(path string) (*chart.Chart, error)
	LoadChartFiles(files []ChartFile) (*chart.Chart, error)
	Render(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts RenderOptions) (*RenderResult, error)
	Install(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts InstallOptions) (*release.Release, error)
	Upgrade(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts UpgradeOptions) (*release.Release, error)
	Uninstall(ctx context.Context, release Release, opts UninstallOptions) error
}

// ChartFile is an in-memory chart file.
type ChartFile struct {
	Name string
	Data []byte
}

// Release identifies a Yar-managed Helm release.
type Release struct {
	Name        string
	Namespace   string
	Project     string
	Environment string
}

// RenderOptions configures deterministic local rendering.
type RenderOptions struct {
	KubeVersion string
	APIVersions []string
	IncludeCRDs bool
}

// RenderResult contains local Helm render output.
type RenderResult struct {
	Manifest string
	Notes    string
}

// InstallOptions configures a Helm install operation.
type InstallOptions struct {
	DryRun bool
}

// UpgradeOptions configures a Helm upgrade operation.
type UpgradeOptions struct {
	DryRun bool
}

// UninstallOptions configures a Helm uninstall operation.
type UninstallOptions struct {
	DryRun bool
}

type clientOptions struct {
	timeout    time.Duration
	kubernetes kubernetes.Client
}

// Option configures a Helm client.
type Option func(*clientOptions)

// WithTimeout sets the Helm lifecycle action timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) {
		options.timeout = timeout
	}
}

// WithKubernetesClient supplies Yar's already-selected in-memory Kubernetes client.
func WithKubernetesClient(client kubernetes.Client) Option {
	return func(options *clientOptions) {
		options.kubernetes = client
	}
}

type client struct {
	timeout   time.Duration
	namespace string
	config    *action.Configuration
	mu        sync.Mutex
}

type redactedCause struct {
	message string
}

func (e *redactedCause) Error() string {
	return e.message
}

// NewClient creates an isolated Helm client.
func NewClient(opts ...Option) (Client, error) {
	options := clientOptions{timeout: defaultTimeout}
	for _, opt := range opts {
		opt(&options)
	}

	if options.timeout < 0 {
		return nil, validationError("timeout", options.timeout, "must not be negative")
	}
	client := &client{timeout: normalizedTimeout(options.timeout)}
	if options.kubernetes != nil {
		configuration, err := newKubernetesConfiguration(options.kubernetes)
		if err != nil {
			return nil, err
		}
		client.config = configuration
		client.namespace = options.kubernetes.Namespace()
	}
	return client, nil
}

func normalizedTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	return timeout
}

func (r Release) validate() error {
	if r.Name == "" {
		return validationError("release.name", r.Name, "is required")
	}
	if r.Namespace == "" {
		return validationError("release.namespace", r.Namespace, "is required")
	}
	if r.Project == "" {
		return validationError("release.project", r.Project, "is required")
	}
	if r.Environment == "" {
		return validationError("release.environment", r.Environment, "is required")
	}
	if err := chartutil.ValidateReleaseName(r.Name); err != nil {
		return validationError("release.name", r.Name, "must be a valid Helm release name")
	}
	if validationErrors := validation.IsDNS1123Label(r.Namespace); len(validationErrors) > 0 {
		return &yarerrors.ValidationError{Field: "release.namespace", Value: r.Namespace, Message: "must be a valid Kubernetes namespace", Errors: validationErrors}
	}
	for field, value := range map[string]string{
		"release.namespace":   r.Namespace,
		"release.project":     r.Project,
		"release.environment": r.Environment,
	} {
		if validationErrors := validation.IsValidLabelValue(value); len(validationErrors) > 0 {
			return &yarerrors.ValidationError{Field: field, Value: value, Message: "must be a valid Kubernetes label value", Errors: validationErrors}
		}
	}
	return nil
}

func releaseLabels(release Release) map[string]string {
	return map[string]string{
		labelProject:     release.Project,
		labelEnvironment: release.Environment,
	}
}

func (r Release) matches(labels map[string]string) bool {
	return labels[labelProject] == r.Project && labels[labelEnvironment] == r.Environment
}

func (c *client) release(release Release) (*release.Release, error) {
	if err := release.validate(); err != nil {
		return nil, err
	}
	current, err := c.config.Releases.Last(release.Name)
	if err != nil {
		return nil, err
	}
	if !release.matches(current.Labels) {
		return nil, ownershipError("helm.release", release)
	}
	if current.Namespace != release.Namespace {
		return nil, ownershipError("helm.release", release)
	}
	return current, nil
}

func ownershipError(op string, release Release) *yarerrors.KubernetesError {
	return &yarerrors.KubernetesError{
		Op:        op,
		Resource:  "helm-release",
		Name:      release.Name,
		Namespace: release.Namespace,
		Err:       errReleaseNotOwned,
	}
}

func lifecycleError(op string, release Release, err error) error {
	if stderrors.Is(err, context.Canceled) || stderrors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return &yarerrors.KubernetesError{Op: op, Resource: "helm-release", Name: release.Name, Namespace: release.Namespace, Err: redactError("Helm operation failed", err)}
}

func validationError(field string, value any, message string) *yarerrors.ValidationError {
	return &yarerrors.ValidationError{Field: field, Value: value, Message: message}
}

func chartError(chart *chart.Chart, message string, err error) *yarerrors.PackError {
	name := "chart"
	if chart != nil && chart.Metadata != nil && chart.Name() != "" {
		name = chart.Name()
	}
	return &yarerrors.PackError{Pack: name, Message: message, Err: redactError("chart operation failed", err)}
}

func (c *client) ensureChart(chart *chart.Chart) error {
	if chart == nil || chart.Metadata == nil {
		return validationError("chart", "", "is required")
	}
	return nil
}

func newKubernetesConfiguration(client kubernetes.Client) (*action.Configuration, error) {
	getter, err := kubernetes.HelmRESTClientGetter(client)
	if err != nil {
		return nil, &yarerrors.KubernetesError{Op: "helm.initialize", Resource: "api-server", Name: "selected-context", Err: err}
	}
	configuration := &action.Configuration{}
	if err := configuration.Init(getter, client.Namespace(), "secret", func(string, ...interface{}) {}); err != nil {
		return nil, &yarerrors.KubernetesError{Op: "helm.initialize", Resource: "api-server", Name: "selected-context", Err: err}
	}
	return configuration, nil
}

func (c *client) ensureLifecycleConfiguration() error {
	if c.config == nil {
		return validationError("kubernetes", "", "a Yar Kubernetes client is required for Helm lifecycle operations")
	}
	return nil
}

func (c *client) validateLifecycleRelease(release Release) error {
	if err := release.validate(); err != nil {
		return err
	}
	if err := c.ensureLifecycleConfiguration(); err != nil {
		return err
	}
	if c.namespace != "" && release.Namespace != c.namespace {
		return validationError("release.namespace", release.Namespace, "must match the selected Kubernetes namespace")
	}
	return nil
}

func cloneValues(values map[string]any) (map[string]any, error) {
	if values == nil {
		return map[string]any{}, nil
	}
	cloned, err := copystructure.Copy(values)
	if err != nil {
		return nil, validationError("values", "", "could not be copied")
	}
	result, ok := cloned.(map[string]any)
	if !ok {
		return nil, validationError("values", "", "must be a map")
	}
	return result, nil
}

func validateChartValues(chart *chart.Chart, values map[string]any) error {
	if len(chart.Metadata.Dependencies) != 0 || len(chart.Dependencies()) != 0 {
		return &yarerrors.PackError{Pack: chart.Name(), Message: "chart dependencies are not supported"}
	}
	return validateValueShape(chart.Values, values, "values")
}

func validateValueShape(defaults, values map[string]any, path string) error {
	for key, value := range values {
		defaultValue, exists := defaults[key]
		if !exists || value == nil || defaultValue == nil {
			continue
		}
		defaultMap, defaultIsMap := defaultValue.(map[string]any)
		valueMap, valueIsMap := value.(map[string]any)
		if defaultIsMap != valueIsMap {
			return validationError(path+"."+key, "", "must match the chart value structure")
		}
		if defaultIsMap {
			if err := validateValueShape(defaultMap, valueMap, path+"."+key); err != nil {
				return err
			}
		}
	}
	return nil
}

func redactError(message string, err error) error {
	if err == nil {
		return nil
	}
	return &redactedCause{message: message}
}
