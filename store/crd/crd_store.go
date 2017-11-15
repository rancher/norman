package crd

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

type Store struct {
	schemas         []*types.Schema
	apiExtClientSet apiextclientset.Interface
	k8sClient       rest.Interface
	schemaStatus    map[*types.Schema]*apiext.CustomResourceDefinition
}

func NewCRDStore(apiExtClientSet apiextclientset.Interface, k8sClient rest.Interface) *Store {
	return &Store{
		apiExtClientSet: apiExtClientSet,
		k8sClient:       k8sClient,
		schemaStatus:    map[*types.Schema]*apiext.CustomResourceDefinition{},
	}
}

func (c *Store) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	crd, ok := c.schemaStatus[schema]
	if !ok {
		return nil, nil
	}

	namespace := ""
	parts := strings.SplitN(id, ":", 2)

	if len(parts) == 2 {
		namespace = parts[0]
		id = parts[1]
	}

	req := c.k8sClient.Get().
		Prefix("apis", crd.Spec.Group, crd.Spec.Version).
		Resource(crd.Status.AcceptedNames.Plural).
		Name(id)

	if namespace != "" {
		req.Namespace(namespace)
	}

	result := &unstructured.Unstructured{}
	err := req.Do().Into(result)
	if err != nil {
		return nil, err
	}

	c.fromInternal(result.Object, schema)

	return result.Object, nil
}

func (c *Store) toInternal(data map[string]interface{}, schema *types.Schema) {
	if schema.Mapper != nil {
		schema.Mapper.ToInternal(data)
	}
}

func (c *Store) fromInternal(data map[string]interface{}, schema *types.Schema) {
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

	if status, ok := c.schemaStatus[schema]; ok {
		if status.Spec.Scope != apiext.NamespaceScoped {
			delete(data, "namespace")
		}
	}
}

func (c *Store) Delete(apiContext *types.APIContext, schema *types.Schema, id string) error {
	crd, ok := c.schemaStatus[schema]
	if !ok {
		return nil
	}

	namespace := ""
	parts := strings.SplitN(id, ":", 2)

	if len(parts) == 2 {
		namespace = parts[0]
		id = parts[1]
	}

	prop := metav1.DeletePropagationForeground
	req := c.k8sClient.Delete().
		Prefix("apis", crd.Spec.Group, crd.Spec.Version).
		Resource(crd.Status.AcceptedNames.Plural).
		Body(&metav1.DeleteOptions{
			PropagationPolicy: &prop,
		}).
		Name(id)

	if namespace != "" {
		req.Namespace(namespace)
	}

	result := &unstructured.Unstructured{}
	return req.Do().Into(result)
}

func (c *Store) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	crd, ok := c.schemaStatus[schema]
	if !ok {
		return nil, nil
	}

	req := c.k8sClient.Get().
		Prefix("apis", crd.Spec.Group, crd.Spec.Version).
		Resource(crd.Status.AcceptedNames.Plural)

	resultList := &unstructured.UnstructuredList{}
	err := req.Do().Into(resultList)
	if err != nil {
		return nil, err
	}

	result := []map[string]interface{}{}

	for _, obj := range resultList.Items {
		c.fromInternal(obj.Object, schema)
		result = append(result, obj.Object)
	}

	return result, nil
}

func (c *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	crd, ok := c.schemaStatus[schema]
	if !ok {
		return nil, nil
	}

	namespace := ""
	parts := strings.SplitN(id, ":", 2)

	if len(parts) == 2 {
		namespace = parts[0]
		id = parts[1]
	}

	req := c.k8sClient.Get().
		Prefix("apis", crd.Spec.Group, crd.Spec.Version).
		Resource(crd.Status.AcceptedNames.Plural).
		Name(id)

	if namespace != "" {
		req.Namespace(namespace)
	}

	result := &unstructured.Unstructured{}
	err := req.Do().Into(result)
	if err != nil {
		return nil, err
	}

	c.fromInternal(result.Object, schema)

	for k, v := range data {
		if k == "metadata" {
			continue
		}
		result.Object[k] = v
	}

	c.toInternal(result.Object, schema)

	req = c.k8sClient.Put().
		Prefix("apis", crd.Spec.Group, crd.Spec.Version).
		Resource(crd.Status.AcceptedNames.Plural).
		Body(result).
		Name(id)

	if namespace != "" {
		req.Namespace(namespace)
	}

	result = &unstructured.Unstructured{}
	err = req.Do().Into(result)
	if err != nil {
		return nil, err
	}

	c.fromInternal(result.Object, schema)
	return result.Object, nil
}

