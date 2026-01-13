package docker

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	dockerclient "github.com/docker/docker/client"
)

// Client provides Docker operations.
type Client interface {
	// Network operations
	NetworkCreate(ctx context.Context, name string, opts NetworkCreateOptions) (string, error)
	NetworkRemove(ctx context.Context, name string) error
	NetworkList(ctx context.Context, opts NetworkListOptions) ([]Network, error)
	NetworkInspect(ctx context.Context, name string) (*Network, error)

	// Ping checks Docker daemon connectivity
	Ping(ctx context.Context) error

	// Close releases resources
	Close() error
}

// clientOptions holds configuration for the Docker client.
type clientOptions struct {
	host       string
	timeout    time.Duration
	apiVersion string
	tlsConfig  *tls.Config
	httpClient *http.Client
}

// Option configures the Docker client.
type Option func(*clientOptions)

// WithHost sets the Docker host (e.g., "unix:///var/run/docker.sock").
func WithHost(host string) Option {
	return func(o *clientOptions) {
		o.host = host
	}
}

// WithTimeout sets the operation timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithAPIVersion sets the Docker API version.
func WithAPIVersion(version string) Option {
	return func(o *clientOptions) {
		o.apiVersion = version
	}
}

// WithTLSConfig sets TLS configuration for remote Docker hosts.
func WithTLSConfig(config *tls.Config) Option {
	return func(o *clientOptions) {
		o.tlsConfig = config
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = httpClient
	}
}

// dockerClient wraps the Docker SDK client.
type dockerClient struct {
	cli     *dockerclient.Client
	timeout time.Duration
}

// NewClient creates a new Docker client with the given options.
func NewClient(opts ...Option) (Client, error) {
	// Apply defaults
	options := &clientOptions{
		timeout: 30 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(options)
	}

	// Build Docker client options
	var clientOpts []dockerclient.Opt

	if options.host != "" {
		clientOpts = append(clientOpts, dockerclient.WithHost(options.host))
	}

	if options.apiVersion != "" {
		clientOpts = append(clientOpts, dockerclient.WithVersion(options.apiVersion))
	} else {
		// Use API version negotiation by default
		clientOpts = append(clientOpts, dockerclient.WithAPIVersionNegotiation())
	}

	if options.httpClient != nil {
		clientOpts = append(clientOpts, dockerclient.WithHTTPClient(options.httpClient))
	} else if options.tlsConfig != nil {
		transport := &http.Transport{
			TLSClientConfig: options.tlsConfig,
		}
		httpClient := &http.Client{
			Transport: transport,
			Timeout:   options.timeout,
		}
		clientOpts = append(clientOpts, dockerclient.WithHTTPClient(httpClient))
	}

	// Create Docker client
	cli, err := dockerclient.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, ErrDaemonConnection(err)
	}

	return &dockerClient{
		cli:     cli,
		timeout: options.timeout,
	}, nil
}

// Ping checks Docker daemon connectivity.
func (c *dockerClient) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	if err != nil {
		return ErrDaemonConnection(err)
	}
	return nil
}

// Close releases resources.
func (c *dockerClient) Close() error {
	return c.cli.Close()
}

// getDockerClient returns the underlying Docker client (for network.go).
func (c *dockerClient) getDockerClient() *dockerclient.Client {
	return c.cli
}
