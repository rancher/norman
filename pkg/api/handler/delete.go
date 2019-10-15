package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
)

func DeleteHandler(request *types.APIRequest) (types.APIObject, error) {
	if err := request.AccessControl.CanDelete(request, types.APIObject{}, request.Schema); err != nil {
		return types.APIObject{}, err
	}

	store := request.Schema.Store
	if store == nil {
		return types.APIObject{}, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	if err := validateGet(request, request.Schema, request.Name); err != nil {
		return types.APIObject{}, err
	}

	return store.Delete(request, request.Schema, request.Name)
}
