package empty

import (
	"github.com/rancher/norman/pkg/types"
)

type Store struct {
}

func (e *Store) Delete(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	return nil, nil
}

func (e *Store) ByID(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	return nil, nil
}

func (e *Store) List(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	return nil, nil
}

func (e *Store) Create(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (e *Store) Update(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	return nil, nil
}

func (e *Store) Watch(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	return nil, nil
}
