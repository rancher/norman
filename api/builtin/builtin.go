package builtin

import (
	"net/http"

	"github.com/rancher/norman/store/empty"
	"github.com/rancher/norman/store/schema"
	"github.com/rancher/norman/types"
)

var (
	Version = types.APIVersion{
		Group:   "io.cattle.builtin",
		Version: "v3",
		Path:    "/v3",
	}

	Schema = types.Schema{
		ID:                "schema",
		Version:           Version,
		CollectionMethods: []string{"GET"},
		ResourceMethods:   []string{"GET"},
		ResourceFields: map[string]types.Field{
			"collectionActions": {Type: "map[json]"},
			"collectionFields":  {Type: "map[json]"},
			"collectionFitlers": {Type: "map[json]"},
			"collectionMethods": {Type: "array[string]"},
			"pluralName":        {Type: "string"},
			"resourceActions":   {Type: "map[json]"},
			"resourceFields":    {Type: "map[json]"},
			"resourceMethods":   {Type: "array[string]"},
			"version":           {Type: "map[json]"},
		},
		Formatter: SchemaFormatter,
		Store:     schema.NewSchemaStore(),
	}

	Error = types.Schema{
		ID:                "error",
		Version:           Version,
		ResourceMethods:   []string{},
		CollectionMethods: []string{},
		ResourceFields: map[string]types.Field{
			"code":    {Type: "string"},
			"detail":  {Type: "string"},
			"message": {Type: "string"},
			"status":  {Type: "int"},
		},
	}

	Collection = types.Schema{
		ID:                "error",
		Version:           Version,
		ResourceMethods:   []string{},
		CollectionMethods: []string{},
		ResourceFields: map[string]types.Field{
			"data":       {Type: "array[json]"},
			"pagination": {Type: "map[json]"},
			"sort":       {Type: "map[json]"},
			"filters":    {Type: "map[json]"},
		},
	}

	APIRoot = types.Schema{
		ID:                "apiRoot",
		Version:           Version,
		ResourceMethods:   []string{},
		CollectionMethods: []string{},
		ResourceFields: map[string]types.Field{
			"apiVersion": {Type: "map[json]"},
			"path":       {Type: "string"},
		},
		Formatter: APIRootFormatter,
		Store:     NewAPIRootStore(nil),
	}

	Schemas = types.NewSchemas().
		AddSchema(&Schema).
		AddSchema(&Error).
		AddSchema(&Collection).
		AddSchema(&APIRoot)
)

func apiVersionFromMap(apiVersion map[string]interface{}) types.APIVersion {
	path, _ := apiVersion["path"].(string)
	version, _ := apiVersion["version"].(string)
	group, _ := apiVersion["group"].(string)

	return types.APIVersion{
		Path:    path,
		Version: version,
		Group:   group,
	}
}

func SchemaFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	data, _ := resource.Values["version"].(map[string]interface{})
	apiVersion := apiVersionFromMap(data)

	schema := apiContext.Schemas.Schema(&apiVersion, resource.ID)
	collectionLink := getSchemaCollectionLink(apiContext, schema)
	if collectionLink != "" {
		resource.Links["collection"] = collectionLink
	}
}

func getSchemaCollectionLink(apiContext *types.APIContext, schema *types.Schema) string {
	if schema != nil && contains(schema.CollectionMethods, http.MethodGet) {
		return apiContext.URLBuilder.Collection(schema)
	}
	return ""
}

func contains(list []string, needle string) bool {
	for _, v := range list {
		if v == needle {
			return true
		}
	}
	return false
}

func APIRootFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	path, _ := resource.Values["path"].(string)
	if path == "" {
		return
	}

	resource.Links["root"] = apiContext.URLBuilder.RelativeToRoot(path)

	data, _ := resource.Values["apiVersion"].(map[string]interface{})
	apiVersion := apiVersionFromMap(data)

	for name, schema := range apiContext.Schemas.SchemasForVersion(apiVersion) {
		collectionLink := getSchemaCollectionLink(apiContext, schema)
		if collectionLink != "" {
			resource.Links[name] = collectionLink
		}
	}
}

type APIRootStore struct {
	empty.Store
	roots []string
}

func NewAPIRootStore(roots []string) types.Store {
	return &APIRootStore{roots: roots}
}

func (a *APIRootStore) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	for _, version := range apiContext.Schemas.Versions() {
		if version.Path == id {
			return apiVersionToAPIRootMap(version), nil
		}
	}
	return nil, nil
}

func (a *APIRootStore) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	roots := []map[string]interface{}{}

	for _, version := range apiContext.Schemas.Versions() {
		roots = append(roots, apiVersionToAPIRootMap(version))
	}

	for _, root := range a.roots {
		roots = append(roots, map[string]interface{}{
			"path": root,
		})
	}

	return roots, nil
}

func apiVersionToAPIRootMap(version types.APIVersion) map[string]interface{} {
	return map[string]interface{}{
		"type": "/v3/apiRoot",
		"apiVersion": map[string]interface{}{
			"version": version.Version,
			"group":   version.Group,
			"path":    version.Path,
		},
		"path": version.Path,
	}
}
