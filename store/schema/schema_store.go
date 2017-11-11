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
	for _, schema := range apiContext.Schemas.Schemas() {
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

func (s *Store) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	schemaData := []map[string]interface{}{}

	data, err := json.Marshal(apiContext.Schemas.Schemas())
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &schemaData); err != nil {
		return nil, err
	}
	return schemaData, nil
}
