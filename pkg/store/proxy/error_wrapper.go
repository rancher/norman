package proxy

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
)

type errorStore struct {
	types.Store
}

func (e *errorStore) ByID(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := e.Store.ByID(apiOp, schema, id)
	return data, translateError(err)
}

func (e *errorStore) List(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	data, err := e.Store.List(apiOp, schema, opt)
	return data, translateError(err)
}

func (e *errorStore) Create(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	data, err := e.Store.Create(apiOp, schema, data)
	return data, translateError(err)

}

func (e *errorStore) Update(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	data, err := e.Store.Update(apiOp, schema, data, id)
	return data, translateError(err)

}

func (e *errorStore) Delete(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := e.Store.Delete(apiOp, schema, id)
	return data, translateError(err)

}

func (e *errorStore) Watch(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	data, err := e.Store.Watch(apiOp, schema, opt)
	return data, translateError(err)
}

func translateError(err error) error {
	if apiError, ok := err.(errors.APIStatus); ok {
		status := apiError.Status()
		return httperror.NewAPIErrorLong(int(status.Code), string(status.Reason), status.Message)
	}
	return err
}
