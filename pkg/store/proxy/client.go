package proxy

import (
	"fmt"
	"net/http"

	"github.com/rancher/norman/pkg/types"

	"github.com/rancher/norman/pkg/httperror"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type ProxyClient struct {
	cfg       rest.Config
	transport http.RoundTripper
	idToGVR   map[string]schema.GroupVersionResource
}

func NewProxyClient(cfg *rest.Config) *ProxyClient {
	return &ProxyClient{
		cfg:       *cfg,
		transport: http.DefaultTransport,
	}
}

func (p *ProxyClient) Register(schema *types.Schema, gvr schema.GroupVersionResource) {
	p.idToGVR[schema.ID] = gvr
}

func (p *ProxyClient) Client(ctx *types.APIOperation, schema *types.Schema) (dynamic.ResourceInterface, error) {
	gvr, ok := p.idToGVR[schema.ID]
	if !ok {
		return nil, httperror.NewAPIError(httperror.NotFound, "Failed to find client for "+schema.ID)
	}

	user, ok := request.UserFrom(ctx.Request.Context())
	if !ok {
		return nil, fmt.Errorf("failed to find user context for client")
	}
	newCfg := p.cfg
	newCfg.Transport = p.transport
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
