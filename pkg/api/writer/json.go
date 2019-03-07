package writer

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rancher/norman/pkg/parse"
	builder2 "github.com/rancher/norman/pkg/parse/builder"

	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/norman/pkg/types/definition"
	"github.com/sirupsen/logrus"
)

type EncodingResponseWriter struct {
	ContentType string
	Encoder     func(io.Writer, interface{}) error
}

func (j *EncodingResponseWriter) start(apiOp *types.APIOperation, code int, obj interface{}) {
	AddCommonResponseHeader(apiOp)
	apiOp.Response.Header().Set("content-type", j.ContentType)
	apiOp.Response.WriteHeader(code)
}

func (j *EncodingResponseWriter) Write(apiOp *types.APIOperation, code int, obj interface{}) {
	j.start(apiOp, code, obj)
	j.Body(apiOp, apiOp.Response, obj)
}

func (j *EncodingResponseWriter) Body(apiOp *types.APIOperation, writer io.Writer, obj interface{}) error {
	return j.VersionBody(apiOp, writer, obj)

}

func (j *EncodingResponseWriter) VersionBody(apiOp *types.APIOperation, writer io.Writer, obj interface{}) error {
	var output interface{}

	builder := builder2.NewBuilder(apiOp)

	switch v := obj.(type) {
	case []interface{}:
		output = j.writeInterfaceSlice(builder, apiOp, v)
	case []map[string]interface{}:
		output = j.writeMapSlice(builder, apiOp, v)
	case map[string]interface{}:
		output = j.convert(builder, apiOp, v)
	case types.RawResource:
		output = v
	}

	if output != nil {
		return j.Encoder(writer, output)
	}

	return nil
}
func (j *EncodingResponseWriter) writeMapSlice(builder *builder2.Builder, apiOp *types.APIOperation, input []map[string]interface{}) *types.GenericCollection {
	collection := newCollection(apiOp)
	for _, value := range input {
		converted := j.convert(builder, apiOp, value)
		if converted != nil {
			collection.Data = append(collection.Data, converted)
		}
	}

	if apiOp.Schema.CollectionFormatter != nil {
		apiOp.Schema.CollectionFormatter(apiOp, collection)
	}

	return collection
}

func (j *EncodingResponseWriter) writeInterfaceSlice(builder *builder2.Builder, apiOp *types.APIOperation, input []interface{}) *types.GenericCollection {
	collection := newCollection(apiOp)
	for _, value := range input {
		switch v := value.(type) {
		case map[string]interface{}:
			converted := j.convert(builder, apiOp, v)
			if converted != nil {
				collection.Data = append(collection.Data, converted)
			}
		default:
			collection.Data = append(collection.Data, v)
		}
	}

	if apiOp.Schema.CollectionFormatter != nil {
		apiOp.Schema.CollectionFormatter(apiOp, collection)
	}

	return collection
}

func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	return fmt.Sprint(val)
}

func (j *EncodingResponseWriter) convert(b *builder2.Builder, context *types.APIOperation, input map[string]interface{}) *types.RawResource {
	schema := context.Schemas.Schema(definition.GetFullType(input))
	if schema == nil {
		return nil
	}
	op := builder2.List
	if context.Method == http.MethodPost {
		op = builder2.ListForCreate
	}
	data, err := b.Construct(schema, input, op)
	if err != nil {
		logrus.Errorf("Failed to construct object on output: %v", err)
		return nil
	}

	rawResource := &types.RawResource{
		ID:          toString(input["id"]),
		Type:        schema.ID,
		Schema:      schema,
		Links:       map[string]string{},
		Actions:     map[string]string{},
		Values:      data,
		ActionLinks: context.Request.Header.Get("X-API-Action-Links") != "",
	}

	j.addLinks(b, schema, context, input, rawResource)

	if schema.Formatter != nil {
		schema.Formatter(context, rawResource)
	}

	return rawResource
}

func (j *EncodingResponseWriter) addLinks(b *builder2.Builder, schema *types.Schema, context *types.APIOperation, input map[string]interface{}, rawResource *types.RawResource) {
	if rawResource.ID == "" {
		return
	}

	self := context.URLBuilder.ResourceLink(rawResource.Schema, rawResource.ID)
	rawResource.Links["self"] = self
	if context.AccessControl.CanUpdate(context, input, schema) == nil {
		rawResource.Links["update"] = self
	}
	if context.AccessControl.CanDelete(context, input, schema) == nil {
		rawResource.Links["remove"] = self
	}
}

func newCollection(apiOp *types.APIOperation) *types.GenericCollection {
	result := &types.GenericCollection{
		Collection: types.Collection{
			Type:         "collection",
			ResourceType: apiOp.Type,
			CreateTypes:  map[string]string{},
			Links: map[string]string{
				"self": apiOp.URLBuilder.Current(),
			},
			Actions: map[string]string{},
		},
		Data: []interface{}{},
	}

	if apiOp.Method == http.MethodGet {
		if apiOp.AccessControl.CanCreate(apiOp, apiOp.Schema) == nil {
			result.CreateTypes[apiOp.Schema.ID] = apiOp.URLBuilder.Collection(apiOp.Schema)
		}
	}

	opts := parse.QueryOptions(apiOp, apiOp.Schema)
	result.Sort = &opts.Sort
	result.Sort.Reverse = apiOp.URLBuilder.ReverseSort(result.Sort.Order)
	result.Sort.Links = map[string]string{}
	result.Pagination = opts.Pagination
	result.Filters = map[string][]types.Condition{}

	for _, cond := range opts.Conditions {
		filters := result.Filters[cond.Field]
		result.Filters[cond.Field] = append(filters, cond.ToCondition())
	}

	for name := range apiOp.Schema.CollectionFilters {
		if _, ok := result.Filters[name]; !ok {
			result.Filters[name] = nil
		}
	}

	for queryField := range apiOp.Schema.CollectionFilters {
		field, ok := apiOp.Schema.ResourceFields[queryField]
		if ok && (field.Type == "string" || field.Type == "enum") {
			result.Sort.Links[queryField] = apiOp.URLBuilder.Sort(queryField)
		}
	}

	if result.Pagination != nil && result.Pagination.Partial {
		if result.Pagination.Next != "" {
			result.Pagination.Next = apiOp.URLBuilder.Marker(result.Pagination.Next)
		}
		if result.Pagination.Previous != "" {
			result.Pagination.Previous = apiOp.URLBuilder.Marker(result.Pagination.Previous)
		}
		if result.Pagination.First != "" {
			result.Pagination.First = apiOp.URLBuilder.Marker(result.Pagination.First)
		}
		if result.Pagination.Last != "" {
			result.Pagination.Last = apiOp.URLBuilder.Marker(result.Pagination.Last)
		}
	}

	return result
}
