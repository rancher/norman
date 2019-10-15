package schema

import (
	"encoding/json"
	"strings"

	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/store/empty"
	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/norman/pkg/types/definition"
)

type Store struct {
	empty.Store
}

func NewSchemaStore() types.Store {
	return &Store{}
}

func (s *Store) ByID(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	for _, schema := range apiOp.Schemas.Schemas() {
		if strings.EqualFold(schema.ID, id) {
			schemaData := map[string]interface{}{}

			data, err := json.Marshal(schema)
			if err != nil {
				return types.APIObject{}, err
			}

			return types.ToAPI(schemaData), json.Unmarshal(data, &schemaData)
		}
	}
	return types.APIObject{}, httperror.NewAPIError(httperror.NotFound, "no such schema")
}

func (s *Store) Watch(apiOp *types.APIRequest, schema *types.Schema, wr types.WatchRequest) (chan types.APIEvent, error) {
	return nil, nil
}

func (s *Store) List(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) (types.APIObject, error) {
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
		return types.APIObject{}, err
	}

	if err := json.Unmarshal(data, &schemaData); err != nil {
		return types.APIObject{}, err
	}
	return types.ToAPI(schemaData), nil
}

func (s *Store) addSchema(apiOp *types.APIRequest, schema *types.Schema, schemaMap map[string]*types.Schema, schemas []*types.Schema, included map[string]bool) []*types.Schema {
	included[schema.ID] = true
	schemas = s.traverseAndAdd(apiOp, schema, schemaMap, schemas, included)
	schemas = append(schemas, schema)
	return schemas
}

func (s *Store) traverseAndAdd(apiOp *types.APIRequest, schema *types.Schema, schemaMap map[string]*types.Schema, schemas []*types.Schema, included map[string]bool) []*types.Schema {
	for _, field := range schema.ResourceFields {
		t := ""
		subType := field.Type
		for subType != t {
			t = subType
			subType = definition.SubType(t)
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
