package objectclient

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/restwatch"
	"github.com/sirupsen/logrus"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	json2 "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	restclientwatch "k8s.io/client-go/rest/watch"
)

type ObjectFactory interface {
	Object() runtime.Object
	List() runtime.Object
}

type UnstructuredObjectFactory struct {
}

func (u *UnstructuredObjectFactory) Object() runtime.Object {
	return &unstructured.Unstructured{}
}

func (u *UnstructuredObjectFactory) List() runtime.Object {
	return &unstructured.UnstructuredList{}
}

type GenericClient interface {
	UnstructuredClient() GenericClient
	GroupVersionKind() schema.GroupVersionKind
	Create(o runtime.Object) (runtime.Object, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (runtime.Object, error)
	Get(name string, opts metav1.GetOptions) (runtime.Object, error)
	Update(name string, o runtime.Object) (runtime.Object, error)
	DeleteNamespaced(namespace, name string, opts *metav1.DeleteOptions) error
	Delete(name string, opts *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	ListNamespaced(namespace string, opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOptions *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Patch(name string, o runtime.Object, patchType types.PatchType, data []byte, subresources ...string) (runtime.Object, error)
	ObjectFactory() ObjectFactory
}

type ObjectClient struct {
	restClient rest.Interface
	resource   *metav1.APIResource
	gvk        schema.GroupVersionKind
	ns         string
	Factory    ObjectFactory
}

func NewObjectClient(namespace string, restClient rest.Interface, apiResource *metav1.APIResource, gvk schema.GroupVersionKind, factory ObjectFactory) *ObjectClient {
	return &ObjectClient{
		restClient: restClient,
		resource:   apiResource,
		gvk:        gvk,
		ns:         namespace,
		Factory:    factory,
	}
}

func (p *ObjectClient) UnstructuredClient() GenericClient {
	return &ObjectClient{
		restClient: p.restClient,
		resource:   p.resource,
		gvk:        p.gvk,
		ns:         p.ns,
		Factory:    &UnstructuredObjectFactory{},
	}
}

func (p *ObjectClient) GroupVersionKind() schema.GroupVersionKind {
	return p.gvk
}

func (p *ObjectClient) getAPIPrefix() string {
	if p.gvk.Group == "" {
		return "api"
	}
	return "apis"
}

func (p *ObjectClient) Create(o runtime.Object) (runtime.Object, error) {
	ns := p.ns
	obj, ok := o.(metav1.Object)
	if ok && obj.GetNamespace() != "" {
		ns = obj.GetNamespace()
	}

	if ok {
		labels := obj.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		} else {
			ls := make(map[string]string)
			for k, v := range labels {
				ls[k] = v
			}
			labels = ls

		}
		labels["cattle.io/creator"] = "norman"
		obj.SetLabels(labels)
	}

	if t, err := meta.TypeAccessor(o); err == nil {
		if t.GetKind() == "" {
			t.SetKind(p.gvk.Kind)
		}
		if t.GetAPIVersion() == "" {
			apiVersion, _ := p.gvk.ToAPIVersionAndKind()
			t.SetAPIVersion(apiVersion)
		}
	}
	result := p.Factory.Object()
	logrus.Tracef("REST CREATE %s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, ns, p.resource.Name)
	err := p.restClient.Post().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		Body(o).
		Do(context.TODO()).
		Into(result)
	return result, err
}

func (p *ObjectClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (runtime.Object, error) {
	result := p.Factory.Object()
	req := p.restClient.Get().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version)
	if namespace != "" {
		req = req.Namespace(namespace)
	}
	err := req.
		Resource(p.resource.Name).
		VersionedParams(&opts, metav1.ParameterCodec).
		Name(name).
		Do(context.TODO()).
		Into(result)
	logrus.Tracef("REST GET %s/%s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, namespace, p.resource.Name, name)
	return result, err

}

func (p *ObjectClient) Get(name string, opts metav1.GetOptions) (runtime.Object, error) {
	result := p.Factory.Object()
	err := p.restClient.Get().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(p.ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		VersionedParams(&opts, metav1.ParameterCodec).
		Name(name).
		Do(context.TODO()).
		Into(result)
	logrus.Tracef("REST GET %s/%s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, p.ns, p.resource.Name, name)
	return result, err
}

func (p *ObjectClient) Update(name string, o runtime.Object) (runtime.Object, error) {
	ns := p.ns
	if obj, ok := o.(metav1.Object); ok && obj.GetNamespace() != "" {
		ns = obj.GetNamespace()
	}
	result := p.Factory.Object()
	if len(name) == 0 {
		return result, errors.New("object missing name")
	}
	logrus.Tracef("REST UPDATE %s/%s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, ns, p.resource.Name, name)
	err := p.restClient.Put().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		Name(name).
		Body(o).
		Do(context.TODO()).
		Into(result)
	return result, err
}

func (p *ObjectClient) DeleteNamespaced(namespace, name string, opts *metav1.DeleteOptions) error {
	req := p.restClient.Delete().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version)
	if namespace != "" {
		req = req.Namespace(namespace)
	}
	logrus.Tracef("REST DELETE %s/%s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, namespace, p.resource.Name, name)
	return req.Resource(p.resource.Name).
		Name(name).
		Body(opts).
		Do(context.TODO()).
		Error()
}

func (p *ObjectClient) Delete(name string, opts *metav1.DeleteOptions) error {
	logrus.Tracef("REST DELETE %s/%s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, p.ns, p.resource.Name, name)
	return p.restClient.Delete().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(p.ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		Name(name).
		Body(opts).
		Do(context.TODO()).
		Error()
}

func (p *ObjectClient) List(opts metav1.ListOptions) (runtime.Object, error) {
	result := p.Factory.List()
	logrus.Tracef("REST LIST %s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, p.ns, p.resource.Name)

	// If ResourceVersion is set to 0 then the Limit is ignored on the API side. Usually
	// that's not a problem, but with very large counts of a single object type the client will
	// hit it's connection timeout
	if opts.ResourceVersion == "0" {
		opts.ResourceVersion = ""
	}

	return result, p.restClient.Get().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(p.ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		VersionedParams(&opts, metav1.ParameterCodec).
		Do(context.TODO()).
		Into(result)
}

func (p *ObjectClient) ListNamespaced(namespace string, opts metav1.ListOptions) (runtime.Object, error) {
	result := p.Factory.List()
	logrus.Tracef("REST LIST %s/%s/%s/%s/%s", p.getAPIPrefix(), p.gvk.Group, p.gvk.Version, namespace, p.resource.Name)

	if opts.ResourceVersion == "0" {
		opts.ResourceVersion = ""
	}

	return result, p.restClient.Get().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(namespace, p.resource.Namespaced).
		Resource(p.resource.Name).
		VersionedParams(&opts, metav1.ParameterCodec).
		Do(context.TODO()).
		Into(result)
}

func (p *ObjectClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	restClient := p.restClient
	if watchClient, ok := restClient.(restwatch.WatchClient); ok {
		restClient = watchClient.WatchClient()
	}

	r, err := restClient.Get().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		Prefix("watch").
		NamespaceIfScoped(p.ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		VersionedParams(&opts, metav1.ParameterCodec).
		Stream(context.TODO())
	if err != nil {
		return nil, err
	}

	embeddedDecoder := &structuredDecoder{
		factory: p.Factory,
	}
	streamDecoder := streaming.NewDecoder(json2.Framer.NewFrameReader(r), embeddedDecoder)
	decoder := restclientwatch.NewDecoder(streamDecoder, embeddedDecoder)
	return watch.NewStreamWatcher(decoder, k8sError.NewClientErrorReporter(http.StatusInternalServerError, "watch", "ClientWatchDecoding")), nil
}

type structuredDecoder struct {
	factory ObjectFactory
}

func (d *structuredDecoder) Decode(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	if into == nil {
		into = d.factory.Object()
	}

	err := json.Unmarshal(data, &into)
	if err != nil {
		status := &metav1.Status{}
		if err := json.Unmarshal(data, status); err == nil && strings.ToLower(status.Kind) == "status" {
			return status, defaults, nil
		}
		return nil, nil, err
	}

	if _, ok := into.(*metav1.Status); !ok && strings.ToLower(into.GetObjectKind().GroupVersionKind().Kind) == "status" {
		into = &metav1.Status{}
		err := json.Unmarshal(data, into)
		if err != nil {
			return nil, nil, err
		}
	}

	return into, defaults, err
}

func (p *ObjectClient) DeleteCollection(deleteOptions *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return p.restClient.Delete().
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(p.ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		VersionedParams(&listOptions, metav1.ParameterCodec).
		Body(deleteOptions).
		Do(context.TODO()).
		Error()
}

func (p *ObjectClient) Patch(name string, o runtime.Object, patchType types.PatchType, data []byte, subresources ...string) (runtime.Object, error) {
	ns := p.ns
	if obj, ok := o.(metav1.Object); ok && obj.GetNamespace() != "" {
		ns = obj.GetNamespace()
	}
	result := p.Factory.Object()
	if len(name) == 0 {
		return result, errors.New("object missing name")
	}
	err := p.restClient.Patch(patchType).
		Prefix(p.getAPIPrefix(), p.gvk.Group, p.gvk.Version).
		NamespaceIfScoped(ns, p.resource.Namespaced).
		Resource(p.resource.Name).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(context.TODO()).
		Into(result)
	return result, err
}

func (p *ObjectClient) ObjectFactory() ObjectFactory {
	return p.Factory
}
