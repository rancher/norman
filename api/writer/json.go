package writer

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rancher/norman/parse/builder"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
)

type JSONResponseWriter struct {
}

func (j *JSONResponseWriter) start(apiContext *types.APIContext, code int, obj interface{}) {
	apiContext.Response.Header().Set("content-type", "application/json")
	apiContext.Response.WriteHeader(code)
}

func (j *JSONResponseWriter) Write(apiContext *types.APIContext, code int, obj interface{}) {
	j.start(apiContext, code, obj)
	j.Body(apiContext, code, obj)
}

func (j *JSONResponseWriter) Body(apiContext *types.APIContext, code int, obj interface{}) {
	var output interface{}

	builder := builder.NewBuilder(apiContext)

	switch v := obj.(type) {
	case []interface{}:
		output = j.writeInterfaceSlice(builder, apiContext, v)
	case []map[string]interface{}:
		output = j.writeMapSlice(builder, apiContext, v)
	case map[string]interface{}:
		output = j.convert(builder, apiContext, v)
	case types.RawResource:
		output = v
	}

	if output != nil {
		json.NewEncoder(apiContext.Response).Encode(output)
	}
}
func (j *JSONResponseWriter) writeMapSlice(builder *builder.Builder, apiContext *types.APIContext, input []map[string]interface{}) *types.GenericCollection {
	collection := newCollection(apiContext)
	for _, value := range input {
		converted := j.convert(builder, apiContext, value)
		if converted != nil {
			collection.Data = append(collection.Data, converted)
		}
	}

	return collection
}

func (j *JSONResponseWriter) writeInterfaceSlice(builder *builder.Builder, apiContext *types.APIContext, input []interface{}) *types.GenericCollection {
	collection := newCollection(apiContext)
	for _, value := range input {
		switch v := value.(type) {
		case map[string]interface{}:
			converted := j.convert(builder, apiContext, v)
			if converted != nil {
				collection.Data = append(collection.Data, converted)
			}
		default:
			collection.Data = append(collection.Data, v)
		}
	}
	return collection
}

func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	return fmt.Sprint(val)
}

func (j *JSONResponseWriter) convert(b *builder.Builder, context *types.APIContext, input map[string]interface{}) *types.RawResource {
	schema := context.Schemas.Schema(context.Version, fmt.Sprint(input["type"]))
	if schema == nil {
		return nil
	}
	data, err := b.Construct(schema, input, builder.List)
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

func (j *JSONResponseWriter) addLinks(b *builder.Builder, schema *types.Schema, context *types.APIContext, input map[string]interface{}, rawResource *types.RawResource) {
	if rawResource.ID != "" {
		rawResource.Links["self"] = context.URLBuilder.ResourceLink(rawResource)
	}
}

func newCollection(apiContext *types.APIContext) *types.GenericCollection {
	result := &types.GenericCollection{
		Collection: types.Collection{
			Type:         "collection",
			ResourceType: apiContext.Type,
			CreateTypes:  map[string]string{},
			Links: map[string]string{
				"self": apiContext.URLBuilder.Current(),
			},
			Actions: map[string]string{},
		},
		Data: []interface{}{},
	}

	if apiContext.Method == http.MethodGet {
		if apiContext.AccessControl.CanCreate(apiContext.Schema) {
			result.CreateTypes[apiContext.Schema.ID] = apiContext.URLBuilder.Collection(apiContext.Schema)
		}
	}

	if apiContext.QueryOptions != nil {
		result.Sort = &apiContext.QueryOptions.Sort
		result.Sort.Reverse = apiContext.URLBuilder.ReverseSort(result.Sort.Order)
		result.Pagination = apiContext.QueryOptions.Pagination
		result.Filters = map[string][]types.Condition{}

		for _, cond := range apiContext.QueryOptions.Conditions {
			filters := result.Filters[cond.Field]
			result.Filters[cond.Field] = append(filters, cond.ToCondition())
		}
	}

	return result
}
