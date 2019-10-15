package transform

import (
	"fmt"

	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"

	"github.com/rancher/norman/pkg/httperror"
)

type TransformerFunc func(apiOp *types.APIRequest, schema *types.Schema, data map[string]interface{}, opt *types.QueryOptions) (map[string]interface{}, error)

type ListTransformerFunc func(apiOp *types.APIRequest, schema *types.Schema, data []map[string]interface{}, opt *types.QueryOptions) ([]map[string]interface{}, error)

type StreamTransformerFunc func(apiOp *types.APIRequest, schema *types.Schema, data chan map[string]interface{}, opt *types.QueryOptions) (chan map[string]interface{}, error)

type Store struct {
	Store             types.Store
	Transformer       TransformerFunc
	ListTransformer   ListTransformerFunc
	StreamTransformer StreamTransformerFunc
}

func (s *Store) ByID(apiOp *types.APIRequest, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := s.Store.ByID(apiOp, schema, id)
	if err != nil {
		return nil, err
	}
	if s.Transformer == nil {
		return data, nil
	}
	obj, err := s.Transformer(apiOp, schema, data, nil)
	if obj == nil && err == nil {
		return obj, httperror.NewAPIError(httperror.NotFound, fmt.Sprintf("%s not found", id))
	}
	return obj, err
}

func (s *Store) Watch(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	c, err := s.Store.Watch(apiOp, schema, opt)
	if err != nil {
		return nil, err
	}

	if s.StreamTransformer != nil {
		return s.StreamTransformer(apiOp, schema, c, opt)
	}

	return convert2.Chan(c, func(data map[string]interface{}) map[string]interface{} {
		item, err := s.Transformer(apiOp, schema, data, opt)
		if err != nil {
			return nil
		}
		return item
	}), nil
}

func (s *Store) List(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	data, err := s.Store.List(apiOp, schema, opt)
	if err != nil {
		return nil, err
	}

	if s.ListTransformer != nil {
		return s.ListTransformer(apiOp, schema, data, opt)
	}

	if s.Transformer == nil {
		return data, nil
	}

	var result []map[string]interface{}
	for _, item := range data {
		item, err := s.Transformer(apiOp, schema, item, opt)
		if err != nil {
			return nil, err
		}
		if item != nil {
			result = append(result, item)
		}
	}

	return result, nil
}

func (s *Store) Create(apiOp *types.APIRequest, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	data, err := s.Store.Create(apiOp, schema, data)
	if err != nil {
		return nil, err
	}
	if s.Transformer == nil {
		return data, nil
	}
	return s.Transformer(apiOp, schema, data, nil)
}

func (s *Store) Update(apiOp *types.APIRequest, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	data, err := s.Store.Update(apiOp, schema, data, id)
	if err != nil {
		return nil, err
	}
	if s.Transformer == nil {
		return data, nil
	}
	return s.Transformer(apiOp, schema, data, nil)
}

func (s *Store) Delete(apiOp *types.APIRequest, schema *types.Schema, id string) (map[string]interface{}, error) {
	obj, err := s.Store.Delete(apiOp, schema, id)
	if err != nil || obj == nil {
		return obj, err
	}
	return s.Transformer(apiOp, schema, obj, nil)
}
