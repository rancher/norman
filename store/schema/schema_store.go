package schema

import (
	"encoding/json"

	"strings"

	"github.com/rancher/norman/store/empty"
	"github.com/rancher/norman/types"
)

type Store struct {
	empty.Store
}

func NewSchemaStore() types.Store {
	return &Store{}
}

func (s *Store) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	for _, schema := range apiContext.Schemas.SchemasForVersion(*apiContext.Version) {
		if strings.EqualFold(schema.ID, id) {
			schemaData := map[string]interface{}{}

			data, err := json.Marshal(schema)
			if err != nil {
				return nil, err
			}

			return schemaData, json.Unmarshal(data, &schemaData)
		}
	}
	return nil, nil
}

func (s *Store) List(apiContext *types.APIContext, schema *types.Schema, opt types.QueryOptions) ([]map[string]interface{}, error) {
	schemaMap := apiContext.Schemas.SchemasForVersion(*apiContext.Version)
	schemas := make([]*types.Schema, 0, len(schemaMap))
	schemaData := make([]map[string]interface{}, 0, len(schemaMap))

	for _, schema := range schemaMap {
		schemas = append(schemas, schema)
	}

	data, err := json.Marshal(schemas)
	if err != nil {
		return nil, err
	}

	return schemaData, json.Unmarshal(data, &schemaData)
}
