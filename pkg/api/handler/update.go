package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
)

func UpdateHandler(apiOp *types.APIOperation, next types.RequestHandler) (interface{}, error) {
	if err := apiOp.AccessControl.CanUpdate(apiOp, nil, apiOp.Schema); err != nil {
		return nil, err
	}

	data, err := ParseAndValidateBody(apiOp, false)
	if err != nil {
		return nil, err
	}

	store := apiOp.Schema.Store
	if store == nil {
		return nil, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	data, err = store.Update(apiOp, apiOp.Schema, data, apiOp.Name)
	if err != nil {
		return nil, err
	}

	return data, nil
}
