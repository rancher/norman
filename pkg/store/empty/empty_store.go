package empty

import (
	"github.com/rancher/norman/pkg/types"
)

type Store struct {
}

func (e *Store) Delete(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (e *Store) ByID(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (e *Store) List(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (e *Store) Create(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (e *Store) Update(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject, id string) (types.APIObject, error) {
	return types.APIObject{}, nil
}

func (e *Store) Watch(apiOp *types.APIRequest, schema *types.Schema, wr types.WatchRequest) (chan types.APIEvent, error) {
	return nil, nil
}
