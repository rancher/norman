package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
)

func CreateHandler(apiOp *types.APIOperation, next types.RequestHandler) (interface{}, error) {
	var err error

	if err := apiOp.AccessControl.CanCreate(apiOp, apiOp.Schema); err != nil {
		return nil, err
	}

	data, err := ParseAndValidateBody(apiOp, true)
	if err != nil {
		return nil, err
	}

	store := apiOp.Schema.Store
	if store == nil {
		return nil, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	data, err = store.Create(apiOp, apiOp.Schema, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
