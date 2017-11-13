package generator

var controllerTemplate = `package {{.schema.Version.Version}}

import (
	"sync"

	"context"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	{{.schema.CodeName}}GroupVersionKind = schema.GroupVersionKind{
		Version: "{{.schema.Version.Version}}",
		Group:   "{{.schema.Version.Group}}",
		Kind:    "{{.schema.CodeName}}",
	}
	{{.schema.CodeName}}Resource = metav1.APIResource{
		Name:         "{{.schema.PluralName | toLower}}",
		SingularName: "{{.schema.ID | toLower}}",
		Namespaced:   false,
		Kind:         {{.schema.CodeName}}GroupVersionKind.Kind,
	}
)

type {{.schema.CodeName}}List struct {
	metav1.TypeMeta   %BACK%json:",inline"%BACK%
	metav1.ObjectMeta %BACK%json:"metadata,omitempty"%BACK%
	Items             []{{.schema.CodeName}}
}

type {{.schema.CodeName}}HandlerFunc func(key string, obj *{{.schema.CodeName}}) error

type {{.schema.CodeName}}Controller interface {
	Informer() cache.SharedIndexInformer
	AddHandler(handler {{.schema.CodeName}}HandlerFunc)
	Enqueue(namespace, name string)
	Start(threadiness int, ctx context.Context) error
}

type {{.schema.CodeName}}Interface interface {
	Create(*{{.schema.CodeName}}) (*{{.schema.CodeName}}, error)
	Get(name string, opts metav1.GetOptions) (*{{.schema.CodeName}}, error)
	Update(*{{.schema.CodeName}}) (*{{.schema.CodeName}}, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*{{.schema.CodeName}}List, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() {{.schema.CodeName}}Controller
}

type {{.schema.ID}}Controller struct {
	controller.GenericController
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
		return handler(key, obj.(*{{.schema.CodeName}}))
	})
}

type {{.schema.ID}}Factory struct {
}

func (c {{.schema.ID}}Factory) Object() runtime.Object {
	return &{{.schema.CodeName}}{}
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

	return c
}

type {{.schema.ID}}Client struct {
	client *Client
	ns string
	objectClient *clientbase.ObjectClient
	controller   {{.schema.CodeName}}Controller
}

func (s *{{.schema.ID}}Client) Create(o *{{.schema.CodeName}}) (*{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Get(name string, opts metav1.GetOptions) (*{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Update(o *{{.schema.CodeName}}) (*{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *{{.schema.ID}}Client) List(opts metav1.ListOptions) (*{{.schema.CodeName}}List, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*{{.schema.CodeName}}List), err
}

func (s *{{.schema.ID}}Client) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

func (s *{{.schema.ID}}Client) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}
`
