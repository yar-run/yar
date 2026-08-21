package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// HelmRESTClientGetter returns an in-memory REST getter for Helm actions.
// The returned getter preserves Yar's selected context and never reads or writes kubeconfig.
func HelmRESTClientGetter(kubernetesAPI Client) (genericclioptions.RESTClientGetter, error) {
	kubernetesClient, ok := kubernetesAPI.(*client)
	if !ok || kubernetesClient == nil {
		return nil, fmt.Errorf("unsupported Yar Kubernetes client")
	}
	if kubernetesClient.restConfig == nil || kubernetesClient.discovery == nil {
		return nil, fmt.Errorf("Kubernetes client is not initialized")
	}

	discoveryClient := memory.NewMemCacheClient(kubernetesClient.discovery)
	return &helmRESTClientGetter{
		restConfig: rest.CopyConfig(kubernetesClient.restConfig),
		discovery:  discoveryClient,
		mapper:     restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient),
		clientConfig: &helmClientConfig{
			restConfig: rest.CopyConfig(kubernetesClient.restConfig),
			namespace:  kubernetesClient.namespace,
		},
	}, nil
}

type helmRESTClientGetter struct {
	restConfig   *rest.Config
	discovery    discovery.CachedDiscoveryInterface
	mapper       meta.RESTMapper
	clientConfig clientcmd.ClientConfig
}

func (g *helmRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return rest.CopyConfig(g.restConfig), nil
}

func (g *helmRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return g.discovery, nil
}

func (g *helmRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return g.mapper, nil
}

func (g *helmRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return g.clientConfig
}

type helmClientConfig struct {
	restConfig *rest.Config
	namespace  string
}

func (c *helmClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, nil
}

func (c *helmClientConfig) ClientConfig() (*rest.Config, error) {
	return rest.CopyConfig(c.restConfig), nil
}

func (c *helmClientConfig) Namespace() (string, bool, error) {
	return c.namespace, true, nil
}

func (c *helmClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return nil
}

var _ genericclioptions.RESTClientGetter = (*helmRESTClientGetter)(nil)
var _ clientcmd.ClientConfig = (*helmClientConfig)(nil)
