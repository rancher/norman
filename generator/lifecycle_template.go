package generator

var lifecycleTemplate = `package {{.schema.Version.Version}}

import (
	{{.importPackage}}
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/rancher/norman/lifecycle"
)

type {{.schema.CodeName}}Lifecycle interface {
	Create(obj *{{.prefix}}{{.schema.CodeName}}) error
	Remove(obj *{{.prefix}}{{.schema.CodeName}}) error
	Updated(obj *{{.prefix}}{{.schema.CodeName}}) error
}

type {{.schema.ID}}LifecycleAdapter struct {
	lifecycle {{.schema.CodeName}}Lifecycle
}

func (w *{{.schema.ID}}LifecycleAdapter) Create(obj runtime.Object) error {
	return w.lifecycle.Create(obj.(*{{.prefix}}{{.schema.CodeName}}))
}

func (w *{{.schema.ID}}LifecycleAdapter) Finalize(obj runtime.Object) error {
	return w.lifecycle.Remove(obj.(*{{.prefix}}{{.schema.CodeName}}))
}

func (w *{{.schema.ID}}LifecycleAdapter) Updated(obj runtime.Object) error {
	return w.lifecycle.Updated(obj.(*{{.prefix}}{{.schema.CodeName}}))
}

func New{{.schema.CodeName}}LifecycleAdapter(name string, client {{.schema.CodeName}}Interface, l {{.schema.CodeName}}Lifecycle) {{.schema.CodeName}}HandlerFunc {
	adapter := &{{.schema.ID}}LifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, adapter, client.ObjectClient())
	return func(key string, obj *{{.prefix}}{{.schema.CodeName}}) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
`
