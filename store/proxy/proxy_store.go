package proxy

import (
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

type Store struct {
	k8sClient      *rest.RESTClient
	prefix         []string
	group          string
	version        string
	kind           string
	resourcePlural string
}

func NewProxyStore(k8sClient *rest.RESTClient,
	prefix []string, group, version, kind, resourcePlural string) *Store {
	return &Store{
		k8sClient:      k8sClient,
		prefix:         prefix,
		group:          group,
		version:        version,
		kind:           kind,
		resourcePlural: resourcePlural,
	}
}

func (p *Store) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	namespace, id := splitID(id)

	req := p.common(namespace, p.k8sClient.Get()).
		Name(id)

	return p.singleResult(schema, req)
}

func (p *Store) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	namespace := getNamespace(apiContext, opt)

	req := p.common(namespace, p.k8sClient.Get())

	resultList := &unstructured.UnstructuredList{}
	err := req.Do().Into(resultList)
	if err != nil {
		return nil, err
	}

	result := []map[string]interface{}{}

	for _, obj := range resultList.Items {
		result = append(result, p.fromInternal(schema, obj.Object))
	}

	return result, nil
}

func getNamespace(apiContext *types.APIContext, opt *types.QueryOptions) string {
	if val, ok := apiContext.SubContext["namespace"]; ok {
		return convert.ToString(val)
	}

	if opt == nil {
		return ""
	}

	for _, condition := range opt.Conditions {
		if condition.Field == "namespace" && len(condition.Values) > 0 {
			return convert.ToString(condition.Values[0])
		}
	}

	return ""
}

func (p *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	namespace, _ := data["namespace"].(string)
	p.toInternal(schema.Mapper, data)

	req := p.common(namespace, p.k8sClient.Post()).
		Body(&unstructured.Unstructured{
			Object: data,
		})

	return p.singleResult(schema, req)
}

func (p *Store) toInternal(mapper types.Mapper, data map[string]interface{}) {
	if mapper != nil {
		mapper.ToInternal(data)
	}

	if p.group == "" {
		data["apiVersion"] = p.version
	} else {
		data["apiVersion"] = p.group + "/" + p.version
	}
	data["kind"] = p.kind
}

func (p *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	existing, err := p.ByID(apiContext, schema, id)
	if err != nil {
		return data, nil
	}

	for k, v := range data {
		existing[k] = v
	}

	p.toInternal(schema.Mapper, existing)
	namespace, id := splitID(id)

	req := p.common(namespace, p.k8sClient.Put()).
		Body(&unstructured.Unstructured{
			Object: existing,
		}).
		Name(id)

	return p.singleResult(schema, req)
}

func (p *Store) Delete(apiContext *types.APIContext, schema *types.Schema, id string) error {
	namespace, id := splitID(id)

	prop := metav1.DeletePropagationForeground
	req := p.common(namespace, p.k8sClient.Delete()).
		Body(&metav1.DeleteOptions{
			PropagationPolicy: &prop,
		}).
		Name(id)

	return req.Do().Error()
}

func (p *Store) singleResult(schema *types.Schema, req *rest.Request) (map[string]interface{}, error) {
	result := &unstructured.Unstructured{}
	err := req.Do().Into(result)
	if err != nil {
		return nil, err
	}

	p.fromInternal(schema, result.Object)
	return result.Object, nil
}

func splitID(id string) (string, string) {
	namespace := ""
	parts := strings.SplitN(id, ":", 2)
	if len(parts) == 2 {
		namespace = parts[0]
		id = parts[1]
	}

	return namespace, id
}

func (p *Store) common(namespace string, req *rest.Request) *rest.Request {
	prefix := append([]string{}, p.prefix...)
	if p.group != "" {
		prefix = append(prefix, p.group)
	}
	prefix = append(prefix, p.version)
	req.Prefix(prefix...).
		Resource(p.resourcePlural)

	if namespace != "" {
		req.Namespace(namespace)
	}

	return req
}

func (p *Store) fromInternal(schema *types.Schema, data map[string]interface{}) map[string]interface{} {
	if schema.Mapper != nil {
		schema.Mapper.FromInternal(data)
	}
	data["type"] = schema.ID
	name, _ := data["name"].(string)
	namespace, _ := data["namespace"].(string)

	if name != "" {
		if namespace == "" {
			data["id"] = name
		} else {
			data["id"] = namespace + ":" + name
		}
	}

	return data
}
