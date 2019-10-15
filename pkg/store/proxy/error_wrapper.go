package proxy

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
)

type errorStore struct {
	types.Store
}

func (e *errorStore) ByID(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	data, err := e.Store.ByID(apiOp, schema, id)
	return data, translateError(err)
}

func (e *errorStore) List(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) (types.APIObject, error) {
	data, err := e.Store.List(apiOp, schema, opt)
	return data, translateError(err)
}

func (e *errorStore) Create(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject) (types.APIObject, error) {
	data, err := e.Store.Create(apiOp, schema, data)
	return data, translateError(err)

}

func (e *errorStore) Update(apiOp *types.APIRequest, schema *types.Schema, data types.APIObject, id string) (types.APIObject, error) {
	data, err := e.Store.Update(apiOp, schema, data, id)
	return data, translateError(err)

}

func (e *errorStore) Delete(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	data, err := e.Store.Delete(apiOp, schema, id)
	return data, translateError(err)

}

func (e *errorStore) Watch(apiOp *types.APIRequest, schema *types.Schema, wr types.WatchRequest) (chan types.APIEvent, error) {
	data, err := e.Store.Watch(apiOp, schema, wr)
	return data, translateError(err)
}

func translateError(err error) error {
	if apiError, ok := err.(errors.APIStatus); ok {
		status := apiError.Status()
		return httperror.NewAPIErrorLong(int(status.Code), string(status.Reason), status.Message)
	}
	return err
}
