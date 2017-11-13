package generator

var k8sClientTemplate = `package {{.version.Version}}

import (
	"sync"

	"github.com/rancher/norman/clientbase"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	RESTClient() rest.Interface
	{{range .schemas}}
	{{.CodeNamePlural}}Getter{{end}}
}

type Client struct {
	sync.Mutex
	restClient         rest.Interface
	{{range .schemas}}
	{{.ID}}Controllers map[string]{{.CodeName}}Controller{{end}}
}

func NewForConfig(config rest.Config) (Interface, error) {
	if config.NegotiatedSerializer == nil {
		configConfig := dynamic.ContentConfig()
		config.NegotiatedSerializer = configConfig.NegotiatedSerializer
	}

	restClient, err := rest.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restClient:         restClient,
	{{range .schemas}}
		{{.ID}}Controllers: map[string]{{.CodeName}}Controller{},{{end}}
	}, nil
}

func (c *Client) RESTClient() rest.Interface {
	return c.restClient
}

{{range .schemas}}
type {{.CodeNamePlural}}Getter interface {
	{{.CodeNamePlural}}(namespace string) {{.CodeName}}Interface
}

func (c *Client) {{.CodeNamePlural}}(namespace string) {{.CodeName}}Interface {
	objectClient := clientbase.NewObjectClient(namespace, c.restClient, &{{.CodeName}}Resource, {{.CodeName}}GroupVersionKind, {{.ID}}Factory{})
	return &{{.ID}}Client{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
{{end}}
`
