package transform

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
)

type TransformerFunc func(apiContext *types.APIContext, data map[string]interface{}) (map[string]interface{}, error)

type ListTransformerFunc func(apiContext *types.APIContext, data []map[string]interface{}) ([]map[string]interface{}, error)

type StreamTransformerFunc func(apiContext *types.APIContext, data chan map[string]interface{}) (chan map[string]interface{}, error)

type Store struct {
	Store             types.Store
	Transformer       TransformerFunc
	ListTransformer   ListTransformerFunc
	StreamTransformer StreamTransformerFunc
}

func (t *Store) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := t.Store.ByID(apiContext, schema, id)
	if err != nil {
		return nil, err
	}
	if t.Transformer == nil {
		return data, nil
	}
	return t.Transformer(apiContext, data)
}

func (t *Store) Watch(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	c, err := t.Store.Watch(apiContext, schema, opt)
	if err != nil {
		return nil, err
	}

	if t.StreamTransformer != nil {
		return t.StreamTransformer(apiContext, c)
	}

	return convert.Chan(c, func(data map[string]interface{}) map[string]interface{} {
		item, err := t.Transformer(apiContext, data)
		if err != nil {
			return nil
		}
		return item
	}), nil
}

func (t *Store) List(apiContext *types.APIContext, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	data, err := t.Store.List(apiContext, schema, opt)
	if err != nil {
		return nil, err
	}

	if t.ListTransformer != nil {
		return t.ListTransformer(apiContext, data)
	}

	if t.Transformer == nil {
		return data, nil
	}

	var result []map[string]interface{}
	for _, item := range data {
		item, err := t.Transformer(apiContext, item)
		if err != nil {
			return nil, err
		}
		if item != nil {
			result = append(result, item)
		}
	}

	return result, nil
}

func (t *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	data, err := t.Store.Create(apiContext, schema, data)
	if err != nil {
		return nil, err
	}
	if t.Transformer == nil {
		return data, nil
	}
	return t.Transformer(apiContext, data)
}

func (t *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	data, err := t.Store.Update(apiContext, schema, data, id)
	if err != nil {
		return nil, err
	}
	if t.Transformer == nil {
		return data, nil
	}
	return t.Transformer(apiContext, data)
}

func (t *Store) Delete(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	return t.Store.Delete(apiContext, schema, id)
}
