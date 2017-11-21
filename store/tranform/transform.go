package tranform

import "github.com/rancher/norman/types"

type TransformerFunc func(apiContext *types.APIContext, data map[string]interface{}) (map[string]interface{}, error)

type ListTransformerFunc func(apiContext *types.APIContext, data []map[string]interface{}) ([]map[string]interface{}, error)

type TransformingStore struct {
	Store           types.Store
	Transformer     TransformerFunc
	ListTransformer ListTransformerFunc
}

func (t *TransformingStore) ByID(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := t.Store.ByID(apiContext, schema, id)
	if err != nil {
		return nil, err
	}
	if t.Transformer == nil {
		return data, nil
	}
	return t.Transformer(apiContext, data)
}

func (t *TransformingStore) List(apiContext *types.APIContext, schema *types.Schema, opt types.QueryOptions) ([]map[string]interface{}, error) {
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
		result = append(result, item)
	}

	return result, nil
}

func (t *TransformingStore) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	data, err := t.Store.Create(apiContext, schema, data)
	if err != nil {
		return nil, err
	}
	if t.Transformer == nil {
		return data, nil
	}
	return t.Transformer(apiContext, data)
}

func (t *TransformingStore) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	data, err := t.Store.Update(apiContext, schema, data, id)
	if err != nil {
		return nil, err
	}
	if t.Transformer == nil {
		return data, nil
	}
	return t.Transformer(apiContext, data)
}

func (t *TransformingStore) Delete(apiContext *types.APIContext, schema *types.Schema, id string) error {
	return t.Store.Delete(apiContext, schema, id)
}