func (c *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	crd, ok := c.schemaStatus[schema]
	if !ok {
		return nil, nil
	}

	namespace, _ := data["namespace"].(string)

	data["apiVersion"] = crd.Spec.Group + "/" + crd.Spec.Version
	data["kind"] = crd.Status.AcceptedNames.Kind

	c.toInternal(data, schema)

	req := c.k8sClient.Post().
		Prefix("apis", crd.Spec.Group, crd.Spec.Version).
		Resource(crd.Status.AcceptedNames.Plural).
		Body(&unstructured.Unstructured{
			Object: data,
		})

	if crd.Spec.Scope == apiext.NamespaceScoped {
		req.Namespace(namespace)
	}

	result := &unstructured.Unstructured{}
	err := req.Do().Into(result)
	if err != nil {
		return nil, err
	}

	c.fromInternal(result.Object, schema)
	return result.Object, nil
}

func (c *Store) AddSchemas(ctx context.Context, schemas *types.Schemas) error {
	if schemas.Err() != nil {
		return schemas.Err()
	}

	for _, schema := range schemas.Schemas() {
		if schema.Store != nil || !contains(schema.CollectionMethods, http.MethodGet) {
			continue
		}

		schema.Store = c
		c.schemas = append(c.schemas, schema)
	}

	ready, err := c.getReadyCRDs()
	if err != nil {
		return err
	}

	for _, schema := range c.schemas {
		crd, err := c.createCRD(schema, ready)
		if err != nil {
			return err
		}
		c.schemaStatus[schema] = crd
	}

	ready, err = c.getReadyCRDs()
	if err != nil {
		return err
	}

	for schema, crd := range c.schemaStatus {
		if _, ok := ready[crd.Name]; !ok {
			if err := c.waitCRD(ctx, crd.Name, schema); err != nil {
				return err
			}
		}
	}

	return nil
}

func contains(list []string, s string) bool {
	for _, i := range list {
		if i == s {
			return true
		}
	}

	return false
}

func (c *Store) waitCRD(ctx context.Context, crdName string, schema *types.Schema) error {
	logrus.Infof("Waiting for CRD %s to become available", crdName)
	defer logrus.Infof("Done waiting for CRD %s to become available", crdName)

	first := true
	return wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		if !first {
			logrus.Infof("Waiting for CRD %s to become available", crdName)
		}
		first = false

		crd, err := c.apiExtClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiext.Established:
				if cond.Status == apiext.ConditionTrue {
					c.schemaStatus[schema] = crd
					return true, err
				}
			case apiext.NamesAccepted:
				if cond.Status == apiext.ConditionFalse {
					logrus.Infof("Name conflict on %s: %v\n", crdName, cond.Reason)
				}
			}
		}

		return false, ctx.Err()
	})
}

func (c *Store) createCRD(schema *types.Schema, ready map[string]apiext.CustomResourceDefinition) (*apiext.CustomResourceDefinition, error) {
	plural := strings.ToLower(schema.PluralName)
	name := strings.ToLower(plural + "." + schema.Version.Group)

	crd, ok := ready[name]
	if ok {
		return &crd, nil
	}

	crd = apiext.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiext.CustomResourceDefinitionSpec{
			Group:   schema.Version.Group,
			Version: schema.Version.Version,
			//Scope:   getScope(schema),
			Scope: apiext.ClusterScoped,
			Names: apiext.CustomResourceDefinitionNames{
				Plural: plural,
				Kind:   capitalize(schema.ID),
			},
		},
	}

	logrus.Infof("Creating CRD %s", name)
	_, err := c.apiExtClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crd)
	if errors.IsAlreadyExists(err) {
		return &crd, nil
	}
	return &crd, err
}

func (c *Store) getReadyCRDs() (map[string]apiext.CustomResourceDefinition, error) {
	list, err := c.apiExtClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := map[string]apiext.CustomResourceDefinition{}

	for _, crd := range list.Items {
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiext.Established:
				if cond.Status == apiext.ConditionTrue {
					result[crd.Name] = crd
				}
			}
		}
	}

	return result, nil
}

func getScope(schema *types.Schema) apiext.ResourceScope {
	for name := range schema.ResourceFields {
		if name == "namespace" {
			return apiext.NamespaceScoped
		}
	}

	return apiext.ClusterScoped
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[0:1]) + s[1:]
}
