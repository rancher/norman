package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
)

func CreateHandler(apiOp *types.APIRequest) (types.APIObject, error) {
	var err error

	if err := apiOp.AccessControl.CanCreate(apiOp, apiOp.Schema); err != nil {
		return types.APIObject{}, err
	}

	data, err := ParseAndValidateBody(apiOp, true)
	if err != nil {
		return types.APIObject{}, err
	}

	store := apiOp.Schema.Store
	if store == nil {
		return types.APIObject{}, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	data, err = store.Create(apiOp, apiOp.Schema, data)
	if err != nil {
		return types.APIObject{}, err
	}

	return data, nil
}
