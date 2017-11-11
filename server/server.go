package server

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rancher/norman/api"
	"github.com/rancher/norman/store/crd"
	"github.com/rancher/norman/types"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewAPIServer(ctx context.Context, kubeConfig string, schemas *types.Schemas) (http.Handler, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build kubeConfig")
	}
	return NewAPIServerFromConfig(ctx, config, schemas)
}

func NewClients(kubeConfig string) (*rest.RESTClient, clientset.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, nil, err
	}
	return NewClientsFromConfig(config)
}

func NewClientsFromConfig(config *rest.Config) (*rest.RESTClient, clientset.Interface, error) {
	dynamicConfig := *config
	if dynamicConfig.NegotiatedSerializer == nil {
		configConfig := dynamic.ContentConfig()
		dynamicConfig.NegotiatedSerializer = configConfig.NegotiatedSerializer
	}

	k8sClient, err := rest.UnversionedRESTClientFor(&dynamicConfig)
	if err != nil {
		return nil, nil, err
	}

	apiExtClient, err := clientset.NewForConfig(&dynamicConfig)
	if err != nil {
		return nil, nil, err
	}

	return k8sClient, apiExtClient, nil
}

func NewAPIServerFromConfig(ctx context.Context, config *rest.Config, schemas *types.Schemas) (http.Handler, error) {
	k8sClient, apiExtClient, err := NewClientsFromConfig(config)
	if err != nil {
		return nil, err
	}
	return NewAPIServerFromClients(ctx, k8sClient, apiExtClient, schemas)
}

func NewAPIServerFromClients(ctx context.Context, k8sClient *rest.RESTClient, apiExtClient clientset.Interface, schemas *types.Schemas) (http.Handler, error) {
	store := crd.NewCRDStore(apiExtClient, k8sClient)
	if err := store.AddSchemas(ctx, schemas); err != nil {
		return nil, err
	}

	server := api.NewAPIServer()
	if err := server.AddSchemas(schemas); err != nil {
		return nil, err
	}

	err := server.Start(ctx)
	return server, err
}
