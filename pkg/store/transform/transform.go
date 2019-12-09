package transform

import (
	"fmt"

	"github.com/rancher/norman/v2/pkg/httperror"
	"github.com/rancher/norman/v2/pkg/types"
)

type TransformerFunc func(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject, opt *types.QueryOptions) (types.APIObject, error)

type ListTransformerFunc func(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject, opt *types.QueryOptions) (types.APIObject, error)

type StreamTransformerFunc func(apiOp *types.APIRequest, schema *types.Schema, data chan types.APIEvent, w types.WatchRequest) (chan types.APIEvent, error)

var _ types.Store = (*Store)(nil)

type Store struct {
	Store             types.Store
	Transformer       TransformerFunc
	ListTransformer   ListTransformerFunc
	StreamTransformer StreamTransformerFunc
}

func (s *Store) ByID(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	data, err := s.Store.ByID(apiOp, schema, id)
	if err != nil {
		return types.ToAPI(nil), err
	}
	if s.Transformer == nil {
		return data, nil
	}
	obj, err := s.Transformer(apiOp, schema, data, nil)
	if obj.IsNil() && err == nil {
		return obj, httperror.NewAPIError(httperror.NotFound, fmt.Sprintf("%s not found", id))
	}
	return obj, err
}

func (s *Store) Watch(apiOp *types.APIRequest, schema *types.Schema, w types.WatchRequest) (chan types.APIEvent, error) {
	c, err := s.Store.Watch(apiOp, schema, w)
	if err != nil {
		return nil, err
	}

	if s.StreamTransformer != nil {
		return s.StreamTransformer(apiOp, schema, c, w)
	}

	return types.APIChan(c, func(data types.APIObject) types.APIObject {
		result, err := s.Transformer(apiOp, schema, data, &types.QueryOptions{})
		if err != nil {
			return types.ToAPI(nil)
		}
		return result
	}), nil
}

func (s *Store) List(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) (types.APIObject, error) {
	data, err := s.Store.List(apiOp, schema, opt)
	if err != nil {
		return types.ToAPI(nil), err
	}

	if s.ListTransformer != nil {
		return s.ListTransformer(apiOp, schema, data, opt)
	}

	if s.Transformer == nil {
		return data, nil
	}

	var result []map[string]interface{}
	for _, item := range data.List() {
		item, err := s.Transformer(apiOp, schema, types.ToAPI(item), opt)
		if err != nil {
			return types.ToAPI(nil), err
		}
		if !item.IsNil() {
			result = append(result, item.Map())
		}
	}

	return types.ToAPI(result), nil
}

func (s *Store) Create(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject) (types.APIObject, error) {
	data, err := s.Store.Create(apiOp, schema, data)
	if err != nil {
		return types.ToAPI(nil), err
	}
	if s.Transformer == nil {
		return data, nil
	}
	return s.Transformer(apiOp, schema, data, nil)
}

func (s *Store) Update(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject, id string) (types.APIObject, error) {
	data, err := s.Store.Update(apiOp, schema, data, id)
	if err != nil {
		return types.ToAPI(nil), err
	}
	if s.Transformer == nil {
		return data, nil
	}
	return s.Transformer(apiOp, schema, data, nil)
}

func (s *Store) Delete(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	obj, err := s.Store.Delete(apiOp, schema, id)
	if err != nil || obj.IsNil() {
		return obj, err
	}
	return s.Transformer(apiOp, schema, obj, nil)
}
