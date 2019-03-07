package types

import (
	"encoding/json"
	"net/http"
	"net/url"

	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/rancher/norman/pkg/data"
)

type ValuesMap struct {
	Foo map[string]interface{}
}

type RawResource struct {
	ID           string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Type         string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Schema       *Schema                `json:"-" yaml:"-"`
	Links        map[string]string      `json:"links,omitempty" yaml:"links,omitempty"`
	Actions      map[string]string      `json:"actions,omitempty" yaml:"actions,omitempty"`
	Values       map[string]interface{} `json:",inline" yaml:",inline"`
	ActionLinks  bool                   `json:"-" yaml:"-"`
	DropReadOnly bool                   `json:"-" yaml:"-"`
}

func (r *RawResource) AddAction(apiOp *APIOperation, name string) {
	r.Actions[name] = apiOp.URLBuilder.Action(r.Schema, r.ID, name)
}

func (r *RawResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.ToMap())
}

func (r *RawResource) ToMap() map[string]interface{} {
	data := data.New()
	for k, v := range r.Values {
		data[k] = v
	}

	if r.ID != "" && !r.DropReadOnly {
		data["id"] = r.ID
	}

	if r.Type != "" && !r.DropReadOnly {
		data["type"] = r.Type
	}

	if len(r.Links) > 0 && !r.DropReadOnly {
		data["links"] = r.Links
	}

	if len(r.Actions) > 0 && !r.DropReadOnly {
		if r.ActionLinks {
			data["actionLinks"] = r.Actions
		} else {
			data["actions"] = r.Actions
		}
	}
	return data
}

type ActionHandler func(actionName string, action *Action, request *APIOperation) error

type RequestHandler func(request *APIOperation, next RequestHandler) (interface{}, error)

type QueryFilter func(opts *QueryOptions, schema *Schema, data []map[string]interface{}) []map[string]interface{}

type Validator func(request *APIOperation, schema *Schema, data map[string]interface{}) error

type InputFormatter func(request *APIOperation, schema *Schema, data map[string]interface{}, create bool) error

type Formatter func(request *APIOperation, resource *RawResource)

type CollectionFormatter func(request *APIOperation, collection *GenericCollection)

type ErrorHandler func(request *APIOperation, err error)

type ResponseWriter interface {
	Write(apiOp *APIOperation, code int, obj interface{})
}

type AccessControl interface {
	CanCreate(apiOp *APIOperation, schema *Schema) error
	CanList(apiOp *APIOperation, schema *Schema) error
	CanGet(apiOp *APIOperation, schema *Schema) error
	CanUpdate(apiOp *APIOperation, obj map[string]interface{}, schema *Schema) error
	CanDelete(apiOp *APIOperation, obj map[string]interface{}, schema *Schema) error
	// CanDo function should not yet be used if a corresponding specific method exists. It has been added to
	// satisfy a specific usecase for the short term until full-blown dynamic RBAC can be implemented.
	CanDo(apiGroup, resource, verb string, apiOp *APIOperation, obj map[string]interface{}, schema *Schema) error

	Filter(apiOp *APIOperation, schema *Schema, obj map[string]interface{}) map[string]interface{}
	FilterList(apiOp *APIOperation, schema *Schema, obj []map[string]interface{}) []map[string]interface{}
}

type APIOperation struct {
	Action             string
	Name               string
	Type               string
	Link               string
	Method             string
	Namespaces         []string
	Schema             *Schema
	Schemas            *Schemas
	Query              url.Values
	ResponseFormat     string
	ReferenceValidator ReferenceValidator
	ResponseWriter     ResponseWriter
	QueryFilter        QueryFilter
	URLBuilder         URLBuilder
	AccessControl      AccessControl
	Pagination         *Pagination

	Request  *http.Request
	Response http.ResponseWriter
}

func (r *APIOperation) GetUser() string {
	user, ok := request.UserFrom(r.Request.Context())
	if ok {
		return user.GetName()
	}
	return ""
}

func (r *APIOperation) Option(key string) string {
	return r.Query.Get("_" + key)
}

func (r *APIOperation) WriteResponse(code int, obj interface{}) {
	r.ResponseWriter.Write(r, code, obj)
}

func (r *APIOperation) FilterList(opts *QueryOptions, schema *Schema, obj []map[string]interface{}) []map[string]interface{} {
	return r.QueryFilter(opts, schema, obj)
}

func (r *APIOperation) FilterObject(opts *QueryOptions, schema *Schema, obj map[string]interface{}) map[string]interface{} {
	opts.Pagination = nil
	result := r.QueryFilter(opts, schema, []map[string]interface{}{obj})
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

func (r *APIOperation) Filter(opts *QueryOptions, schema *Schema, obj interface{}) interface{} {
	switch v := obj.(type) {
	case []map[string]interface{}:
		return r.FilterList(opts, schema, v)
	case map[string]interface{}:
		return r.FilterObject(opts, schema, v)
	}

	return nil
}

var (
	ASC  = SortOrder("asc")
	DESC = SortOrder("desc")
)

type QueryOptions struct {
	Sort       Sort
	Pagination *Pagination
	Conditions []*QueryCondition
}

type ReferenceValidator interface {
	Validate(resourceType, resourceID string) bool
	Lookup(resourceType, resourceID string) *RawResource
}

type URLBuilder interface {
	Current() string

	Collection(schema *Schema) string
	CollectionAction(schema *Schema, action string) string
	ResourceLink(schema *Schema, id string) string
	Link(schema *Schema, id string, linkName string) string
	FilterLink(schema *Schema, fieldName string, value string) string
	Action(schema *Schema, id string, action string) string

	RelativeToRoot(path string) string
	Marker(marker string) string
	ReverseSort(order SortOrder) string
	Sort(field string) string
}

type Store interface {
	ByID(apiOp *APIOperation, schema *Schema, id string) (map[string]interface{}, error)
	List(apiOp *APIOperation, schema *Schema, opt *QueryOptions) ([]map[string]interface{}, error)
	Create(apiOp *APIOperation, schema *Schema, data map[string]interface{}) (map[string]interface{}, error)
	Update(apiOp *APIOperation, schema *Schema, data map[string]interface{}, id string) (map[string]interface{}, error)
	Delete(apiOp *APIOperation, schema *Schema, id string) (map[string]interface{}, error)
	Watch(apiOp *APIOperation, schema *Schema, opt *QueryOptions) (chan map[string]interface{}, error)
}
