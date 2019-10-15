package proxy

import (
	"fmt"

	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type ClientFactory struct {
	cfg     rest.Config
	idToGVR map[string]schema.GroupVersionResource
}

func NewClientFactory(cfg *rest.Config) *ClientFactory {
	return &ClientFactory{
		cfg:     *cfg,
		idToGVR: map[string]schema.GroupVersionResource{},
	}
}

func (p *ClientFactory) Register(schema *types.Schema, gvr schema.GroupVersionResource) {
	p.idToGVR[schema.ID] = gvr

	schema.Store = NewProxyStore(p)
	schema.Mapper = AddAPIVersionKind{
		APIVersion: fmt.Sprintf("%s/%s", gvr.Group, gvr.Version),
		Kind:       schema.CodeName,
		Next:       schema.Mapper,
	}
}

func (p *ClientFactory) Client(ctx *types.APIRequest, schema *types.Schema) (dynamic.ResourceInterface, error) {
	gvr, ok := p.idToGVR[schema.ID]
	if !ok {
		return nil, httperror.NewAPIError(httperror.NotFound, "Failed to find client for "+schema.ID)
	}

	user, ok := request.UserFrom(ctx.Request.Context())
	if !ok {
		return nil, fmt.Errorf("failed to find user context for client")
	}
	newCfg := p.cfg
	newCfg.Impersonate.UserName = user.GetName()
	newCfg.Impersonate.Groups = user.GetGroups()
	newCfg.Impersonate.Extra = user.GetExtra()

	client, err := dynamic.NewForConfig(&newCfg)
	if err != nil {
		return nil, err
	}

	if len(ctx.Namespaces) > 0 {
		return client.Resource(gvr).Namespace(ctx.Namespaces[0]), nil
	}

	return client.Resource(gvr), nil
}
