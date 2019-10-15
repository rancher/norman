package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
)

func UpdateHandler(apiOp *types.APIRequest) (types.APIObject, error) {
	if err := apiOp.AccessControl.CanUpdate(apiOp, types.APIObject{}, apiOp.Schema); err != nil {
		return types.APIObject{}, err
	}

	data, err := ParseAndValidateBody(apiOp, false)
	if err != nil {
		return types.APIObject{}, err
	}

	store := apiOp.Schema.Store
	if store == nil {
		return types.APIObject{}, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	err = validateGet(apiOp, apiOp.Schema, apiOp.Name)
	if err != nil {
		return types.APIObject{}, err
	}

	data, err = store.Update(apiOp, apiOp.Schema, data, apiOp.Name)
	if err != nil {
		return types.APIObject{}, err
	}

	return apiOp.FilterObject(nil, apiOp.Schema, data), nil
}
