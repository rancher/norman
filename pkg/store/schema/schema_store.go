package schema

import (
	"encoding/json"
	"net/http"
	"strings"

	empty2 "github.com/rancher/norman/pkg/store/empty"
	"github.com/rancher/norman/pkg/types"
	definition2 "github.com/rancher/norman/pkg/types/definition"
	slice2 "github.com/rancher/norman/pkg/types/slice"

	"github.com/rancher/norman/pkg/httperror"
)

type Store struct {
	empty2.Store
}

func NewSchemaStore() types.Store {
	return &Store{}
}

func (s *Store) ByID(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	for _, schema := range apiOp.Schemas.Schemas() {
		if strings.EqualFold(schema.ID, id) {
			schemaData := map[string]interface{}{}

			data, err := json.Marshal(s.modifyForAccessControl(apiOp, *schema))
			if err != nil {
				return nil, err
			}

			return schemaData, json.Unmarshal(data, &schemaData)
		}
	}
	return nil, httperror.NewAPIError(httperror.NotFound, "no such schema")
}

func (s *Store) modifyForAccessControl(context *types.APIOperation, schema types.Schema) *types.Schema {
	var resourceMethods []string
	if slice2.ContainsString(schema.ResourceMethods, http.MethodPut) && schema.CanUpdate(context) == nil {
		resourceMethods = append(resourceMethods, http.MethodPut)
	}
	if slice2.ContainsString(schema.ResourceMethods, http.MethodDelete) && schema.CanDelete(context) == nil {
		resourceMethods = append(resourceMethods, http.MethodDelete)
	}
	if slice2.ContainsString(schema.ResourceMethods, http.MethodGet) && schema.CanGet(context) == nil {
		resourceMethods = append(resourceMethods, http.MethodGet)
	}

	var collectionMethods []string
	if slice2.ContainsString(schema.CollectionMethods, http.MethodPost) && schema.CanCreate(context) == nil {
		collectionMethods = append(collectionMethods, http.MethodPost)
	}
	if slice2.ContainsString(schema.CollectionMethods, http.MethodGet) && schema.CanList(context) == nil {
		collectionMethods = append(collectionMethods, http.MethodGet)
	}

	schema.ResourceMethods = resourceMethods
	schema.CollectionMethods = collectionMethods

	return &schema
}

func (s *Store) Watch(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	return nil, nil
}

func (s *Store) List(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	schemaMap := apiOp.Schemas.SchemasByID()
	schemas := make([]*types.Schema, 0, len(schemaMap))
	schemaData := make([]map[string]interface{}, 0, len(schemaMap))

	included := map[string]bool{}

	for _, schema := range schemaMap {
		if included[schema.ID] {
			continue
		}

		if schema.CanList(apiOp) == nil || schema.CanGet(apiOp) == nil {
			schemas = s.addSchema(apiOp, schema, schemaMap, schemas, included)
		}
	}

	data, err := json.Marshal(schemas)
	if err != nil {
		return nil, err
	}

	return schemaData, json.Unmarshal(data, &schemaData)
}

func (s *Store) addSchema(apiOp *types.APIOperation, schema *types.Schema, schemaMap map[string]*types.Schema, schemas []*types.Schema, included map[string]bool) []*types.Schema {
	included[schema.ID] = true
	schemas = s.traverseAndAdd(apiOp, schema, schemaMap, schemas, included)
	schemas = append(schemas, s.modifyForAccessControl(apiOp, *schema))
	return schemas
}

func (s *Store) traverseAndAdd(apiOp *types.APIOperation, schema *types.Schema, schemaMap map[string]*types.Schema, schemas []*types.Schema, included map[string]bool) []*types.Schema {
	for _, field := range schema.ResourceFields {
		t := ""
		subType := field.Type
		for subType != t {
			t = subType
			subType = definition2.SubType(t)
		}

		if refSchema, ok := schemaMap[t]; ok && !included[t] {
			schemas = s.addSchema(apiOp, refSchema, schemaMap, schemas, included)
		}
	}

	for _, action := range schema.ResourceActions {
		for _, t := range []string{action.Output, action.Input} {
			if t == "" {
				continue
			}

			if refSchema, ok := schemaMap[t]; ok && !included[t] {
				schemas = s.addSchema(apiOp, refSchema, schemaMap, schemas, included)
			}
		}
	}

	for _, action := range schema.CollectionActions {
		for _, t := range []string{action.Output, action.Input} {
			if t == "" {
				continue
			}

			if refSchema, ok := schemaMap[t]; ok && !included[t] {
				schemas = s.addSchema(apiOp, refSchema, schemaMap, schemas, included)
			}
		}
	}

	return schemas
}
