package generator

var controllerTemplate = `package {{.schema.Version.Version}}

import (
	"context"

	{{.importPackage}}
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/controller"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	{{.schema.CodeName}}GroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "{{.schema.CodeName}}",
	}
	{{.schema.CodeName}}Resource = metav1.APIResource{
		Name:         "{{.schema.PluralName | toLower}}",
		SingularName: "{{.schema.ID | toLower}}",
{{- if eq .schema.Scope "namespace" }}
		Namespaced:   true,
{{ else }}
		Namespaced:   false,
{{- end }}
		Kind:         {{.schema.CodeName}}GroupVersionKind.Kind,
	}
)

type {{.schema.CodeName}}List struct {
	metav1.TypeMeta   %BACK%json:",inline"%BACK%
	metav1.ListMeta   %BACK%json:"metadata,omitempty"%BACK%
	Items             []{{.prefix}}{{.schema.CodeName}}
}

type {{.schema.CodeName}}HandlerFunc func(key string, obj *{{.prefix}}{{.schema.CodeName}}) error

type {{.schema.CodeName}}Lister interface {
	List(namespace string, selector labels.Selector) (ret []*{{.prefix}}{{.schema.CodeName}}, err error)
	Get(namespace, name string) (*{{.prefix}}{{.schema.CodeName}}, error)
}

type {{.schema.CodeName}}Controller interface {
	Informer() cache.SharedIndexInformer
	Lister() {{.schema.CodeName}}Lister
	AddHandler(handler {{.schema.CodeName}}HandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type {{.schema.CodeName}}Interface interface {
    ObjectClient() *clientbase.ObjectClient
	Create(*{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
	GetNamespace(name, namespace string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error)
	Get(name string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error)
	Update(*{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespace(name, namespace string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*{{.schema.CodeName}}List, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() {{.schema.CodeName}}Controller
	AddSyncHandler(sync {{.schema.CodeName}}HandlerFunc)
	AddLifecycle(name string, lifecycle {{.schema.CodeName}}Lifecycle)
}

type {{.schema.ID}}Lister struct {
	controller *{{.schema.ID}}Controller
}

func (l *{{.schema.ID}}Lister) List(namespace string, selector labels.Selector) (ret []*{{.prefix}}{{.schema.CodeName}}, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*{{.prefix}}{{.schema.CodeName}}))
	})
	return
}

func (l *{{.schema.ID}}Lister) Get(namespace, name string) (*{{.prefix}}{{.schema.CodeName}}, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group: {{.schema.CodeName}}GroupVersionKind.Group,
			Resource: "{{.schema.ID}}",
		}, name)
	}
	return obj.(*{{.prefix}}{{.schema.CodeName}}), nil
}

type {{.schema.ID}}Controller struct {
	controller.GenericController
}

func (c *{{.schema.ID}}Controller) Lister() {{.schema.CodeName}}Lister {
	return &{{.schema.ID}}Lister{
		controller: c,
	}
}


func (c *{{.schema.ID}}Controller) AddHandler(handler {{.schema.CodeName}}HandlerFunc) {
	c.GenericController.AddHandler(func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*{{.prefix}}{{.schema.CodeName}}))
	})
}

type {{.schema.ID}}Factory struct {
}

func (c {{.schema.ID}}Factory) Object() runtime.Object {
	return &{{.prefix}}{{.schema.CodeName}}{}
}

func (c {{.schema.ID}}Factory) List() runtime.Object {
	return &{{.schema.CodeName}}List{}
}

func (s *{{.schema.ID}}Client) Controller() {{.schema.CodeName}}Controller {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.{{.schema.ID}}Controllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController({{.schema.CodeName}}GroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &{{.schema.ID}}Controller{
		GenericController: genericController,
	}

	s.client.{{.schema.ID}}Controllers[s.ns] = c
    s.client.starters = append(s.client.starters, c)

	return c
}

type {{.schema.ID}}Client struct {
	client *Client
	ns string
	objectClient *clientbase.ObjectClient
	controller   {{.schema.CodeName}}Controller
}

func (s *{{.schema.ID}}Client) ObjectClient() *clientbase.ObjectClient {
	return s.objectClient
}

func (s *{{.schema.ID}}Client) Create(o *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Get(name string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) GetNamespace(name, namespace string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.GetNamespace(name, namespace, opts)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Update(o *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *{{.schema.ID}}Client) DeleteNamespace(name, namespace string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespace(name, namespace, options)
}

func (s *{{.schema.ID}}Client) List(opts metav1.ListOptions) (*{{.schema.CodeName}}List, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*{{.schema.CodeName}}List), err
}

func (s *{{.schema.ID}}Client) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *{{.schema.ID}}Client) Patch(o *{{.prefix}}{{.schema.CodeName}}, data []byte, subresources ...string) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *{{.schema.ID}}Client) AddSyncHandler(sync {{.schema.CodeName}}HandlerFunc) {
	s.Controller().AddHandler(sync)
}

func (s *{{.schema.ID}}Client) AddLifecycle(name string, lifecycle {{.schema.CodeName}}Lifecycle) {
	sync := New{{.schema.CodeName}}LifecycleAdapter(name, s, lifecycle)
	s.AddSyncHandler(sync)
}
`
