package handler

import (
	"github.com/rancher/norman/pkg/httperror"
	"github.com/rancher/norman/pkg/types"
)

func DeleteHandler(request *types.APIOperation, next types.RequestHandler) (interface{}, error) {
	if err := request.AccessControl.CanDelete(request, nil, request.Schema); err != nil {
		return nil, err
	}

	store := request.Schema.Store
	if store == nil {
		return nil, httperror.NewAPIError(httperror.NotFound, "no store found")
	}

	data, err := store.Delete(request, request.Schema, request.Name)
	if data == nil {
		// ensure nil, not just nil value interface
		return nil, err
	}
	return data, err
}
