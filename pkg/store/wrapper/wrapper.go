package wrapper

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"
)

func Wrap(store types.Store) types.Store {
	return &StoreWrapper{
		store: store,
	}
}

type StoreWrapper struct {
	store types.Store
}

func (s *StoreWrapper) ByID(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := s.store.ByID(apiOp, schema, id)
	if err != nil {
		return nil, err
	}

	return apiOp.FilterObject(nil, schema, data), nil
}

func (s *StoreWrapper) List(apiOp *types.APIOperation, schema *types.Schema, opts *types.QueryOptions) ([]map[string]interface{}, error) {
	data, err := s.store.List(apiOp, schema, opts)
	if err != nil {
		return nil, err
	}

	return apiOp.FilterList(opts, schema, data), nil
}

func (s *StoreWrapper) Watch(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	c, err := s.store.Watch(apiOp, schema, opt)
	if err != nil || c == nil {
		return nil, err
	}

	return convert2.Chan(c, func(data map[string]interface{}) map[string]interface{} {
		return apiOp.FilterObject(nil, schema, data)
	}), nil
}

func (s *StoreWrapper) Create(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	data, err := s.store.Create(apiOp, schema, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *StoreWrapper) Update(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	err := validateGet(apiOp, schema, id)
	if err != nil {
		return nil, err
	}

	data, err = s.store.Update(apiOp, schema, data, id)
	if err != nil {
		return nil, err
	}

	return apiOp.FilterObject(nil, schema, data), nil
}

func (s *StoreWrapper) Delete(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	if err := validateGet(apiOp, schema, id); err != nil {
		return nil, err
	}

	return s.store.Delete(apiOp, schema, id)
}

func validateGet(apiOp *types.APIOperation, schema *types.Schema, id string) error {
	store := schema.Store
	if store == nil {
		return nil
	}

	existing, err := store.ByID(apiOp, schema, id)
	if err != nil {
		return err
	}

	if apiOp.Filter(nil, schema, existing) == nil {
		return httperror.NewAPIError(httperror.NotFound, "failed to find "+id)
	}

	return nil
}
